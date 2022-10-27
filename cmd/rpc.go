package main

import (
	"context"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/dtynn/dix"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v0api"
	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/ipfs-force-community/metrics"
	"github.com/ipfs-force-community/metrics/ratelimit"
	logging "github.com/ipfs/go-log/v2"
	"go.opencensus.io/plugin/ochttp"
)

func serveRPC(ctx context.Context, authEndpoint, rateLimitRedis, listen string, mCnf *metrics.TraceConfig, jwt jwtclient.IJwtAuthClient, full api.FullNode, stop dix.StopFunc, maxRequestSize int64) error {
	serverOptions := []jsonrpc.ServerOption{}
	if maxRequestSize > 0 {
		serverOptions = append(serverOptions, jsonrpc.WithMaxRequestSize(maxRequestSize))
	}

	rpcServer := jsonrpc.NewServer(serverOptions...)
	rpcServer2 := jsonrpc.NewServer(serverOptions...)

	var remoteJwtCli *jwtclient.AuthClient
	if len(authEndpoint) > 0 {
		remoteJwtCli, _ = jwtclient.NewAuthClient(authEndpoint)
	}

	// register hander to verify token in venus-auth
	var handler, handler2 http.Handler
	if remoteJwtCli != nil {
		handler = (http.Handler)(jwtclient.NewAuthMux(jwt, jwtclient.WarpIJwtAuthClient(remoteJwtCli), rpcServer))
		handler2 = (http.Handler)(jwtclient.NewAuthMux(jwt, jwtclient.WarpIJwtAuthClient(remoteJwtCli), rpcServer2))
	} else {
		handler = (http.Handler)(jwtclient.NewAuthMux(jwt, nil, rpcServer))
		handler2 = (http.Handler)(jwtclient.NewAuthMux(jwt, nil, rpcServer2))
	}

	if repoter, err := metrics.RegisterJaeger(mCnf.ServerName, mCnf); err != nil {
		log.Fatalf("register %s JaegerRepoter to %s failed:%s", mCnf.ServerName, mCnf.JaegerEndpoint, err)
	} else if repoter != nil {
		log.Infof("register jaeger-tracing exporter to %s, with node-name:%s", mCnf.JaegerEndpoint, mCnf.ServerName)
		defer metrics.UnregisterJaeger(repoter)
		handler = &ochttp.Handler{Handler: handler}
		handler2 = &ochttp.Handler{Handler: handler2}
	}

	limitWrapper := full
	if len(rateLimitRedis) > 0 && remoteJwtCli != nil {
		log.Infof("use rate limit %s", rateLimitRedis)
		limiter, err := ratelimit.NewRateLimitHandler(
			rateLimitRedis,
			nil, &jwtclient.ValueFromCtx{},
			jwtclient.WarpLimitFinder(remoteJwtCli),
			logging.Logger("rate-limit"))
		_ = logging.SetLogLevel("rate-limit", "debug")
		if err != nil {
			return err
		}

		var rateLimitAPI api.FullNodeStruct
		limiter.WarpFunctions(full, &rateLimitAPI.Internal)
		limiter.WarpFunctions(full, &rateLimitAPI.VenusAPIStruct.Internal)
		limiter.WarpFunctions(full, &rateLimitAPI.CommonStruct.Internal)
		limitWrapper = &rateLimitAPI
	}

	pma := api.PermissionedFullAPI(limitWrapper)

	serveRpc := func(path string, hnd interface{}, handler http.Handler, rpcSer *jsonrpc.RPCServer) {
		rpcSer.Register("Filecoin", hnd)
		http.Handle(path, handler)
	}

	serveRpc("/rpc/v0", &v0api.WrapperV1Full{FullNode: pma}, handler, rpcServer)
	serveRpc("/rpc/v1", pma, handler2, rpcServer2)

	server := http.Server{
		Addr:    listen,
		Handler: http.DefaultServeMux,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	sigCh := make(chan os.Signal, 2)

	go func() {
		select {
		case <-ctx.Done():

		case sig := <-sigCh:
			log.Infof("signal %s captured", sig)
		}

		if err := server.Shutdown(context.Background()); err != nil {
			log.Warnf("shutdown http server: %s", err)
		}

		if err := stop(context.Background()); err != nil {
			log.Warnf("call app stop func: %s", err)
		}

		log.Sync() // nolint:errcheck
	}()

	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	log.Infow("start http server", "addr", listen)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	log.Info("gracefull down")
	return nil
}
