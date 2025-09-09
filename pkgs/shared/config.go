package shared

import "app/pkgs/models"

var AppConfig models.AppConfig

func init() {
	AppConfig.YTApiKey = GetEnvString("YOUTUBE_API_KEY", "")
	AppConfig.GEMINIApiKey = GetEnvString("GEMINI_API_KEY", "")
	AppConfig.GCSBucketName = GetEnvString("GCS_BUCKET_NAME", "yt-sentiment-bucket")
	AppConfig.GCPProject = GetEnvString("GCP_PROJECT", "")
	AppConfig.GCPLocation = GetEnvString("GCP_LOCATION", "us-central1")
	AppConfig.BQDataset = GetEnvString("BQ_DATASET", "yt_sentiment_data")
	AppConfig.GEMINIModel = GetEnvString("GEMINI_MODEL", "gemini-2.5-pro")
	AppConfig.MaxCommentsToFetch = GetEnvInt("MAX_COMMENTS_TO_FETCH", 5000)
	AppConfig.Port = GetEnvString("PORT", "8080")
}
