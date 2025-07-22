package dockerops

import (
	"context"
	"fmt"
	"strings" //custom log

	"docker-server-mgr/internal/dockerops/types"
	"docker-server-mgr/utils"
	clog "docker-server-mgr/utils/log" //custom log

	"github.com/docker/docker/client"
)

// NewDockerClient creates a new Docker API client using environment variables.
func NewDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(
		client.FromEnv,                     // 환경 변수에서 Docker API 설정 읽음.
		client.WithAPIVersionNegotiation(), // API 버전 조회해서 호환 가능한 버전으로 통신하게 함.
	)
}

func GetContainerStatus(
	ctx context.Context,
	cli *client.Client,
	containerID string,
) (string, error) {

	containerJSON, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		clog.Error("Error inspecting container", "containerID", containerID, "err", err)
		return "", fmt.Errorf("inspect error: %w", err)
	}

	return containerJSON.State.Status, nil
}

func GetContainerName(
	ctx context.Context,
	cli *client.Client,
	containerID string,
) (string, error) {
	inspect, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		clog.Error("Error inspecting container", "containerID", containerID, "err", err)
		return "", err
	}
	name := inspect.Name
	clog.Info("Container name", "conatainerName", strings.TrimPrefix(name, "/"))

	return strings.TrimPrefix(name, "/"), nil
}

func GetContainerAllInfo(
	ctx context.Context,
	cli *client.Client,
	containerID string,
) (types.ContainerAllInfo, error) {
	var containerInfo types.ContainerAllInfo
	inspect, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		clog.Error("Error inspecting container", "containerID", containerID, "err", err)
		return containerInfo, err
	}
	containerInfo.Name = strings.TrimPrefix(inspect.Name, "/")

	// Status
	containerInfo.Status = inspect.State.Status

	// CreatedAt
	containerInfo.CreateAt, err = utils.ConvertRFC3339ToDatetime(inspect.Created)
	if err != nil {
		containerInfo.CreateAt = ""
	}

	// Image & Tag
	image := inspect.Config.Image // 예: "nginx:latest"
	split := strings.Split(image, ":")
	containerInfo.Image = split[0]
	if len(split) > 1 {
		if len(split[1]) <= 0 {
			containerInfo.Tag = "latest"
		} else {
			containerInfo.Tag = split[1]
		}
	}

	// Ports
	ports := []types.PortMapping{}
	for containerPort, bindings := range inspect.NetworkSettings.Ports {
		for _, binding := range bindings {
			parts := strings.Split(string(containerPort), "/")
			port := parts[0]
			ports = append(ports, types.PortMapping{
				HostPort:      binding.HostPort,
				ContainerPort: port,
			})
		}
	}
	containerInfo.Ports = ports

	return containerInfo, nil
}
