package shared

import (
	"encoding/json"
	"log"
)

func LogJSON(severity string, message string, trackingID string) {
	logEntry := map[string]string{
		"severity":   severity,
		"message":    message,
		"trackingId": trackingID,
	}
	logBytes, _ := json.Marshal(logEntry)
	log.Println(string(logBytes))
}
