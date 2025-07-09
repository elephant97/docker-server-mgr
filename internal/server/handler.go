package server

import (
	"context"
	"log"
	"net/http"

	"docker-server-mgr/internal/appctx"
)

func StartHTTPServer(ctx context.Context, deps *appctx.Dependencies) {

	mux := http.NewServeMux()

	RegisterContainerRoutes(mux, deps)

	log.Println("Listening on :8082")
	if err := http.ListenAndServe(":8082", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
