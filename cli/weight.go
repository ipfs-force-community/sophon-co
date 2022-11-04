package cli

import (
	"fmt"
	"strconv"
	"text/tabwriter"

	"github.com/urfave/cli/v2"
)

var WeightCmd = &cli.Command{
	Name:  "weight",
	Usage: "manipulate the weight of node",
	Subcommands: []*cli.Command{
		weightListCmd,
		weightSetCmd,
	},
}

var weightListCmd = &cli.Command{
	Name:  "list",
	Usage: "list the weight and priority of node",
	Flags: []cli.Flag{},
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		addr := cctx.String("listen")
		client, closer, err := NewLocalRPCClient(ctx, addr)
		if err != nil {
			return err
		}
		defer closer()

		weight, err := client.ListWeight(ctx)
		if err != nil {
			return err
		}

		priority, err := client.ListPriority(ctx)
		if err != nil {
			return err
		}
		tw := tabwriter.NewWriter(cctx.App.Writer, 2, 4, 2, ' ', 0)
		fmt.Fprintln(tw, "Address\tWeight\tPriority")
		for addr, w := range weight {
			fmt.Fprintf(tw, "%s\t%d\t%d", addr, w, priority[addr])
			fmt.Fprintln(tw)
		}
		return tw.Flush()
	},
}

var weightSetCmd = &cli.Command{
	Name:      "set",
	Usage:     "set the weight of node",
	ArgsUsage: "[addr] [weight]",
	Flags:     []cli.Flag{},
	Action: func(cctx *cli.Context) error {
		// check args
		if cctx.NArg() != 2 {
			return fmt.Errorf("must specify a node address and weight")
		}
		node := cctx.Args().Get(0)
		w := cctx.Args().Get(1)
		weight, err := strconv.Atoi(w)
		if err != nil {
			return err
		}

		ctx := cctx.Context
		listen := cctx.String("listen")
		client, closer, err := NewLocalRPCClient(ctx, listen)
		if err != nil {
			return err
		}
		defer closer()

		err = client.SetWeight(ctx, node, weight)
		if err != nil {
			return err
		}
		return nil
	},
}
