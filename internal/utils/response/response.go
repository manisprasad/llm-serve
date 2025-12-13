package response

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Data   any    `json:"data,omitempty"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func WriteJson(w http.ResponseWriter, httpStatus int, data any, errMsg string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	resp := APIResponse{
		Status: "success",
		Data:   data,
	}

	if errMsg != "" {
		resp.Status = "error"
		resp.Error = errMsg
		resp.Data = nil
	}

	return json.NewEncoder(w).Encode(resp)
}
