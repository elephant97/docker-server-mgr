package container

import (
	"encoding/json"
	"fmt"
	"net/http"

	"docker-server-mgr/internal/appctx"
	"docker-server-mgr/internal/common/request"
	"docker-server-mgr/internal/common/response"
	"docker-server-mgr/internal/dockerops"
	"docker-server-mgr/internal/dockerops/types"
	"docker-server-mgr/internal/mysqlops"
)

func StartHandler(deps *appctx.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req request.StartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		status, err := handleStartRequest(deps, req)
		if err != nil {
			response.WriteResponse(w, status, err.Error())
			return
		}

		response.WriteResponse(w, status, "Conatiner Start Success")
	}
}

func handleStartRequest(
	deps *appctx.Dependencies,
	req request.StartRequest,
) (int, error) {
	containerStatus, err := mysqlops.SelectQueryRowsToStructs[types.ContainerStatus](deps.MySQLClient,
		"SELECT status FROM containers WHERE id = ? and status = 'Exited'",
		req.ContainerId)
	if err != nil || len(containerStatus) < 1 || !containerStatus[0].Status.Valid {
		return http.StatusInternalServerError, fmt.Errorf("failed to fetch container status")
	}

	if err := dockerops.StartContainer(deps, req.ContainerId); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed start container")
	}

	// mysql db update는 monitoring go루틴에서 진행하도록 함

	return http.StatusOK, nil
}
