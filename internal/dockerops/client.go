package dockerops

import (
	"context"
	"fmt"
	"log"
	"strings"

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
		log.Printf("Error inspecting container %s: %v", containerID, err)
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
		log.Printf("Error inspecting container %s: %v", containerID, err)
		return "", err
	}
	name := inspect.Name
	log.Println("Container name:", strings.TrimPrefix(name, "/"))

	return strings.TrimPrefix(name, "/"), nil
}
