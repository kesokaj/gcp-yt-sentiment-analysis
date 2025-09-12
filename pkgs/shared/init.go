package shared

import (
	"log"
	"log/slog"
	"os"
)

func init() {
	// 1. Initialize logger
	Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(Logger)

	// 2. Initialize config
	AppConfig.YTApiKey = GetEnvString("YOUTUBE_API_KEY", "")
	AppConfig.GEMINIApiKey = GetEnvString("GEMINI_API_KEY", "")

	if AppConfig.YTApiKey == "" || AppConfig.GEMINIApiKey == "" {
		log.Fatal("CRITICAL: YOUTUBE_API_KEY and GEMINI_API_KEY environment variables must be set.")
	}

	AppConfig.GCSBucketName = GetEnvString("GCS_BUCKET_NAME", "yt-sentiment-bucket")
	AppConfig.GCPProject = GetEnvString("GCP_PROJECT", "")
	AppConfig.GCPLocation = GetEnvString("GCP_LOCATION", "us-central1")
	AppConfig.BQDataset = GetEnvString("BQ_DATASET", "yt_sentiment_data")
	AppConfig.GEMINIModel = GetEnvString("GEMINI_MODEL", "gemini-1.5-pro-latest")
	AppConfig.MaxCommentsToFetch = GetEnvInt("MAX_COMMENTS_TO_FETCH", 5000)
	AppConfig.Port = GetEnvString("PORT", "8080")
}
