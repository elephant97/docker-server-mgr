package server

import (
	"context"
	"net/http"

	"docker-server-mgr/internal/appctx"
	clog "docker-server-mgr/utils/log" //custom log
)

func StartHTTPServer(ctx context.Context, deps *appctx.Dependencies) {

	mux := http.NewServeMux()

	RegisterContainerRoutes(mux, deps)

	clog.Debug("Listening on :8082")
	if err := http.ListenAndServe(":8082", mux); err != nil {
		clog.Fatal("Server failed", "err", err)
	}
}
