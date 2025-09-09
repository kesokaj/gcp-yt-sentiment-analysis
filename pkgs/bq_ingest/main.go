package bq_ingest

import (
	"app/pkgs/models"
	"app/pkgs/shared"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func recordExists(ctx context.Context, client *bigquery.Client, project, dataset, table, trackingID string) (bool, error) {
	queryStr := fmt.Sprintf(
		"SELECT COUNT(1) as count FROM `%s.%s.%s` WHERE tracking_id = @trackingID",
		project, dataset, table,
	)
	q := client.Query(queryStr)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "trackingID", Value: trackingID},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return false, fmt.Errorf("could not query table %s: %w", table, err)
	}

	var row struct{ Count int }
	if err := it.Next(&row); err != nil && err != iterator.Done {
		return false, fmt.Errorf("failed to read query result from table %s: %w", table, err)
	}

	return row.Count > 0, nil
}

func IngestData(cfg *models.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		ctx := r.Context()
		trackingID := r.URL.Query().Get("trackingId")
		if trackingID == "" {
			shared.LogJSON("WARNING", "Missing 'trackingId' query parameter", "")
			shared.JSONErrorResponse(w, "", http.StatusBadRequest, "Missing 'trackingId' query parameter")
			return
		}
		shared.LogJSON("INFO", fmt.Sprintf("Received request: %s %s", r.Method, r.URL.String()), trackingID)

		client, err := bigquery.NewClient(ctx, cfg.GCPProject)
		if err != nil {
			shared.LogJSON("ERROR", "could not create BigQuery client: "+err.Error(), trackingID)
			shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to connect to BigQuery")
			return
		}
		defer client.Close()

		var messages []string
		var ingestionOccurred bool

		rawExists, err := recordExists(ctx, client, cfg.GCPProject, cfg.BQDataset, "videos", trackingID)
		if err != nil {
			shared.LogJSON("ERROR", "could not query for existing raw data: "+err.Error(), trackingID)
			shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to query BigQuery for raw data")
			return
		}

		if rawExists {
			shared.LogJSON("INFO", "Raw data already exists in BigQuery. Skipping.", trackingID)
			messages = append(messages, fmt.Sprintf("Raw data for tracking ID %s already exists in BigQuery. Skipping.", trackingID))
		} else {
			shared.LogJSON("INFO", "New raw data tracking ID. Proceeding with ingestion.", trackingID)
			objectName := fmt.Sprintf("%s.json", trackingID)
			fileData, err := shared.GetFileFromGCS(ctx, cfg.GCSBucketName, objectName)
			if err != nil {
				shared.LogJSON("ERROR", "could not get raw data file from GCS: "+err.Error(), trackingID)
				shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to retrieve raw data file")
				return
			}

			var fullData models.VideoData
			if err := json.Unmarshal(fileData, &fullData); err != nil {
				shared.LogJSON("ERROR", "could not unmarshal raw data JSON: "+err.Error(), trackingID)
				shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Invalid raw data format")
				return
			}

			videoForBQ := models.VideoRecord{
				ID:            fullData.ID,
				ChannelID:     fullData.ChannelID,
				ChannelTitle:  fullData.ChannelTitle,
				TrackingID:    fullData.TrackingID,
				RunDate:       fullData.RunDate,
				Title:         fullData.Title,
				Description:   fullData.Description,
				ThumbnailURL:  fullData.ThumbnailURL,
				Duration:      fullData.Duration,
				CategoryID:    fullData.CategoryID,
				ViewCount:     fullData.ViewCount,
				LikeCount:     fullData.LikeCount,
				FavoriteCount: fullData.FavoriteCount,
				CommentCount:  fullData.CommentCount,
			}
			videoInserter := client.Dataset(cfg.BQDataset).Table("videos").Inserter()
			if err := videoInserter.Put(ctx, &videoForBQ); err != nil {
				shared.LogJSON("ERROR", "could not insert video data into BigQuery: "+err.Error(), trackingID)
				shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to ingest video data")
				return
			}
			messages = append(messages, fmt.Sprintf("Successfully ingested video data for video ID %s.", fullData.ID))
			ingestionOccurred = true

			if len(fullData.Comments) > 0 {
				commentsForBQ := fullData.Comments
				for i := range commentsForBQ {
					commentsForBQ[i].VideoID = fullData.ID
				}
				commentsInserter := client.Dataset(cfg.BQDataset).Table("comments").Inserter()
				if err := commentsInserter.Put(ctx, commentsForBQ); err != nil {
					shared.LogJSON("ERROR", "could not insert comments data into BigQuery: "+err.Error(), trackingID)
					shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to ingest comments data")
					return
				}
				messages = append(messages, fmt.Sprintf("Successfully ingested %d comments.", len(commentsForBQ)))
			}
		}

		analyzedExists, err := recordExists(ctx, client, cfg.GCPProject, cfg.BQDataset, "analyzed", trackingID)
		if err != nil {
			shared.LogJSON("ERROR", "could not query for existing analyzed data: "+err.Error(), trackingID)
			shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to query BigQuery for analyzed data")
			return
		}

		if analyzedExists {
			shared.LogJSON("INFO", "Analyzed data already exists in BigQuery. Skipping.", trackingID)
			messages = append(messages, fmt.Sprintf("Analyzed data for tracking ID %s already exists in BigQuery. Skipping.", trackingID))
		} else {
			analyzedObjectName := fmt.Sprintf("%s_analyzed.json", trackingID)
			fileData, err := shared.GetFileFromGCS(ctx, cfg.GCSBucketName, analyzedObjectName)
			if err != nil {
				if errors.Is(err, storage.ErrObjectNotExist) {
					msg := fmt.Sprintf("Analyzed data file %s not found in GCS. Skipping.", analyzedObjectName)
					shared.LogJSON("INFO", msg, trackingID)
					messages = append(messages, msg)
				} else {
					shared.LogJSON("ERROR", "could not get analyzed file from GCS: "+err.Error(), trackingID)
					shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to retrieve analyzed data file")
					return
				}
			} else {
				var analysisRecord models.AnalysisRecord
				if err := json.Unmarshal(fileData, &analysisRecord); err != nil {
					errorMsg := fmt.Sprintf("could not unmarshal analyzed data JSON: %s. Raw data: %s", err.Error(), string(fileData))
					shared.LogJSON("ERROR", errorMsg, trackingID)
					shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Invalid analyzed data format")
					return
				}

				inserter := client.Dataset(cfg.BQDataset).Table("analyzed").Inserter()
				if err := inserter.Put(ctx, &analysisRecord); err != nil {
					shared.LogJSON("ERROR", "could not insert analyzed data into BigQuery: "+err.Error(), trackingID)
					shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to ingest analyzed data")
					return
				}
				shared.LogJSON("INFO", "Successfully ingested analyzed data.", trackingID)
				messages = append(messages, fmt.Sprintf("Successfully ingested analyzed data for tracking ID %s.", trackingID))
				ingestionOccurred = true
			}
		}

		status := "skipped"
		if ingestionOccurred {
			status = "success"
		}

		response := models.APIResponse{
			TrackingID:     trackingID,
			ProcessingTime: time.Since(startTime).String(),
			Status:         status,
			Message:        strings.Join(messages, " "),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(response); err != nil {
			shared.LogJSON("ERROR", "could not write JSON response: "+err.Error(), trackingID)
		}
	}
}
