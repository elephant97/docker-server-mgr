package container

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"docker-server-mgr/internal/appctx"
	"docker-server-mgr/internal/common/errs"
	"docker-server-mgr/internal/common/request"
	"docker-server-mgr/internal/common/response"
	"docker-server-mgr/internal/dockerops"
	"docker-server-mgr/internal/mysqlops"
	"docker-server-mgr/internal/redisops"
)

func CreateHandler(deps *appctx.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req request.CreateRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Validation request: %v", err)
			response.WriteResponse(w, http.StatusBadRequest, "Invalid request")
			return
		}

		errdefs := validateCreateRequest(&req)
		if errdefs.Code != 0 && errdefs.Code != int(errs.DefaultSet) {
			log.Printf("Validation error: %v", errdefs.Message)
			response.WriteResponse(w, http.StatusBadRequest, errdefs.Message.Error())
			return
		}

		containerID, status, err := handleCreateRequest(context.Background(), deps, req)
		if err != nil {
			response.WriteResponse(w, status, err.Error())
			return
		}

		response.WriteResponse(w, status, map[string]string{"container_id": containerID})
	}
}

func validateCreateRequest(req *request.CreateRequest) errs.ErrorDetail {
	if req.UserId == 0 {
		log.Println("User ID is 0")
		return errs.ErrorDetail{
			Code:    int(errs.Invaild),
			Message: fmt.Errorf("user ID is required"),
		}
	}

	log.Println("Validating create request:", req.UserId)

	if req.Image == "" {
		log.Println("Image is 0")
		return errs.ErrorDetail{
			Code:    int(errs.Invaild),
			Message: fmt.Errorf("image is required"),
		}
	}

	if req.ContainerName == "" {
		log.Println("Container name is required")
		req.ContainerName = fmt.Sprintf("container-%d", time.Now().UnixNano())
		return errs.ErrorDetail{
			Code:    int(errs.DefaultSet),
			Message: fmt.Errorf("ContainerName set default"),
		}
	}

	return errs.ErrorDetail{}
}

func handleCreateRequest(
	ctx context.Context,
	deps *appctx.Dependencies,
	req request.CreateRequest,
) (string, int, error) {
	exposed, bindings, err := dockerops.BuildPortConfig(req.Ports)
	if err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("port config error: %w", err)
	}

	tag := req.Tag
	if tag == "" {
		tag = "latest"
	}
	fullImage := fmt.Sprintf("%s:%s", req.Image, tag)

	if err := dockerops.PrepareImage(deps.DockerClient, ctx, fullImage); err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("image prepare error: %w", err)
	}

	containerID, err := dockerops.CreateContainer(ctx, deps.DockerClient, fullImage, req, exposed, bindings)
	if err != nil {
		failCreateTask(ctx, deps, containerID)
		return "", http.StatusBadRequest, fmt.Errorf("create error[%s]: %w", containerID, err)
	}

	log.Printf("Container created: %s", containerID)

	status, err := dockerops.GetContainerStatus(ctx, deps.DockerClient, containerID)
	if err != nil {
		failCreateTask(ctx, deps, containerID)
		return "", http.StatusInternalServerError, fmt.Errorf("failed to get container status: %w", err)
	} else if status != "running" {
		failCreateTask(ctx, deps, containerID)
		return "", http.StatusInternalServerError, fmt.Errorf("container %s can not running, status: %s", containerID, status)
	}

	tx, err := deps.MySQLClient.Begin()
	if err != nil {
		failCreateTask(ctx, deps, containerID)
		return "", http.StatusForbidden, fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = mysqlops.ExecQuery(deps.MySQLClient, "INSERT INTO containers (user_id, id, container_name, image, tag, ttl, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		req.UserId, containerID, req.ContainerName, req.Image, tag, req.TTL, status)
	if err != nil {
		failCreateTask(ctx, deps, containerID)
		return "", http.StatusForbidden, fmt.Errorf("mysql insert error: %w", err)
	}

	for _, port := range req.Ports {
		_, err := mysqlops.ExecQuery(deps.MySQLClient,
			"INSERT INTO container_ports (container_id, host_port, container_port) VALUES (?, ?, ?)",
			containerID, port.HostPort, port.ContainerPort)
		if err != nil {
			failCreateTask(ctx, deps, containerID)
			tx.Rollback()
			return "", http.StatusForbidden, fmt.Errorf("failed to insert port mapping: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", http.StatusForbidden, fmt.Errorf("failed to commit transaction: %w", err)
	}

	if req.TTL > 0 {
		err = redisops.RegisterContainerTTL(ctx, deps.RedisClient, containerID, time.Duration(req.TTL)*time.Second)
		if err != nil {
			failCreateTask(ctx, deps, containerID)
			tx.Rollback()
			return "", http.StatusForbidden, fmt.Errorf("redis TTL register error: %w", err)
		}
	}

	log.Printf("Container %s created and registered successfully", containerID)

	return containerID, http.StatusOK, nil
}

func failCreateTask(
	ctx context.Context,
	deps *appctx.Dependencies,
	containerID string,
) {
	// 항상 컨테이너 제거
	dockerops.RemoveContainer(ctx, deps, containerID)
}
