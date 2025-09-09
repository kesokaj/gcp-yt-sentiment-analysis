package gemini_magic

import (
	"app/pkgs/models"
	"app/pkgs/shared"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/generative-ai-go/genai"
	"golang.org/x/time/rate"
	"google.golang.org/api/option"
)

const mapPrompt = `
	You are an expert YouTube marketing strategist and data analyst. Your task is to perform a partial analysis of the provided YouTube video data and a chunk of its comments. The goal is to produce a concise summary of this specific chunk, which will be used in a later step for a full analysis.

	**Input Data:**
	A JSON object containing details about a YouTube video and its comments will be provided.

	%s

	**Analysis Tasks & Output Structure:**
	Your entire output MUST be a single, minified JSON object. Your response must be raw JSON, starting with '{' and ending with '}'. All string values must be properly escaped.

	1.  **Sentiment Analysis ('sentiment_analysis'):**
	    *   'positive_comments': An integer count of positive comments in this chunk.
	    *   'negative_comments': An integer count of negative comments in this chunk.
	    *   'neutral_comments': An integer count of neutral comments in this chunk.
	    *   'summary': A brief 1-sentence summary of the sentiment in this chunk.

	2.  **Key Discussion Themes ('key_themes'):** Identify the top 3-5 dominant themes discussed in this comment chunk. For each theme, provide:
	    *   'theme_title': A short, descriptive title (e.g., 'Game Performance Issues').
	    *   'summary': A 1-3 sentence explanation of the theme based on comments in this chunk.
	    *   'representative_comment': The text of one comment from this chunk that best exemplifies this theme.

	3.  **Engagement Highlights ('engagement_highlights'):** Identify the top 2 comments from this chunk that generated the most engagement (likes or replies). For each comment, provide:
	    *   'comment_text': The full text of the comment.
	    *   'engagement_count': An integer representing the sum of likes and replies for that comment.
	    *   'reason_for_engagement': A brief explanation of *why* this comment was so engaging (e.g., 'Controversial opinion,' 'Humorous take,' 'Helpful advice').
	`

