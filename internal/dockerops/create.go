package dockerops

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"docker-server-mgr/internal/common/request"
	clog "docker-server-mgr/utils/log" //custom log
)

func BuildPortConfig(portMappings []request.PortMapping) (nat.PortSet, nat.PortMap, error) {
	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}

	for _, pm := range portMappings {
		port := nat.Port(fmt.Sprintf("%s/tcp", pm.ContainerPort))
		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{{
			HostIP:   "0.0.0.0",
			HostPort: fmt.Sprintf("%s", pm.HostPort),
		}}
	}

	return exposedPorts, portBindings, nil
}

func PrepareImage(cli *client.Client, ctx context.Context, image string) error {
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	imageExists := false
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == image {
				imageExists = true
				break
			}
		}
		if imageExists {
			break
		}
	}

	if !imageExists {
		clog.Debug("Image %s not found locally. Pulling...", image)

		reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}
		defer reader.Close()
		io.Copy(io.Discard, reader)
		clog.Debug("Image %s pulled successfully", image)
	}

	return nil
}

func CreateContainer(
	ctx context.Context,
	cli *client.Client,
	image string,
	req request.CreateRequest,
	exposed nat.PortSet,
	bindings nat.PortMap,
) (string, error) {

	// 컨테이너 생성
	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:        image,
			Cmd:          req.Cmd,
			ExposedPorts: exposed,
		},
		&container.HostConfig{
			PortBindings: bindings,
			Binds:        req.Volumes, // 호스트 디렉토리 바인딩 (필요시 설정)
		},
		nil,               // networkingConfig
		nil,               // platform (또는 runtime.GOOS)
		req.ContainerName, // container name
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// 컨테이너 시작
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return resp.ID, err
	}

	// 컨테이너 ID 반환
	return resp.ID, nil
}
