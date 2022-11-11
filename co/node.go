package co

import (
	"context"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"

	"github.com/filecoin-project/lotus/api/v1api"
	"github.com/ipfs-force-community/venus-common-utils/apiinfo"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/hashicorp/go-multierror"

	"github.com/ipfs/go-cid"
	"go.uber.org/zap"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
)

// NodeInfoList is a type def for dependency injection
type NodeInfoList []NodeInfo

// DefaultNodeOption returns default options
func DefaultNodeOption() NodeOption {
	return NodeOption{
		ReListenMinInterval: 4 * time.Second,
		ReListenMaxInterval: 32 * time.Second,
		APITimeout:          10 * time.Second,
	}
}

// NodeOption is for node configuration
type NodeOption struct {
	ReListenMinInterval time.Duration
	ReListenMaxInterval time.Duration

	APITimeout time.Duration
}

// NodeInfo is a type combine cliutil.APIInfo and protocol version
type NodeInfo struct {
	apiinfo.APIInfo
	Version string
}

func NewNodeInfo(addr string, version string) NodeInfo {
	return NodeInfo{
		APIInfo: apiinfo.ParseApiInfo(addr),
		Version: version,
	}
}

// Node is a FullNode client
type Node struct {
	opt  NodeOption
	info NodeInfo
	Addr string

	reListenInterval time.Duration

	ctx    context.Context
	cancel context.CancelFunc

	sctx *Ctx

	upstream struct {
		full   v1api.FullNode
		closer jsonrpc.ClientCloser
	}

	blkCache *lru.ARCCache
	log      *zap.SugaredLogger
}

func NewNode(cctx *Ctx, info NodeInfo) (*Node, error) {
	addr, err := info.DialArgs(info.Version)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(cctx.lc)
	blkCache, err := lru.NewARC(100)
	if err != nil {
		cancel()
		return nil, err
	}
	return &Node{
		reListenInterval: cctx.nodeOpt.ReListenMinInterval,
		opt:              cctx.nodeOpt,
		info:             info,
		ctx:              ctx,
		cancel:           cancel,
		sctx:             cctx,
		Addr:             info.Addr,
		blkCache:         blkCache,
		log:              log.With("remote", addr),
	}, nil
}

func (n *Node) Connect() error {
	info := n.info
	addr, err := info.DialArgs(info.Version)
	if err != nil {
		return err
	}

	full, closer, err := client.NewFullNodeRPCV1(n.ctx, addr, info.AuthHeader())
	if err != nil {
		return err
	}

	n.upstream.full = full
	n.upstream.closer = closer
	return nil
}

// Start starts a head change loop
func (n *Node) Start() {
	n.log.Info("start head change loop")
	defer n.log.Info("stop head change loop")

	for {
		ch, err := n.reListen()
		if err != nil {
			if err != context.Canceled && err != context.DeadlineExceeded {
				n.log.Errorf("failed to listen head change: %s", err)
			}
			return
		}

		chLifeCtx, chLifeCancel := context.WithCancel(n.ctx)

	CHANGES_LOOP:
		for {
			select {
			case <-n.ctx.Done():
				chLifeCancel()
				return

			case changes, ok := <-ch:
				if !ok {
					break CHANGES_LOOP
				}

				go n.applyChanges(chLifeCtx, changes)
			}
		}

		chLifeCancel()
	}
}

// Stop closes current node
func (n *Node) Stop() error {
	n.cancel()
	if n.upstream.closer != nil {
		n.upstream.closer()
	}
	return nil
}

// FullNode returns the client to the upstream node
func (n *Node) FullNode() v1api.FullNode {
	return n.upstream.full
}

func (n *Node) reListen() (<-chan []*api.HeadChange, error) {
	for {
		var err error
		var ch <-chan []*api.HeadChange
		// if full node client is nil,try reconnect
		if n.upstream.full == nil {
			err = n.Connect()
		}
		if err == nil {
			ch, err = n.upstream.full.ChainNotify(n.ctx)
			if err != nil {
				n.log.Errorf("call CahinNotify fail: %s", err)
			}
		} else {
			n.log.Errorf("failed to connect to upstream node: %s", err)
		}

		if err != nil {
			n.log.Infof("retry after %s", n.reListenInterval)
			n.sctx.errNodeCh <- n.info.Addr

			select {
			case <-n.ctx.Done():
				return nil, n.ctx.Err()

			case <-time.After(n.reListenInterval):

				n.reListenInterval *= 2
				if n.reListenInterval > n.opt.ReListenMaxInterval {
					n.reListenInterval = n.opt.ReListenMaxInterval
				}

			}

			continue
		}

		n.reListenInterval = n.opt.ReListenMinInterval
		return ch, nil
	}
}

