package container

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"docker-server-mgr/internal/appctx"
	"docker-server-mgr/internal/common/request"
	"docker-server-mgr/internal/common/response"
	"docker-server-mgr/internal/dockerops/types"
	"docker-server-mgr/internal/mysqlops"
	"docker-server-mgr/internal/redisops"
)

func DeleteHandler(deps *appctx.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req request.DeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		status, err := handleDeleteRequest(context.Background(), deps, req)
		if err != nil {
			response.WriteResponse(w, status, err.Error())
			return
		}

		response.WriteResponse(w, status, "Container will be deleted in 1 sec")
	}
}

func handleDeleteRequest(
	ctx context.Context,
	deps *appctx.Dependencies,
	req request.DeleteRequest,
) (int, error) {

	containerStatus, err := mysqlops.SelectQueryRowsToStructs[types.ContainerStatus](deps.MySQLClient,
		"SELECT status FROM containers WHERE id = ? and deleted_at IS NULL",
		req.ContainerId)
	if err != nil || len(containerStatus) < 1 {
		return http.StatusInternalServerError, fmt.Errorf("failed to fetch container status")
	}

	if containerStatus[0].Status.Valid && containerStatus[0].Status.String == "deleted" {
		return http.StatusBadRequest, fmt.Errorf("container %s is already deleted", containerStatus[0].Name.String)
	}

	// Redis에 TTL 1로 등록하고 바로 삭제 처리 되도록 등록
	if err := redisops.RegisterContainerTTL(ctx, deps.RedisClient, req.ContainerId, 1); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to register container TTL in Redis: %w", err)
	}

	return http.StatusOK, nil
}
