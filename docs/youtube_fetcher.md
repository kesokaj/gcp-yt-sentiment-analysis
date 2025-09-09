# YouTube Data Fetcher

## Overview

This service endpoint is responsible for the first step in the analysis pipeline. It connects to the YouTube Data API v3 to fetch comprehensive data for a specific video, including its metadata, statistics, and comments.

The collected data is then consolidated into a single JSON file (`<trackingId>.json`) and uploaded to a Google Cloud Storage (GCS) bucket for subsequent processing by other services.

## How to Use

The service is triggered by making a GET request to the `/youtube` endpoint.

*   **Endpoint**: `GET /youtube`
*   **Query Parameters**:
    *   `videoId` (required): The unique ID of the YouTube video you want to analyze.
    *   `trackingId` (optional): A UUID to track this specific job through the pipeline. If one is not provided, it will be generated automatically.
*   **Example**:
    ```bash
    curl "http://localhost:8080/youtube?videoId=dQw4w9WgXcQ"
    ```

## Key Functions

*   **`yt_video.FetchData`**: This is the factory function that returns the main `http.HandlerFunc`. It orchestrates the entire process, from validating input parameters and fetching data from the YouTube API to uploading the final JSON result to GCS.

*   **`yt_video.getYouTubeService`**: This function initializes and returns a singleton instance of the YouTube service client. It uses `sync.Once` to ensure the client is created only on the first request, improving performance for subsequent calls.

*   **Comment Fetching Loop**: The handler fetches comments by ordering them by `relevance` to get the most impactful comments first. It iteratively calls the `CommentThreads.List` endpoint, but note that when ordering by relevance, the YouTube API may not return all available comments. The loop continues until the API stops providing more pages, the `MAX_COMMENTS_TO_FETCH` limit is reached, or a quota error occurs. This approach prioritizes comment quality over quantity for the analysis.
