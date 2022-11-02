package co

import (
	"context"
	"sync"
	"time"

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

// NodeInfo is a type alias for cliutil.APIInfo
type NodeInfo = apiinfo.APIInfo

// ParseNodeInfo is an alias to the cliutil.ParseApiInfo function
var ParseNodeInfo = apiinfo.ParseApiInfo

// NewConnector constructs a Connector instance
func NewConnector(ctx *Ctx) (*Connector, error) {
	return &Connector{
		Ctx: ctx,
	}, nil
}

// Connector is a helper for connecting upstream nodes
type Connector struct {
	*Ctx
}

// Connect connects to the specified node with given info
func (c *Connector) Connect(info NodeInfo, version string) (*Node, error) {
	addr, err := info.DialArgs(version)
	if err != nil {
		return nil, err
	}

	full, closer, err := client.NewFullNodeRPCV1(c.Ctx.lc, addr, info.AuthHeader())
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(c.Ctx.lc)
	node := &Node{
		opt:    c.Ctx.nodeOpt,
		info:   info,
		ctx:    ctx,
		cancel: cancel,
		sctx:   c.Ctx,
		log:    log.With("remote", addr),
	}

	node.upstream.full = full
	node.upstream.closer = closer

	return node, nil
}

// Node is a FullNode client
type Node struct {
	opt  NodeOption
	info NodeInfo

	reListenInterval time.Duration

	ctx    context.Context
	cancel context.CancelFunc

	sctx *Ctx

	upstream struct {
		full   v1api.FullNode
		closer jsonrpc.ClientCloser
	}

	log *zap.SugaredLogger
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
	n.upstream.closer()
	return nil
}

// FullNode returns the client to the upstream node
func (n *Node) FullNode() v1api.FullNode {
	return n.upstream.full
}

func (n *Node) Host() (string, error) {
	return n.info.Host()
}

func (n *Node) reListen() (<-chan []*api.HeadChange, error) {
	for {
		ch, err := n.upstream.full.ChainNotify(n.ctx)
		if err != nil {
			n.log.Errorf("call CahinNotify: %s, will re-call in %s", err, n.reListenInterval)
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
		case store.HCCurrent, store.HCApply:
			idx = i
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

const (
	ADD    = true
	REMOVE = false
)

type INodeProvider interface {
	GetNode(host string) *Node
	GetHosts() []string
	AddHook(func(alter map[string]bool))
}

var _ INodeProvider = (*NodeProvider)(nil)

type NodeProvider struct {
	nodes map[string]*Node
	lk    sync.RWMutex
	hooks []func(alter map[string]bool)
}

func NewNodeProvider() *NodeProvider {
	return &NodeProvider{
		nodes: make(map[string]*Node),
	}
}

func (p *NodeProvider) GetNode(host string) *Node {
	p.lk.RLock()
	defer p.lk.RUnlock()
	return p.nodes[host]
}

func (p *NodeProvider) GetHosts() []string {
	p.lk.RLock()
	defer p.lk.RUnlock()
	hosts := make([]string, 0, len(p.nodes))
	for host := range p.nodes {
		hosts = append(hosts, host)
	}
	return hosts
}

func (p *NodeProvider) AddHook(hook func(alter map[string]bool)) {
	p.lk.Lock()
	defer p.lk.Unlock()
	p.hooks = append(p.hooks, hook)
}

func (p *NodeProvider) AddNodes(add []*Node) {
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

	for _, hook := range p.hooks {
		hook(alt)
	}
}
