package dockerops

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/redis/go-redis/v9"

	"docker-server-mgr/internal/mysqlops"
	"docker-server-mgr/internal/redisops"
)

// RemoveContainer forcibly removes a Docker container by its ID.
func RemoveContainer(
	ctx context.Context,
	dockerClient *client.Client,
	mysqlClient *sql.DB,
	redisClient *redis.Client,
	containerID string,
) {
	log.Printf("Removing: %s\n", containerID)

	err := dockerClient.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force: true, // 강제 삭제 (실행 중이어도 삭제)
	})
	if err != nil {
		log.Printf("Failed to remove container %s: %v", containerID, err)
		redisops.RegisterContainerTTL(ctx, redisClient, containerID, 10*time.Second) // Redis에 TTL 10으로 등록하고 다음에 재시도
		return                                                                       // 에러가 발생하면 로그에 남기고 종료
	} else if client.IsErrNotFound(err) {
		log.Printf("Notfound container %s: %v", containerID, err)
	} else {
		log.Printf("Container %s removed successfully", containerID)
	}

	_, err = mysqlops.ExecQuery(mysqlClient, "UPDATE containers SET deleted_at = NOW(), last_check_time = NOW(), status = 'deleted' WHERE id = ?", containerID)
	if err != nil {
		log.Printf("Failed to update MySQL for container %s: %v", containerID, err)
	} else {
		log.Printf("MySQL updated for removed container %s", containerID)
	}

}
