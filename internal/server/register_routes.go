package server

import (
	"net/http"

	"docker-server-mgr/internal/appctx"
	"docker-server-mgr/internal/server/container"
)

func RegisterContainerRoutes(mux *http.ServeMux, deps *appctx.Dependencies) {
	mux.HandleFunc("/v1/container/create", container.CreateHandler(deps))
	mux.HandleFunc("/v1/container/delete", container.DeleteHandler(deps))
	mux.HandleFunc("/v1/container/stop", container.StopHandler(deps))
	mux.HandleFunc("/v1/container/start", container.StartHandler(deps))
}
