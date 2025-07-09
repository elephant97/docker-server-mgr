package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type APIResponse struct {
	Status  string      `json:"status"`            // "success" or "error"
	Message string      `json:"message,omitempty"` // only for errors
	Data    interface{} `json:"data,omitempty"`    // payload (for success)
}

func WriteResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := APIResponse{}
	if statusCode >= 200 && statusCode < 300 {
		resp.Status = "success"
		resp.Data = payload
	} else {
		resp.Status = "error"
		switch v := payload.(type) {
		case error:
			resp.Message = v.Error()
		case string:
			resp.Message = v
		default:
			resp.Message = fmt.Sprintf("%v", v)
		}
	}

	_ = json.NewEncoder(w).Encode(resp)
}
