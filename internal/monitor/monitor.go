package monitor

import (
	"context"
	"log"
	"time"

	"docker-server-mgr/internal/appctx"
	"docker-server-mgr/internal/dockerops"
	"docker-server-mgr/internal/dockerops/types"
	"docker-server-mgr/internal/mysqlops"
	"docker-server-mgr/utils"
)

func CheckDockerStatus(
	ctx context.Context,
	deps *appctx.Dependencies,
) {
	log.Println("Starting Docker CheckDockerStatus...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Docker Status watcher stopped.")
			return
		default:
			containers, err := mysqlops.SelectQueryRowsToStructs[types.ContainerInfo](deps.MySQLClient, "SELECT id, status FROM containers WHERE deleted_at IS NULL or status != 'deleted'")
			if err != nil {
				log.Printf("Error fetching containers from MySQL: %v", err)
				time.Sleep(10 * time.Second)
				continue
			}

			dockerops.WatchDockerStatus(ctx, deps.DockerClient, deps.MySQLClient, utils.ListToMap(containers,
				func(c types.ContainerInfo) string { return c.ID },
				func(c types.ContainerInfo) string {
					if c.Status.Valid {
						return c.Status.String
					}
					return ""
				},
			))
			time.Sleep(10 * time.Second)
		}
	}
}

func CheckImageStatus(
	ctx context.Context,
	deps *appctx.Dependencies,
) {
	log.Println("Starting Docker CheckImageStatus...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Docker Image Status watcher stopped.")
			return
		default:
			dockerops.WatchImageUsingStatus(ctx, deps.DockerClient, deps.MySQLClient)
			time.Sleep(60 * time.Second)
		}
	}
}
