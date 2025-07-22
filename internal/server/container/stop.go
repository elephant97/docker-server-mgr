package container

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"docker-server-mgr/internal/appctx"
	"docker-server-mgr/internal/common/request"
	"docker-server-mgr/internal/common/response"
	"docker-server-mgr/internal/dockerops/types"
	"docker-server-mgr/internal/mysqlops"
)

func StopHandler(deps *appctx.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req request.StopRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		status, err := handleStopRequest(context.Background(), deps, req)
		if err != nil {
			response.WriteResponse(w, status, err.Error())
			return
		}

		response.WriteResponse(w, status, "Conatiner Stop immediate Success")
	}
}

func handleStopRequest(
	ctx context.Context,
	deps *appctx.Dependencies,
	req request.StopRequest,
) (int, error) {
	containerStatus, err := mysqlops.SelectQueryRowsToStructs[types.ContainerStatus](deps.MySQLClient,
		"SELECT status FROM containers WHERE id = ? and status = 'running'",
		req.ContainerId)
	if err != nil || len(containerStatus) < 1 || !containerStatus[0].Status.Valid {
		return http.StatusInternalServerError, fmt.Errorf("failed to fetch container status")
	}

	timeout := 10 * time.Second
	if err := deps.DockerClient.ContainerStop(ctx, req.ContainerId, &timeout); err != nil {
		return http.StatusBadRequest, fmt.Errorf("container %s stop failed, err: %v", containerStatus[0].Name.String, err)
	}

	// mysql db update는 monitoring go루틴에서 진행하도록 함

	return http.StatusOK, nil
}
