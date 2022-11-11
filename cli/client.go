package cli

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/filecoin-project/go-jsonrpc"
	local_api "github.com/ipfs-force-community/chain-co/cli/api"
)

func NewLocalRPCClient(ctx context.Context, addr string, opts ...jsonrpc.Option) (local_api.LocalAPI, jsonrpc.ClientCloser, error) {
	port := strings.Split(addr, ":")[1]
	endpoint := fmt.Sprintf("http://127.0.0.1:%s/rpc/admin/v0", port)

	token, err := os.ReadFile("./token")
	token = bytes.TrimSpace(token)
	if err != nil {
		return nil, nil, err
	}

	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+string(token))

	var res local_api.LocalAPIStruct
	closer, err := jsonrpc.NewMergeClient(ctx, endpoint, "Filecoin",
		[]interface{}{
			&res,
		},
		headers,
		opts...,
	)

	return &res, closer, err
}
