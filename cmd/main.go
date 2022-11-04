package main

import (
	"context"
	"os"

	lcli "github.com/ipfs-force-community/chain-co/cli"
	"github.com/ipfs-force-community/chain-co/version"

	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
	"go.opencensus.io/trace"

	"github.com/filecoin-project/lotus/lib/lotuslog"
	"github.com/filecoin-project/lotus/lib/tracing"
)

const cliName = "chain-ro"

var log = logging.Logger(cliName)

func main() {
	lotuslog.SetupLogLevels()

	local := []*cli.Command{
		runCmd,
		lcli.WeightCmd,
	}

	jaeger := tracing.SetupJaegerTracing(cliName)
	defer func() {
		if jaeger != nil {
			_ = jaeger.ForceFlush(context.Background())
		}
	}()

	for _, cmd := range local {
		cmd := cmd
		originBefore := cmd.Before
		cmd.Before = func(cctx *cli.Context) error {
			if jaeger != nil {
				_ = jaeger.Shutdown(cctx.Context)
			}
			jaeger = tracing.SetupJaegerTracing(cliName + "/" + cmd.Name)

			if originBefore != nil {
				return originBefore(cctx)
			}
			return nil
		}
	}

	ctx, span := trace.StartSpan(context.Background(), "/cli")
	defer span.End()

	app := &cli.App{
		Name:                 cliName,
		Usage:                "read-only chain node for filecoin",
		EnableBashCompletion: true,
		Version:              version.Version + version.CurrentCommit,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "listen",
				Usage: "listen address for the service",
				Value: "0.0.0.0:1234",
			},
		},

		Commands: local,
	}

	app.Setup()
	app.Metadata["traceContext"] = ctx

	if err := app.Run(os.Args); err != nil {
		log.Errorf("CLI error: %s", err)
		os.Exit(1)
	}
}
