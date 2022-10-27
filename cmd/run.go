package main

import (
	"context"
	"io/ioutil"
	"strings"

	"github.com/ipfs-force-community/metrics"

	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/lotus/api/v1api"

	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/ipfs-force-community/chain-co/dep"
	"github.com/ipfs-force-community/chain-co/service"
)

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "start the chain-co daemon",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "listen",
			Usage: "listen address for the service",
			Value: ":1234",
		},
		&cli.Int64Flag{
			Name:  "max-req-size",
			Usage: "max request size",
			Value: 10 << 20,
		},
		&cli.StringSliceFlag{
			Name:  "node",
			Usage: "node info",
		},
		&cli.StringFlag{
			Name:  "auth-url",
			Usage: "specify url for connect to venus-auth",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "rate-limit-redis",
			Usage: "config redis to request api limit",
		},
		&cli.StringFlag{
			Name:        "version",
			Usage:       "rpc api version",
			Value:       "v1",
			DefaultText: "v1",
		},
		&cli.StringFlag{
			Name: "jaeger-proxy",
		},
		&cli.Float64Flag{
			Name:  "trace-sampler",
			Value: 1.0,
		},
		&cli.StringFlag{
			Name:  "trace-node-name",
			Value: "venus-node-co",
		},
	},
	Action: func(cctx *cli.Context) error {
		appCtx, appCancel := context.WithCancel(cctx.Context)
		defer appCancel()

		var full v1api.FullNode

		localJwt, token, err := jwtclient.NewLocalAuthClient()
		if err != nil {
			return err
		}
		err = ioutil.WriteFile("./token", token, 0o666)

		if err != nil {
			return err
		}
		stop, err := service.Build(
			appCtx,

			dep.MetricsCtxOption(appCtx, cliName),

			dep.APIVersionOption(cctx.String("version")),
			service.ParseNodeInfoList(cctx.StringSlice("node"), cctx.String("version")),
			service.FullNode(&full),
		)
		if err != nil {
			return err
		}

		defer stop(context.Background()) // nolint:errcheck

		mCnf := &metrics.TraceConfig{}
		proxy, sampler, serverName := strings.TrimSpace(cctx.String("jaeger-proxy")),
			cctx.Float64("trace-sampler"),
			strings.TrimSpace(cctx.String("trace-node-name"))

		if mCnf.JaegerTracingEnabled = len(proxy) != 0; mCnf.JaegerTracingEnabled {
			mCnf.ProbabilitySampler, mCnf.JaegerEndpoint, mCnf.ServerName = sampler, proxy, serverName
		}

		return serveRPC(
			appCtx,
			cctx.String("auth-url"),
			cctx.String("rate-limit-redis"),
			cctx.String("listen"),
			mCnf,
			localJwt,
			full,
			func(ctx context.Context) error {
				appCancel()
				stop(ctx) // nolint:errcheck
				return nil
			},
			cctx.Int64("max-req-size"),
		)
	},
}
