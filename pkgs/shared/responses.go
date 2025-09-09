package shared

import (
	"app/pkgs/models"
	"encoding/json"
	"net/http"
)

func JSONErrorResponse(w http.ResponseWriter, trackingID string, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := models.APIResponse{
		TrackingID: trackingID,
		Status:     "error",
		Message:    message,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(response); err != nil {
		LogJSON("ERROR", "FATAL: could not write JSON error response: "+err.Error(), trackingID)
	}
}
