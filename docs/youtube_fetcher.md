# YouTube Fetcher

**Package:** `pkgs/yt_video`
**File:** `fetcher.go`

This service is responsible for fetching video and comment data from the YouTube Data API v3.

## Functions

### `FetchData(cfg *models.AppConfig) http.HandlerFunc`

This is the main HTTP handler for the service. It takes a YouTube video ID, fetches the relevant data, and stores it in a JSON file in Google Cloud Storage (GCS).

**Endpoint:** `/youtube`

**Query Parameters:**

*   `videoId` (required): The ID of the YouTube video.
*   `trackingId` (optional): A unique identifier for the job. If not provided, a new UUID will be generated.

**Logic:**

1.  **Fetch Video Details**: Retrieves video metadata, including statistics and content details.
2.  **Fetch Comments**: Fetches the most relevant comments for the video, up to the limit defined by the `MAX_COMMENTS_TO_FETCH` environment variable.
3.  **Store in GCS**: Saves the combined video and comment data as a JSON file (`<trackingId>.json`) in the specified GCS bucket.

## Usage

```bash
curl "http://localhost:8080/youtube?videoId=<your-video-id>"
```

## Error Handling

Errors are wrapped with context using `fmt.Errorf` for detailed logging. The function handles common issues such as:

*   Failure to create the YouTube service client.
*   Errors calling the YouTube API, including quota exceeded errors.
*   Failure to marshal the data to JSON.
*   Errors uploading the data file to GCS.