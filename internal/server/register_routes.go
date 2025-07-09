package server

import (
	"net/http"

	"docker-server-mgr/internal/appctx"
	"docker-server-mgr/internal/server/container"
)

func RegisterContainerRoutes(mux *http.ServeMux, deps *appctx.Dependencies) {
	mux.HandleFunc("/v1/container/create", container.CreateHandler(deps))
}
