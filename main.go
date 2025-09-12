package main

import (
	"app/pkgs/bq_ingest"
	"app/pkgs/gemini_magic"
	"app/pkgs/shared"
	"app/pkgs/ui_handler"
	"app/pkgs/yt_video"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/", ui_handler.Info)
	http.HandleFunc("/ui", ui_handler.ServeUI)
	http.HandleFunc("/ui/process", ui_handler.ProcessHandler)
	http.HandleFunc("/youtube", yt_video.FetchData(&shared.AppConfig))
	http.HandleFunc("/magic", gemini_magic.AnalyzeData(&shared.AppConfig))
	http.HandleFunc("/ingest", bq_ingest.IngestData(&shared.AppConfig))

	slog.Info("Starting server", "port", shared.AppConfig.Port)

	server := &http.Server{
		Addr:         ":" + shared.AppConfig.Port,
		Handler:      http.DefaultServeMux,
		WriteTimeout: 35 * time.Minute,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
