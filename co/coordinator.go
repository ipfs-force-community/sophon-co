package co

import (
	"fmt"
	"sync"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/whyrusleeping/pubsub"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
)

// common errors
var (
	ErrNoNodeAvailable = fmt.Errorf("no node available")
)

var log = logging.Logger("chain-co")

const (
	tipsetChangeTopic = "tschange"
)

// NewCoordinator constructs a Coordinator instance
func NewCoordinator(ctx *Ctx, head *types.TipSet, weight types.BigInt, sel *Selector) (*Coordinator, error) {
	return &Coordinator{
		ctx:    ctx,
		head:   head,
		weight: weight,
		nodes:  make([]string, 0, 16),
		sel:    sel,
		tspub:  pubsub.New(256),
	}, nil
}

// Coordinator tries to setup the best nodes based on their incoming chain head
type Coordinator struct {
	ctx *Ctx

	headMu sync.RWMutex
	head   *types.TipSet
	weight types.BigInt
	nodes  []string

	sel *Selector

	tspub *pubsub.PubSub
}

// Start starts the coordinate loop
func (c *Coordinator) Start() {
	log.Info("start head coordinator loop")
	defer log.Info("stop head coordinator loop")

	for {
		select {
		case <-c.ctx.lc.Done():
			return
		case hc := <-c.ctx.headCh:
			c.handleCandidate(hc)
		case addr := <-c.ctx.errNodeCh:
			c.delNodeAddr(addr)
		}
	}
}

// Stop shuts down the included components
func (c *Coordinator) Stop() error {
	c.tspub.Shutdown()
	return nil
}
func (c *Coordinator) delNodeAddr(addr string) {
	c.sel.setPriority(ErrPriority, addr)
}
func (c *Coordinator) handleCandidate(hc *headCandidate) {
	addr := hc.node.info.Addr
	clog := log.With("node", addr, "h", hc.ts.Height(), "w", hc.weight, "drift", time.Now().Unix()-int64(hc.ts.MinTimestamp()))

	c.headMu.Lock()
	defer c.headMu.Unlock()
	//1. more weight
	//2. if equal weight. select more blocks
	if c.head == nil || hc.weight.GreaterThan(c.weight) || (hc.weight.Equals(c.weight) && len(hc.ts.Blocks()) > len(c.head.Blocks())) {
		clog.Info("head replaced")

		prev := c.head
		next := hc.ts
		headChanges, err := c.applyTipSetChange(prev, next, hc.node) //todo if network become slow
		if err != nil {
			clog.Errorf("apply tipset change: %s", err)
		}
		if headChanges == nil {
			return
		}

		c.head = hc.ts
		c.weight = hc.weight
		c.nodes = append(c.nodes[:0], addr)

		preAddrs := c.sel.getAddrOfPriority(CatchUpPriority)
		c.sel.setPriority(DelayPriority, preAddrs...)
		c.sel.setPriority(CatchUpPriority, addr)

		c.tspub.Pub(headChanges, tipsetChangeTopic)

		return
	}

	if c.head.Equals(hc.ts) {
		contains := false
		for ni := range c.nodes {
			if c.nodes[ni] == addr {
				contains = true
				break
			}
		}

		if !contains {
			c.nodes = append(c.nodes, addr)
			c.sel.setPriority(CatchUpPriority, addr)
			clog.Infof("another node %s caught up", addr)
		}
		return
	}

	clog.Debug("ignored a lighter head")
}

func (c *Coordinator) applyTipSetChange(prev, next *types.TipSet, node *Node) ([]*api.HeadChange, error) {
	revert, apply, err := store.ReorgOps(c.ctx.lc, node.loadTipSet, prev, next)
	if err != nil {
		return nil, err
	}

	hc := make([]*api.HeadChange, 0, len(revert)+len(apply))
	for i := range revert {
		hc = append(hc, &api.HeadChange{
			Type: store.HCRevert,
			Val:  revert[i],
		})
	}

	for i := range apply {
		hc = append(hc, &api.HeadChange{
			Type: store.HCApply,
			Val:  apply[i],
		})
	}

	if len(hc) == 0 {
		return nil, nil
	}
	return hc, nil
}
