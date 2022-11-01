package service

import (
	"context"
	"fmt"
	"github.com/ipfs-force-community/chain-co/dep"
	"time"

	"github.com/dtynn/dix"
	"go.uber.org/fx"

	"github.com/ipfs-force-community/chain-co/co"
	"github.com/ipfs-force-community/chain-co/proxy"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
)

const extractFullNodeAPIKey dix.Invoke = 1

// Build constructs the app with given di options
func Build(ctx context.Context, overrides ...dix.Option) (dix.StopFunc, error) {
	opts := []dix.Option{
		dix.Override(new(co.NodeOption), co.DefaultNodeOption),
		dix.Override(new(*co.Ctx), co.NewCtx),
		dix.Override(new(*co.Connector), co.NewConnector),
		dix.Override(new(*co.Coordinator), buildCoordinator),
		dix.Override(new(*co.Selector), co.NewSelector),
		dix.Override(new(*proxy.Proxy), buildProxyAPI),
		dix.Override(new(*proxy.Local), buildLocalAPI),
		dix.Override(new(*proxy.UnSupport), buildUnSupportAPI),
	}
	opts = append(opts, overrides...)
	return dix.New(ctx, opts...)
}

// FullNode extracts api.FullNode from inside di
func FullNode(full *api.FullNode) dix.Option {
	return dix.Override(extractFullNodeAPIKey, func(srv Service) error {
		*full = &srv
		return nil
	})
}

// ParseNodeInfoList is provided to the higer-lvel
func ParseNodeInfoList(raws []string, version string) dix.Option {
	return dix.Override(new(co.NodeInfoList), func() (co.NodeInfoList, error) {
		list := make(co.NodeInfoList, 0, len(raws))
		for _, str := range raws {
			info := co.ParseNodeInfo(str)
			if _, err := info.DialArgs(version); err != nil {
				return nil, fmt.Errorf("invalid node info: %s", str)
			}

			list = append(list, info)
		}

		return list, nil
	})
}

func buildCoordinator(lc fx.Lifecycle, ctx *co.Ctx, connector *co.Connector, infos co.NodeInfoList, version dep.APIVersion, sel *co.Selector) (*co.Coordinator, error) {
	nodes := make([]*co.Node, 0, len(infos))
	allDone := false
	defer func() {
		if !allDone {
			for i := range nodes {
				nodes[i].Stop() // nolint:errcheck
			}
		}
	}()

	var head *types.TipSet
	weight := types.NewInt(0)

	for i := range infos {
		info := infos[i]
		nlog := log.With("host", info.Addr)

		node, err := connector.Connect(info, string(version))
		if err != nil {
			nlog.Errorf("connect failed: %s", err)
			continue
		}

		full := node.FullNode()
		h, w, err := getHeadCandidate(full)
		if err != nil {
			node.Stop() // nolint:errcheck
			nlog.Errorf("failed to get head: %s", err)
			continue
		}

		if head == nil || w.GreaterThan(weight) {
			head = h
			weight = w
		}

		nlog.Infof("add new node %s", info.Addr)
		nodes = append(nodes, node)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available node")
	}

	coordinator, err := co.NewCoordinator(ctx, head, weight, sel)
	if err != nil {
		return nil, err
	}

	nodeProvider := co.NewNodeProvider()
	nodeProvider.AddNodes(nodes)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go coordinator.Start()
			sel.SetNodeProvider(nodeProvider)
			return nil
		},
		OnStop: func(context.Context) error {
			coordinator.Stop() // nolint:errcheck
			return nil
		},
	})

	allDone = true
	return coordinator, nil
}

func getHeadCandidate(full api.FullNode) (*types.TipSet, types.BigInt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	head, err := full.ChainHead(ctx)
	if err != nil {
		return nil, types.BigInt{}, err
	}

	weight, err := full.ChainTipSetWeight(ctx, head.Key())
	if err != nil {
		return nil, types.BigInt{}, err
	}

	return head, weight, nil
}

func buildProxyAPI(sel *co.Selector) *proxy.Proxy {
	return &proxy.Proxy{
		Select: func() (proxy.ProxyAPI, error) {
			node, err := sel.Select()
			if err != nil {
				return nil, err
			}

			return node.FullNode(), nil
		},
	}
}

func buildLocalAPI(lsrv LocalChainService) *proxy.Local {
	return &proxy.Local{
		Select: func() (proxy.LocalAPI, error) {
			return &lsrv, nil
		},
	}
}

func buildUnSupportAPI() *proxy.UnSupport {
	return &proxy.UnSupport{
		Select: func() (proxy.UnSupportAPI, error) {
			return nil, fmt.Errorf("api not supported")
		},
	}
}
