package dockerops

import (
	"context"
	"database/sql"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	dtypes "docker-server-mgr/internal/dockerops/types"
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
		clog.Error("Error listing containers", "err", err)
	}

	dockerIDs := compareUserAndDockerContainers(ctx, containers, userContainers)

	for _, container := range containers {
		status, err := GetContainerStatus(ctx, dockerClient, container.ID)
		if err != nil {
			clog.Error("Error getting status for container",
				"containerID", container.ID,
				"error", err)
		}

		if _, exists := userContainers[container.ID]; !exists {
			clog.Warn("⚠️ 등록 안 된 컨테이너 추후 삭제 예정", "containerID", container.ID) //TODO
			containerAllInfo, err := GetContainerAllInfo(ctx, dockerClient, container.ID)
			if err != nil {
				clog.Error("unknown Conatiner GetContainerAllInfo failed", "err", err)
			}

			if err := unknownContainerRegister(mysqlClient, container.ID, containerAllInfo); err != nil {
				clog.Error("unknown Conatiner Register failed", "err", err)
			} else {
				clog.Info("unknown Conatiner Register success")
			}

			continue
		}

		if userContainers[container.ID] != status {
			clog.Info("Container status changed",
				"containerID", container.ID,
				"from", userContainers[container.ID],
				"to", status,
			)
			updateConatinerStatus(mysqlClient, container.ID, status)
		}

		if status == "exited" || status == "dead" {
			clog.Info("Container has exited or is dead. Removing...", "containerID", container.ID)
			// go RemoveContainer(ctx, dockerClient, container.ID)
		}
	}

	for containerId := range userContainers {
		if _, exists := dockerIDs[containerId]; !exists {
			clog.Info("🗑️ Docker에서 사라진 컨테이너 감지됨 deleted 처리", "containerId", containerId)
			_, err := mysqlops.ExecQuery(mysqlClient, "UPDATE containers SET status = 'deleted', last_check_time = NOW(), deleted_at = NOW() WHERE id = ?", containerId)
			if err != nil {
				clog.Error("❌ DB 삭제 마킹 실패", "containerId", containerId, "err", err)
			}
		}
	}

}

func updateConatinerStatus(mysqlClient *sql.DB,
	containerID string, status string) {
	clog.Debug("Updating container status", "containerID", containerID, "status", status)

	_, err := mysqlops.ExecQuery(mysqlClient, "UPDATE containers SET status = ?, last_check_time = NOW() WHERE id = ?",
		status, containerID)

	if err != nil {
		clog.Error("Error updating container status in MySQL", "err", err)
	} else {
		clog.Info("Container status updated successfully", "containerID", containerID, "status", status)
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

func unknownContainerRegister(mysqlClient *sql.DB, containerID string, containerAllInfo dtypes.ContainerAllInfo) error {
	tx, err := mysqlClient.Begin()
	if err != nil {
		return err
	}

	_, err = mysqlops.ExecQuery(mysqlClient, "INSERT INTO containers (user_id, id, container_name, image, tag, ttl, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		1, containerID, containerAllInfo.Name, containerAllInfo.Image, containerAllInfo.Tag, 0, containerAllInfo.Status, containerAllInfo.CreateAt)
	if err != nil {
		return err
	}

	for _, port := range containerAllInfo.Ports {
		_, err := mysqlops.ExecQuery(mysqlClient,
			"INSERT INTO container_ports (container_id, host_port, container_port) VALUES (?, ?, ?)",
			containerID, port.HostPort, port.ContainerPort)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}
