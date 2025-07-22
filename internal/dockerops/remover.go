package dockerops

import (
	"context"
	"strings"
	"time"

	"github.com/docker/docker/api/types"

	"docker-server-mgr/internal/appctx"
	"docker-server-mgr/internal/mysqlops"
	"docker-server-mgr/internal/redisops"
	clog "docker-server-mgr/utils/log" //custom log
)

// RemoveContainer forcibly removes a Docker container by its ID.
func RemoveContainer(
	ctx context.Context,
	deps *appctx.Dependencies,
	containerID string,
) {
	clog.Info("Removing: %s\n", containerID)

	err := deps.DockerClient.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force: true, // 강제 삭제 (실행 중이어도 삭제)
	})
	if err != nil {
		if strings.Contains(err.Error(), "No such container") {
			clog.Error("Notfound container %s: %v", containerID, err)
			return
		} else {
			clog.Error("Failed to remove container %s: %v", containerID, err)
			redisops.RegisterContainerTTL(ctx, deps.RedisClient, containerID, 10*time.Second) // Redis에 TTL 10으로 등록하고 다음에 재시도
			return                                                                            // 에러가 발생하면 로그에 남기고 종료
		}

	} else {
		clog.Info("Container %s removed successfully", containerID)
	}

	_, err = mysqlops.ExecQuery(deps.MySQLClient, "UPDATE containers SET deleted_at = NOW(), last_check_time = NOW(), status = 'deleted' WHERE id = ?", containerID)
	if err != nil {
		clog.Error("Failed to update MySQL for container %s: %v", containerID, err)
	} else {
		clog.Debug("MySQL updated for removed container %s", containerID)
	}

}
