package dockerops

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"

	"docker-server-mgr/internal/appctx"
)

func StartContainer(
	deps *appctx.Dependencies,
	containerID string,
) error {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := deps.DockerClient.ContainerStart(timeoutCtx, containerID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	return nil
}