const reducePrompt = `
	You are an expert YouTube marketing strategist and data analyst. You have been provided with video metadata and a series of partial analyses from multiple chunks of that video's comments. Your task is to synthesize these partial analyses into a single, comprehensive final report. Your tone should be professional, insightful, and encouraging, aimed at helping the creator understand their audience and grow their channel.

	**Video Metadata:**
	This JSON object contains the video's overall statistics like view_count, like_count, and total comment_count. Use this for the final performance metrics, not the aggregated counts from chunks.
	%s

	**Partial Comment Analyses:**
	This is an array of JSON objects, where each object is a summary of a chunk of comments from the video.
	%s

	**Analysis Tasks & Final Output Structure:**
	Your entire output MUST be a single, minified JSON object, ready for ingestion into a BigQuery table. Your response must be raw JSON, starting with '{' and ending with '}'. Do NOT wrap the JSON in markdown code blocks. Crucially, all string values within the JSON must be properly escaped. For example, any double quotes (") inside a string must be escaped as \" and backslashes (\) must be escaped as \\. This is essential for creating valid JSON.

	1.  **Executive Summary ('executive_summary'):** Provide a concise, high-level overview (5-10 sentences) summarizing the video's overall performance, audience reception, and the most critical takeaway for the channel owner.

	2.  **Performance Metrics ('performance_metrics'):** Calculate and interpret key performance indicators.
	    * **'video_statistics'**: Use the raw counts for 'view_count', 'like_count', and 'comment_count' directly from the **Video Metadata**.
	    * **'engagement_ratios'**: Calculate the 'like_to_view_ratio' (likes/views) and 'comment_to_view_ratio' (comments/views) using the **Video Metadata**. Format as a decimal number.
	    * **'interpretation'**: Provide a 3-10 sentence qualitative analysis of these metrics (e.g., 'The video shows strong engagement with a high like-to-view ratio, suggesting the content resonated well with the core audience.').

	3.  **Audience Analysis ('audience_analysis'):** Analyze the sentiment and characteristics of the audience based on the partial analyses.
	    This MUST be an object containing the following fields:
	    * **'sentiment_label'**: A string with the overall sentiment ('Overwhelmingly Positive', 'Positive', 'Mixed', 'Negative', 'Overwhelmingly Negative').
	    * **'summary'**: A string (2-5 sentences) explaining the dominant sentiment and its drivers.
	    * **'positive_comments'**: An integer count, which is the SUM of all 'positive_comments' from the partial analyses.
	    * **'negative_comments'**: An integer count, which is the SUM of all 'negative_comments' from the partial analyses.
	    * **'neutral_comments'**: An integer count, which is the SUM of all 'neutral_comments' from the partial analyses.
	    * **'audience_persona'**: A string (2-5 sentences) describing the likely viewer persona.

	4.  **Content Feedback ('content_feedback')**: Synthesize direct feedback about the video's content from the comments.
	    * **'positive_feedback'**: An array of the top 5 most common points of positive feedback. Each object must have 'point' (a summary of the feedback) and 'representative_comment'.
	    * **'constructive_criticism'**: An array of the top 5 most common points of constructive criticism. Each object must have 'point' and 'representative_comment'.
	    * **'unanswered_questions'**: An array of the top 5 recurring questions from the audience. Each object must have 'question' and 'representative_comment'.

	5.  **Key Discussion Themes ('key_themes'):** Identify the top 10 dominant themes discussed in the comments. For each theme, provide:
	    * **'theme_title'**: A short, descriptive title (e.g., 'Brand Loyalty & Criticism').
	    * **'summary'**: A 2-3 sentence explanation of the theme.
	    * **'representative_comment'**: The text of one comment that best exemplifies this theme.

	6.  **Engagement Highlights ('engagement_highlights'):** From all partial analyses, identify the top 10 comments that generated the most engagement (likes + replies). For each comment, provide:
	    * **'comment_text'**: The full text of the comment.
	    * **'engagement_count'**: The sum of likes and replies for that comment.
	    * **'reason_for_engagement'**: A brief explanation of *why* the comment was engaging (e.g., 'Humorous observation,' 'Helpful technical tip,' 'Controversial opinion').

	7.  **SWOT Analysis ('swot_analysis'):** Perform a brief SWOT analysis based on the video and comments to identify strategic insights. Each section should be 3-5 sentences.
	    * **'strengths'**: What aspects are working well? (e.g., 'The main personality, 'Skylten', is very popular and drives positive community sentiment.').
	    * **'weaknesses'**: What are the identifiable shortcomings? (e.g., 'Some viewers react negatively to specific brands or on-screen actions, indicating a potential brand perception issue.').
	    * **'opportunities'**: What are potential areas for growth? (e.g., 'High engagement on technical tips suggests an opportunity for dedicated 'how-to' videos.').
	    * **'threats'**: What are the potential risks? (e.g., 'Technical inaccuracies mentioned in comments could damage channel credibility if not addressed.').

	8.  **Actionable Recommendations ('actionable_recommendations'):** Provide specific, data-driven recommendations for the channel creator.
	    * **'content_strategy'**: An array of objects. Each object MUST have two string fields: 'idea' (a concrete content idea) and 'reason' (a justification based on the analysis).
	    * **'video_improvements'**: An array of objects. Each object MUST have two string fields: 'suggestion' (e.g., "Improve audio quality") and 'reason' (e.g., "Multiple comments mentioned the background noise was distracting.").
	    * **'community_management'**: A single string providing a specific tip for community engagement.
	    * **'monetization_opportunities'**: An array of objects. Each object MUST have a 'category' (string, e.g., "Workwear Brands") and 'products' (an array of strings, e.g., ["BrandA", "BrandB"]).

	**Final Output Constraint:** The final output must only be the minified JSON object containing the analytical fields: 'executive_summary', 'performance_metrics', 'audience_analysis', 'content_feedback', 'key_themes', 'engagement_highlights', 'swot_analysis', and 'actionable_recommendations'. Do NOT include 'tracking_id' or 'run_date' in your output, as they are handled programmatically.
	`

func chunkComments(comments []*models.Comment, chunkSize int) [][]*models.Comment {
	var chunks [][]*models.Comment
	if len(comments) == 0 {
		return chunks
	}
	for i := 0; i < len(comments); i += chunkSize {
		end := i + chunkSize
		if end > len(comments) {
			end = len(comments)
		}
		chunks = append(chunks, comments[i:end])
	}
	return chunks
}

func cleanAndFinalizeAnalysis(rawResponse string, trackingID string, runDate string) ([]byte, error) {
	start := strings.Index(rawResponse, "{")
	end := strings.LastIndex(rawResponse, "}")
	if start == -1 || end == -1 || end < start {
		return nil, fmt.Errorf("could not find valid JSON object in response: %s", rawResponse)
	}
	cleanedJSONStr := rawResponse[start : end+1]

	// Use a decoder to be more robust against trailing characters.
	// This handles cases where the LLM returns a valid JSON object followed by a comma or other text.
	decoder := json.NewDecoder(strings.NewReader(cleanedJSONStr))
	decoder.DisallowUnknownFields() // Be strict about the struct fields
	var record models.AnalysisRecord
	if err := decoder.Decode(&record); err != nil {
		// If decoding fails, even with the decoder, the JSON is truly invalid.
		return nil, fmt.Errorf("cleaned JSON is not valid: %w. Raw JSON string: %s", err, cleanedJSONStr)
	}

	// The decoder.Decode call above successfully parsed a JSON object into `record`.
	// We can now ignore any trailing characters that might have caused the original error
	// and proceed with the valid data we have.
	record.TrackingID = trackingID
	record.RunDate = runDate

	return json.Marshal(record)
}

