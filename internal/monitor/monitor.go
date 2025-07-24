package monitor

import (
	"context"
	"time"

	"docker-server-mgr/internal/appctx"
	"docker-server-mgr/internal/dockerops"
	"docker-server-mgr/internal/dockerops/types"
	"docker-server-mgr/internal/mysqlops"
	"docker-server-mgr/utils"
	clog "docker-server-mgr/utils/log" //custom log
)

func CheckDockerStatus(
	ctx context.Context,
	deps *appctx.Dependencies,
) {
	clog.Debug("Starting Docker CheckDockerStatus...")

	for {
		select {
		case <-ctx.Done():
			clog.Error("Docker Status watcher stopped.")
			return
		default:
			containers, err := mysqlops.SelectQueryRowsToStructs[types.ContainerInfo](deps.MySQLClient, "SELECT id, status FROM containers WHERE deleted_at IS NULL or status != 'deleted'")
			if err != nil {
				clog.Error("Error fetching containers from MySQL", "err", err)
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
	clog.Debug("Starting Docker CheckImageStatus...")

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// 1시간마다 실행
	for {
		select {
		case <-ctx.Done():
			clog.Error("Docker Image Status watcher stopped.")
			return
		case <-ticker.C:
			clog.Info("Running WatchImageUsingStatus (every 1 hour)")
			dockerops.WatchImageUsingStatus(ctx, deps.DockerClient, deps.MySQLClient)
		}
	}
}
