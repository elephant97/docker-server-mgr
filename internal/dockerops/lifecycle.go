package dockerops

import (
	"context"
	"database/sql"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"docker-server-mgr/internal/mysqlops"
	clog "docker-server-mgr/utils/log" //custom log
)

func WatchDockerStatus(
	ctx context.Context,
	dockerClient *client.Client,
	mysqlClient *sql.DB,
	userContainers map[string]string, // Map of container ID to status
) {
	clog.Info("Starting Docker lifecycle watcher...")

	// Check for container status changes
	containers, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		clog.Error("Error listing containers: %v", err)
	}

	dockerIDs := compareUserAndDockerContainers(ctx, containers, userContainers)

	for _, container := range containers {
		status, err := GetContainerStatus(ctx, dockerClient, container.ID)
		if err != nil {
			clog.Error("Error getting status for container %s: %v", container.ID, err)
		}

		if _, exists := userContainers[container.ID]; !exists {
			clog.Warn("⚠️ 나중에는 삭제 처리 할 등록안된 컨테이너 :%s", container.ID) //TODO
			continue
		}

		if userContainers[container.ID] != status {
			clog.Info("Container status changed: %s from %s to %s", container.ID, userContainers[container.ID], status)
			updateConatinerStatus(mysqlClient, container.ID, status)
		}

		if status == "exited" || status == "dead" {
			clog.Info("Container %s has exited or is dead. Removing...", container.ID)
			// go RemoveContainer(ctx, dockerClient, container.ID)
		}
	}

	for containerId := range userContainers {
		if _, exists := dockerIDs[containerId]; !exists {
			clog.Info("🗑️ Docker에서 사라진 컨테이너 감지됨: %s → deleted 처리", containerId)
			_, err := mysqlops.ExecQuery(mysqlClient, "UPDATE containers SET status = 'deleted', last_check_time = NOW(), deleted_at = NOW() WHERE id = ?", containerId)
			if err != nil {
				clog.Error("❌ DB 삭제 마킹 실패: %s, error: %v", containerId, err)
			}
		}
	}

}

func updateConatinerStatus(mysqlClient *sql.DB,
	containerID string, status string) {
	clog.Debug("Updating container %s status to %s", containerID, status)

	_, err := mysqlops.ExecQuery(mysqlClient, "UPDATE containers SET status = ?, last_check_time = NOW() WHERE id = ?",
		status, containerID)

	if err != nil {
		clog.Error("Error updating container status in MySQL: %v", err)
	} else {
		clog.Info("Container %s status updated to %s successfully", containerID, status)
	}
}

func compareUserAndDockerContainers(
	ctx context.Context,
	dockerContainerList []types.Container,
	userContainers map[string]string,
) map[string]bool {
	clog.Debug("Comparing user containers with Docker containers...")
	dockerIDs := make(map[string]bool)

	for _, c := range dockerContainerList {
		if _, exists := userContainers[c.ID]; exists {
			dockerIDs[c.ID] = true
		} else {
			dockerIDs[c.ID] = false
		}

	}

	return dockerIDs
}