func AnalyzeData(cfg *models.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		ctx := r.Context()
		trackingID := r.URL.Query().Get("trackingId")
		if trackingID == "" {
			shared.LogJSON("WARNING", "Missing 'trackingId' query parameter", "")
			shared.JSONErrorResponse(w, "", http.StatusBadRequest, "Missing 'trackingId' query parameter")
			return
		}
		runDate := time.Now().Format("2006-01-02")
		shared.LogJSON("INFO", fmt.Sprintf("Received request: %s %s", r.Method, r.URL.String()), trackingID)

		objectName := fmt.Sprintf("%s.json", trackingID)
		shared.LogJSON("INFO", fmt.Sprintf("Reading from gs://%s/%s", cfg.GCSBucketName, objectName), trackingID)
		fileData, err := shared.GetFileFromGCS(ctx, cfg.GCSBucketName, objectName)
		if err != nil {
			shared.LogJSON("ERROR", fmt.Sprintf("could not get file from GCS (%s): %s", objectName, err), trackingID)
			shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to retrieve data file")
			return
		}

		var fullData models.VideoData
		if err := json.Unmarshal(fileData, &fullData); err != nil {
			shared.LogJSON("ERROR", "could not unmarshal JSON: "+err.Error(), trackingID)
			shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Invalid data format")
			return
		}
		shared.LogJSON("INFO", fmt.Sprintf("Successfully unmarshaled JSON data for video ID %s", fullData.ID), trackingID)

		client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GEMINIApiKey))
		if err != nil {
			shared.LogJSON("ERROR", "Failed to create Gemini client: "+err.Error(), trackingID)
			shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to create Gemini client")
			return
		}
		defer client.Close()

		model := client.GenerativeModel(cfg.GEMINIModel)
		model.SafetySettings = []*genai.SafetySetting{
			{
				Category:  genai.HarmCategoryHarassment,
				Threshold: genai.HarmBlockNone,
			},
			{
				Category:  genai.HarmCategoryHateSpeech,
				Threshold: genai.HarmBlockNone,
			},
			{
				Category:  genai.HarmCategorySexuallyExplicit,
				Threshold: genai.HarmBlockNone,
			},
			{
				Category:  genai.HarmCategoryDangerousContent,
				Threshold: genai.HarmBlockNone,
			},
		}

		shared.LogJSON("INFO", fmt.Sprintf("Sending request to Gemini model '%s' for analysis...", cfg.GEMINIModel), trackingID)

		const commentChunkSize = 100
		commentChunks := chunkComments(fullData.Comments, commentChunkSize)
		shared.LogJSON("INFO", fmt.Sprintf("Split %d comments into %d chunks of size %d.", len(fullData.Comments), len(commentChunks), commentChunkSize), trackingID)

		limiter := rate.NewLimiter(rate.Every(600*time.Millisecond), 1)

		var wg sync.WaitGroup
		analysisChunksChan := make(chan string, len(commentChunks))
		errChan := make(chan error, len(commentChunks))

		baseVideoData := fullData
		baseVideoData.Comments = nil

		for i, chunk := range commentChunks {
			wg.Add(1)

			go func(chunkIndex int, commentChunk []*models.Comment) {
				defer wg.Done()

				if err := limiter.Wait(ctx); err != nil {
					errChan <- fmt.Errorf("rate limiter wait error: %w", err)
					return
				}
				chunkDataForPrompt := baseVideoData
				chunkDataForPrompt.Comments = commentChunk
				chunkDataBytes, err := json.Marshal(chunkDataForPrompt)
				if err != nil {
					errChan <- fmt.Errorf("chunk %d marshal error: %w", chunkIndex, err)
					return
				}

				mapPromptFormatted := fmt.Sprintf(mapPrompt, string(chunkDataBytes))

				shared.LogJSON("INFO", fmt.Sprintf("Analyzing comment chunk %d/%d...", chunkIndex+1, len(commentChunks)), trackingID)
				resp, err := model.GenerateContent(ctx, genai.Text(mapPromptFormatted))
				if err != nil {
					errChan <- fmt.Errorf("chunk %d Gemini error: %w", chunkIndex, err)
					return
				}

				if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
					if analysis, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
						analysisChunksChan <- string(analysis)
					} else {
						errChan <- fmt.Errorf("chunk %d: Gemini response part is not text", chunkIndex)
					}
				} else {
					errChan <- fmt.Errorf("chunk %d: received empty response from Gemini", chunkIndex)
				}
			}(i, chunk)
		}

		wg.Wait()
		close(analysisChunksChan)
		close(errChan)

		if len(errChan) > 0 {
			for e := range errChan {
				shared.LogJSON("ERROR", "Error during chunk analysis: "+e.Error(), trackingID)
			}
			shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "One or more chunks failed analysis. See logs for details.")
			return
		}

		var analysisChunks []string
		for analysis := range analysisChunksChan {
			analysisChunks = append(analysisChunks, analysis)
		}
		combinedAnalyses := "[" + strings.Join(analysisChunks, ",") + "]"
		shared.LogJSON("INFO", "All chunks analyzed. Starting final reduction step.", trackingID)

		baseVideoDataBytes, err := json.Marshal(baseVideoData)
		if err != nil {
			shared.LogJSON("ERROR", "Failed to marshal base video data: "+err.Error(), trackingID)
			shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to prepare data for final analysis")
			return
		}
		reducePromptFormatted := fmt.Sprintf(reducePrompt, string(baseVideoDataBytes), combinedAnalyses)

		var finalJSON []byte
		const maxRetries = 3

		for attempt := 1; attempt <= maxRetries; attempt++ {
			shared.LogJSON("INFO", fmt.Sprintf("Attempt %d/%d: Generating final analysis from Gemini.", attempt, maxRetries), trackingID)
			geminiStartTime := time.Now()
			resp, err := model.GenerateContent(ctx, genai.Text(reducePromptFormatted))
			if err != nil {
				shared.LogJSON("WARNING", fmt.Sprintf("Attempt %d: Gemini API call failed: %s", attempt, err), trackingID)
				if attempt == maxRetries {
					shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, fmt.Sprintf("Failed to generate content from Gemini after %d retries", maxRetries))
					return
				}
				time.Sleep(2 * time.Second)
				continue
			}

			if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
				shared.LogJSON("WARNING", fmt.Sprintf("Attempt %d: Received empty response from Gemini", attempt), trackingID)
				if attempt == maxRetries {
					shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, fmt.Sprintf("Received empty response from Gemini after %d retries", maxRetries))
					return
				}
				continue
			}

			analysisPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
			if !ok {
				shared.LogJSON("WARNING", fmt.Sprintf("Attempt %d: Gemini response part is not text", attempt), trackingID)
				if attempt == maxRetries {
					shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, fmt.Sprintf("Invalid response format from Gemini after %d retries", maxRetries))
					return
				}
				continue
			}
			shared.LogJSON("INFO", fmt.Sprintf("Successfully received analysis from Gemini in %s.", time.Since(geminiStartTime)), trackingID)

			finalJSON, err = cleanAndFinalizeAnalysis(string(analysisPart), trackingID, runDate)
			if err == nil {
				shared.LogJSON("INFO", "Successfully parsed and validated Gemini response.", trackingID)
				break
			}

			shared.LogJSON("WARNING", fmt.Sprintf("Attempt %d: Failed to clean and validate Gemini response: %s. Raw response: %s", attempt, err, string(analysisPart)), trackingID)
			if attempt == maxRetries {
				shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, fmt.Sprintf("Failed to process analysis from Gemini after %d retries", maxRetries))
				return
			}
			time.Sleep(2 * time.Second) // Wait before retrying
		}

		analysisObjectName := fmt.Sprintf("%s_analyzed.json", trackingID)
		err = shared.UploadToGCS(ctx, cfg.GCSBucketName, analysisObjectName, finalJSON)
		if err != nil {
			shared.LogJSON("ERROR", "Failed to upload final analysis to GCS: "+err.Error(), trackingID)
			shared.JSONErrorResponse(w, trackingID, http.StatusInternalServerError, "Failed to upload analysis to GCS")
			return
		}
		shared.LogJSON("INFO", fmt.Sprintf("Successfully uploaded analysis to GCS object: gs://%s/%s", cfg.GCSBucketName, analysisObjectName), trackingID)

		nextActionURI := fmt.Sprintf("/ingest?trackingId=%s", trackingID)
		response := models.APIResponse{
			TrackingID:     trackingID,
			ProcessingTime: time.Since(startTime).String(),
			Status:         "success",
			Message:        fmt.Sprintf("Successfully analyzed data and uploaded result to %s", analysisObjectName),
			NextActionURI:  nextActionURI,
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
