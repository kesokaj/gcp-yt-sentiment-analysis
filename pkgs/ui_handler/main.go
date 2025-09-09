package ui_handler

import (
	"app/pkgs/bq_ingest"
	"app/pkgs/gemini_magic"
	"app/pkgs/models"
	"app/pkgs/shared"
	"app/pkgs/yt_video"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

// ServeUI serves the main HTML page for the user interface.
func ServeUI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/ui.html")
}

// ProcessHandler orchestrates the multi-step analysis and streams progress via SSE.
func ProcessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	youtubeURL := r.URL.Query().Get("url")
	if youtubeURL == "" {
		sendSSEMessage(w, flusher, map[string]string{"status": "error", "message": "Missing 'url' query parameter"})
		return
	}

	videoID, err := extractVideoID(youtubeURL)
	if err != nil {
		sendSSEMessage(w, flusher, map[string]string{"status": "error", "message": err.Error()})
		return
	}
	sendSSEMessage(w, flusher, map[string]string{"status": "processing", "message": fmt.Sprintf("Extracted Video ID: %s", videoID)})

	ctx := r.Context()

	// --- Step 1: Fetch YouTube Data ---
	sendSSEMessage(w, flusher, map[string]string{"status": "processing", "message": "Step 1/3: Fetching YouTube data..."})
	youtubeEndpoint := fmt.Sprintf("/youtube?videoId=%s", videoID)
	var step1Response models.APIResponse
	if err := callHandler(ctx, yt_video.FetchData(&shared.AppConfig), youtubeEndpoint, &step1Response); err != nil {
		sendSSEMessage(w, flusher, map[string]string{"status": "error", "message": "Step 1 failed: " + err.Error()})
		return
	}
	step1Message := fmt.Sprintf("Step 1/3 succeeded: %s (Tracking ID: %s, Time: %s)", step1Response.Message, step1Response.TrackingID, step1Response.ProcessingTime)
	sendSSEMessage(w, flusher, map[string]string{"status": "success", "message": step1Message})

	nextAction := step1Response.NextActionURI
	if nextAction == "" {
		sendSSEMessage(w, flusher, map[string]string{"status": "error", "message": "Error: Step 1 response did not contain a valid next action."})
		return
	}
	sendSSEMessage(w, flusher, map[string]string{"status": "processing", "message": fmt.Sprintf("Next action: %s", nextAction)})

	// --- Step 2: Analyze with Gemini ---
	sendSSEMessage(w, flusher, map[string]string{"status": "processing", "message": "Step 2/3: Analyzing data with Gemini..."})
	var step2Response models.APIResponse
	if err := callHandler(ctx, gemini_magic.AnalyzeData(&shared.AppConfig), nextAction, &step2Response); err != nil {
		sendSSEMessage(w, flusher, map[string]string{"status": "error", "message": "Step 2 failed: " + err.Error()})
		return
	}
	step2Message := fmt.Sprintf("Step 2/3 succeeded: %s (Time: %s)", step2Response.Message, step2Response.ProcessingTime)
	sendSSEMessage(w, flusher, map[string]string{"status": "success", "message": step2Message})

	nextAction = step2Response.NextActionURI
	if nextAction == "" {
		sendSSEMessage(w, flusher, map[string]string{"status": "error", "message": "Error: Step 2 response did not contain a valid next action."})
		return
	}
	sendSSEMessage(w, flusher, map[string]string{"status": "processing", "message": fmt.Sprintf("Next action: %s", nextAction)})

	// --- Step 3: Ingest into BigQuery ---
	sendSSEMessage(w, flusher, map[string]string{"status": "processing", "message": "Step 3/3: Ingesting data into BigQuery..."})
	var step3Response models.APIResponse
	if err := callHandler(ctx, bq_ingest.IngestData(&shared.AppConfig), nextAction, &step3Response); err != nil {
		sendSSEMessage(w, flusher, map[string]string{"status": "error", "message": "Step 3 failed: " + err.Error()})
		return
	}
	step3Message := fmt.Sprintf("Step 3/3 succeeded: %s (Time: %s)", step3Response.Message, step3Response.ProcessingTime)
	sendSSEMessage(w, flusher, map[string]string{"status": "success", "message": step3Message})

	sendSSEMessage(w, flusher, map[string]string{"status": "complete", "message": "All steps completed successfully!"})
}

// callHandler invokes another HTTP handler in-process, avoiding network overhead.
func callHandler(ctx context.Context, handler http.HandlerFunc, targetURL string, targetStruct interface{}) error {
	req := httptest.NewRequest("GET", targetURL, nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusOK {
		return fmt.Errorf("handler returned non-200 status: %d, body: %s", rr.Code, rr.Body.String())
	}

	if err := json.Unmarshal(rr.Body.Bytes(), targetStruct); err != nil {
		return fmt.Errorf("failed to unmarshal handler response: %w. Body: %s", err, rr.Body.String())
	}

	// Check for application-level success status from the APIResponse
	if response, ok := targetStruct.(*models.APIResponse); ok {
		if response.Status != "success" && response.Status != "skipped" {
			return fmt.Errorf("handler returned application error: %s", response.Message)
		}
	}

	return nil
}

func sendSSEMessage(w http.ResponseWriter, flusher http.Flusher, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf(`{"message": "Failed to marshal SSE data: %v", "severity": "ERROR"}`, err)
		return
	}

	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()
}

func extractVideoID(youtubeURL string) (string, error) {
	u, err := url.Parse(youtubeURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if strings.Contains(u.Host, "youtube.com") {
		if videoID := u.Query().Get("v"); videoID != "" {
			return videoID, nil
		}
	}
	if strings.Contains(u.Host, "youtu.be") {
		if videoID := strings.TrimPrefix(u.Path, "/"); videoID != "" {
			return videoID, nil
		}
	}
	return "", errors.New("could not find video ID in URL")
}
