package dep

import (
	"context"

	"github.com/dtynn/dix"
	metricsi "github.com/ipfs/go-metrics-interface"

	"github.com/filecoin-project/lotus/node/modules/helpers"
)

type APIVersion string

func APIVersionOption(version string) dix.Option {
	return dix.Override(new(APIVersion), APIVersion(version))
}

// MetricsCtxOption retuns a Option to provide metric context
func MetricsCtxOption(ctx context.Context, scope string) dix.Option {
	return dix.Override(new(helpers.MetricsCtx), func() context.Context {
		return metricsi.CtxScope(ctx, scope)
	})
}
