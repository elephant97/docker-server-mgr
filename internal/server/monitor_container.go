package server

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/docker/docker/client"
	"github.com/redis/go-redis/v9"

	"docker-server-mgr/internal/dockerops"
	"docker-server-mgr/internal/dockerops/types"
	"docker-server-mgr/internal/mysqlops"
	"docker-server-mgr/utils"
)

func CheckDockerLifecycle(
	ctx context.Context,
	dockerClient *client.Client,
	mysqlClient *sql.DB,
	redisClient *redis.Client,
) {
	log.Println("Starting Docker CheckDockerLifecycle...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Docker lifecycle watcher stopped.")
			return
		default:
			containers, err := mysqlops.SelectQueryRowsToStructs[types.ContainerInfo](mysqlClient, "SELECT id, status FROM containers WHERE deleted_at IS NULL or status != 'deleted'")
			if err != nil {
				log.Printf("Error fetching containers from MySQL: %v", err)
				time.Sleep(10 * time.Second)
				continue
			}

			dockerops.WatchDockerLifecycle(ctx, dockerClient, mysqlClient, utils.ListToMap(containers,
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
