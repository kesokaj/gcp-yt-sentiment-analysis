package main

import (
	"app/pkgs/bq_ingest"
	"app/pkgs/gemini_magic"
	"app/pkgs/shared"
	"app/pkgs/ui_handler"
	"app/pkgs/yt_video"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func init() {
	log.SetFlags(0)
}

func main() {
	http.HandleFunc("/", info)
	http.HandleFunc("/ui", ui_handler.ServeUI)
	http.HandleFunc("/ui/process", ui_handler.ProcessHandler)
	http.HandleFunc("/youtube", yt_video.FetchData(&shared.AppConfig))
	http.HandleFunc("/magic", gemini_magic.AnalyzeData(&shared.AppConfig))
	http.HandleFunc("/ingest", bq_ingest.IngestData(&shared.AppConfig))

	logEntry := map[string]string{
		"severity": "INFO",
		"message":  "Starting server on http://localhost:" + shared.AppConfig.Port,
	}
	logBytes, _ := json.Marshal(logEntry)
	log.Println(string(logBytes))

	server := &http.Server{
		Addr:         ":" + shared.AppConfig.Port,
		Handler:      http.DefaultServeMux,
		WriteTimeout: 35 * time.Minute,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		errorLogEntry := map[string]interface{}{
			"severity": "ERROR",
			"message":  "Failed to start server",
			"error":    err.Error(),
		}
		errorLogBytes, _ := json.Marshal(errorLogEntry)
		log.Println(string(errorLogBytes))
		os.Exit(1)
	}
}

func info(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "YouTube Analysis Service Endpoints:")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "1. /youtube?videoId=<YOUTUBE_VIDEO_ID>[&trackingId=<UUID>]")
	fmt.Fprintln(w, "   - Fetches video details, statistics, and the most relevant comments for the given YouTube video ID.")
	fmt.Fprintln(w, "   - Generates a unique 'trackingId' if one is not provided.")
	fmt.Fprintln(w, "   - Saves the complete data as a JSON file to GCS: gs://<bucket>/<trackingId>.json")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "2. /magic?trackingId=<TRACKING_ID>")
	fmt.Fprintln(w, "   - Retrieves the JSON file from GCS using the 'trackingId'.")
	fmt.Fprintln(w, "   - Sends the data to the Gemini API for a comprehensive marketing and sentiment analysis.")
	fmt.Fprintln(w, "   - Saves the resulting analysis as a new JSON file to GCS: gs://<bucket>/<trackingId>_analyzed.json")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "3. /ingest?trackingId=<TRACKING_ID>")
	fmt.Fprintln(w, "   - Reads both the raw data (<trackingId>.json) and the analyzed data (<trackingId>_analyzed.json) from GCS.")
	fmt.Fprintln(w, "   - Ingests the raw data into the 'videos' and 'comments' tables in BigQuery.")
	fmt.Fprintln(w, "   - Ingests the analyzed data into the 'analyzed' table in BigQuery.")
	fmt.Fprintln(w, "   - The endpoint is idempotent and will skip ingestion if the data for the trackingId already exists in BigQuery.")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "4. /ui")
	fmt.Fprintln(w, "   - Serves a web interface to run the full analysis pipeline.")
}
