package service

import (
	"context"

	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"

	local_api "github.com/ipfs-force-community/chain-co/cli/api"
	"github.com/ipfs-force-community/chain-co/co"
	"github.com/ipfs-force-community/chain-co/proxy"
)

var log = logging.Logger("chain-ro-srv")

// LocalChainService impls proxy.Local
type LocalChainService struct {
	fx.In
	*co.Coordinator
}

// Service impls api.FullNode
type Service struct {
	fx.In

	*proxy.Proxy
	*proxy.Local
	*proxy.UnSupport
}

// LocalChainService impls cli/api.LocalAPI
type LocalAPIService struct {
	fx.In
	*co.Selector
}

var _ local_api.LocalAPI = (*LocalAPIService)(nil)

func (l *LocalAPIService) SetWeight(ctx context.Context, addr string, weight int) error {
	return l.Selector.SetWeight(addr, weight)
}

func (l *LocalAPIService) ListWeight(ctx context.Context) (map[string]int, error) {
	return l.Selector.ListWeight(), nil
}

func (l *LocalAPIService) ListPriority(ctx context.Context) (map[string]int, error) {
	return l.Selector.ListPriority(), nil
}
