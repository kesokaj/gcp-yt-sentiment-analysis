package yt_video

import (
	"app/pkgs/models"
	"app/pkgs/shared"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	youtubeService *youtube.Service
	serviceOnce    sync.Once
	serviceErr     error
)

func getYouTubeService(ctx context.Context, apiKey string) (*youtube.Service, error) {
	serviceOnce.Do(func() {
		youtubeService, serviceErr = youtube.NewService(ctx, option.WithAPIKey(apiKey))
	})
	return youtubeService, serviceErr
}

func FetchData(cfg *models.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		trackingID := r.URL.Query().Get("trackingId")
		if trackingID == "" {
			trackingID = uuid.New().String()
		}
		shared.LogJSON("INFO", fmt.Sprintf("Received request: %s %s", r.Method, r.URL.String()), trackingID)

		runDate := time.Now().Format("2006-01-02")
		videoId := r.URL.Query().Get("videoId")
		if videoId == "" {
			shared.LogJSON("WARNING", "Missing 'videoId' query parameter", trackingID)
			shared.JSONErrorResponse(w, trackingID, http.StatusBadRequest, "Missing 'videoId' query parameter")
			return
		}

		ctx := r.Context()

		ytService, err := getYouTubeService(ctx, cfg.YTApiKey)
		if err != nil {
			shared.LogJSON("ERROR", "Unable to create YouTube service: "+err.Error(), trackingID)
			shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Internal Server Error: Unable to create YouTube service")
			return
		}

		videoCall := ytService.Videos.List([]string{"snippet", "contentDetails", "statistics"}).Id(videoId).Context(ctx)
		videoResponse, err := videoCall.Do()
		if err != nil {
			shared.LogJSON("ERROR", "Error fetching video details: "+err.Error(), trackingID)
			shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to fetch video details")
			return
		}
		if len(videoResponse.Items) == 0 {
			shared.LogJSON("INFO", "Video not found for videoId: "+videoId, trackingID)
			shared.JSONErrorResponse(w, trackingID, http.StatusNotFound, "Video not found")
			return
		}
		shared.LogJSON("INFO", "Successfully fetched video details for videoId: "+videoId, trackingID)

		video := videoResponse.Items[0]

		thumbnailURL := ""
		if video.Snippet.Thumbnails.Maxres != nil {
			thumbnailURL = video.Snippet.Thumbnails.Maxres.Url
		} else if video.Snippet.Thumbnails.Standard != nil {
			thumbnailURL = video.Snippet.Thumbnails.Standard.Url
		} else if video.Snippet.Thumbnails.High != nil {
			thumbnailURL = video.Snippet.Thumbnails.High.Url
		}

		data := models.VideoData{
			ID:            video.Id,
			ChannelID:     video.Snippet.ChannelId,
			ChannelTitle:  video.Snippet.ChannelTitle,
			TrackingID:    trackingID,
			RunDate:       runDate,
			Title:         video.Snippet.Title,
			Description:   video.Snippet.Description,
			ThumbnailURL:  thumbnailURL,
			Duration:      video.ContentDetails.Duration,
			CategoryID:    video.Snippet.CategoryId,
			ViewCount:     int64(video.Statistics.ViewCount),
			LikeCount:     int64(video.Statistics.LikeCount),
			FavoriteCount: int64(video.Statistics.FavoriteCount),
			CommentCount:  int64(video.Statistics.CommentCount),
			Comments:      []*models.Comment{},
		}

		var comments []*models.Comment
		videoChannelId := video.Snippet.ChannelId
		nextPageToken := ""

		shared.LogJSON("INFO", "Fetching comments ordered by 'relevance'. Note: This may not retrieve all available comments.", trackingID)

	FetchCommentsLoop:
		for {
			call := ytService.CommentThreads.List([]string{"snippet", "replies"}).
				VideoId(videoId).
				TextFormat("plainText").
				MaxResults(100).
				Order("relevance")

			if nextPageToken != "" {
				call = call.PageToken(nextPageToken)
			}

			response, err := call.Context(ctx).Do()
			if err != nil {
				if strings.Contains(err.Error(), "quotaExceeded") {
					shared.LogJSON("WARNING", "YouTube API quota exceeded while fetching comments. Proceeding with fetched comments.", trackingID)
					break FetchCommentsLoop
				}
				shared.LogJSON("ERROR", "Error fetching comments: "+err.Error(), trackingID)
				shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to fetch comments")
				return
			}

			for _, item := range response.Items {
				topLevelComment := item.Snippet.TopLevelComment
				comments = append(comments, &models.Comment{
					ID:         topLevelComment.Id,
					ParentID:   "", // Top-level comments have no parent
					ChannelID:  videoChannelId,
					Text:       topLevelComment.Snippet.TextDisplay,
					LikeCount:  topLevelComment.Snippet.LikeCount,
					ReplyCount: item.Snippet.TotalReplyCount,
					TrackingID: trackingID,
					RunDate:    runDate,
				})
				if len(comments) >= cfg.MaxCommentsToFetch {
					break FetchCommentsLoop
				}

				if item.Replies != nil {
					for _, reply := range item.Replies.Comments {
						comments = append(comments, &models.Comment{
							ID:         reply.Id,
							ParentID:   topLevelComment.Id,
							ChannelID:  videoChannelId,
							Text:       reply.Snippet.TextDisplay,
							LikeCount:  reply.Snippet.LikeCount,
							ReplyCount: 0,
							TrackingID: trackingID,
							RunDate:    runDate,
						})
						if len(comments) >= cfg.MaxCommentsToFetch {
							break FetchCommentsLoop
						}
					}
				}
			}

			nextPageToken = response.NextPageToken
			if nextPageToken == "" {
				break
			}
		}

		if len(comments) >= cfg.MaxCommentsToFetch {
			shared.LogJSON("INFO", fmt.Sprintf("Reached comment fetch limit. Processing %d comments.", cfg.MaxCommentsToFetch), trackingID)
			comments = comments[:cfg.MaxCommentsToFetch]
		}

		data.Comments = comments
		shared.LogJSON("INFO", fmt.Sprintf("Successfully fetched %d comments for videoId: %s", len(data.Comments), videoId), trackingID)

		jsonData, err := json.Marshal(data)
		if err != nil {
			shared.LogJSON("ERROR", "Error marshalling data to JSON for GCS upload: "+err.Error(), trackingID)
			shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to marshal data for storage")
			return
		}

		objectName := trackingID + ".json"
		err = shared.UploadToGCS(ctx, cfg.GCSBucketName, objectName, jsonData)
		if err != nil {
			shared.LogJSON("ERROR", "GCS upload failed: "+err.Error(), trackingID)
			shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to save data to GCS")
			return
		}
		shared.LogJSON("INFO", fmt.Sprintf("Successfully uploaded data to gs://%s/%s", cfg.GCSBucketName, objectName), trackingID)

		processingTime := time.Since(startTime)
		nextActionURI := fmt.Sprintf("/magic?trackingId=%s", trackingID)
		response := models.APIResponse{
			TrackingID:     trackingID,
			ProcessingTime: processingTime.String(),
			Status:         "success",
			Message:        fmt.Sprintf("Successfully fetched %d comments.", len(data.Comments)),
			NextActionURI:  nextActionURI,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(response); err != nil {
			shared.LogJSON("ERROR", "Could not write JSON response: "+err.Error(), trackingID)
		}
	}
}