func (n *Node) applyChanges(lifeCtx context.Context, changes []*api.HeadChange) {
	n.sctx.bcache.add(changes)

	idx := -1
	for i := range changes {
		switch changes[i].Type {
		case store.HCCurrent:
			idx = i
		case store.HCApply:
			idx = i
			n.storeKey(changes[i].Val.Key())
		}
	}

	if idx == -1 {
		return
	}

	ts := changes[idx].Val

	callCtx, callCancel := context.WithTimeout(lifeCtx, n.opt.APITimeout)
	weight, err := n.upstream.full.ChainTipSetWeight(callCtx, ts.Key())
	callCancel()

	if err != nil {
		n.log.Errorf("call ChainTipSetWeight: %s", err)
		return
	}

	hc := &headCandidate{
		node:   n,
		ts:     ts,
		weight: weight,
	}

	slow := time.NewTicker(5 * time.Second)
	defer slow.Stop()

	t := time.Now()

	for {
		select {
		case <-lifeCtx.Done():
			return

		case n.sctx.headCh <- hc:
			return

		case tick := <-slow.C:
			n.log.Warnf("it took too long before we can send the new head change, ts=%s, height=%d, weight=%s, delay=%s", ts.Key(), ts.Height(), weight, tick.Sub(t))
		}
	}
}

func (n *Node) loadTipSet(ctx context.Context, tsk types.TipSetKey) (*types.TipSet, error) {
	reqCtx, reqCancel := context.WithTimeout(ctx, n.opt.APITimeout)
	defer reqCancel()

	var wg multierror.Group
	cids := tsk.Cids()
	blks := make([]*types.BlockHeader, len(cids))
	for ci := range cids {
		i := ci
		wg.Go(func() error {
			blk, err := n.loadBlockHeader(reqCtx, cids[i])
			if err != nil {
				return err
			}

			blks[i] = blk
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return types.NewTipSet(blks)
}

func (n *Node) loadBlockHeader(ctx context.Context, c cid.Cid) (*types.BlockHeader, error) {
	if blk, ok := n.sctx.bcache.load(c); ok {
		return blk, nil
	}

	blk, err := n.upstream.full.ChainGetBlock(ctx, c)
	return blk, err
}

func (n *Node) hasKey(key types.TipSetKey) bool {
	for _, blkCid := range key.Cids() {
		_, has := n.blkCache.Peek(blkCid.String())
		if !has {
			return false
		}
	}
	return true
}

func (n *Node) storeKey(key types.TipSetKey) {
	for _, blkCid := range key.Cids() {
		_, has := n.blkCache.Peek(blkCid.String())
		if !has {
			n.blkCache.Add(blkCid.String(), nil)
		}
	}
}

const (
	ADD    = true
	REMOVE = false
)

//go:generate mockgen -destination=./node_store_mock.go -package=co github.com/ipfs-force-community/chain-co/co INodeStore
type INodeStore interface {
	GetNode(host string) *Node
	GetHosts() []string
	AddNodes([]*Node)
}

var _ INodeStore = (*NodeStore)(nil)

type NodeStore struct {
	nodes map[string]*Node
	lk    sync.RWMutex
}

func NewNodeStore() *NodeStore {
	return &NodeStore{
		nodes: make(map[string]*Node),
	}
}

func (p *NodeStore) GetNode(host string) *Node {
	p.lk.RLock()
	defer p.lk.RUnlock()
	return p.nodes[host]
}

func (p *NodeStore) GetHosts() []string {
	p.lk.RLock()
	defer p.lk.RUnlock()
	hosts := make([]string, 0, len(p.nodes))
	for host := range p.nodes {
		hosts = append(hosts, host)
	}
	return hosts
}

func (p *NodeStore) AddNodes(add []*Node) {
	p.lk.Lock()
	defer p.lk.Unlock()
	alt := make(map[string]bool)
	for _, node := range add {
		if _, exist := p.nodes[node.info.Addr]; !exist {
			p.nodes[node.info.Addr] = node
		} else {
			pre := p.nodes[node.info.Addr]
			pre.Stop() // nolint:errcheck
			p.nodes[node.info.Addr] = node

			alt[node.info.Addr] = ADD
		}
		go node.Start()
	}
}
