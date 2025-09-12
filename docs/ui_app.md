# UI Application

**Package:** `pkgs/ui_handler`
**File:** `handler.go`

This package provides the web interface for the application.

## Handlers

### `Info(w http.ResponseWriter, r *http.Request)`

Serves a plain text page with information about the available API endpoints.

**Endpoint:** `/`

### `ServeUI(w http.ResponseWriter, r *http.Request)`

Serves the main HTML page for the user interface from `web/ui.html`.

**Endpoint:** `/ui`

### `ProcessHandler(w http.ResponseWriter, r *http.Request)`

Orchestrates the multi-step analysis pipeline and streams progress back to the client using Server-Sent Events (SSE).

**Endpoint:** `/ui/process`

**Query Parameters:**

*   `url` (required): The full URL of the YouTube video to be analyzed.

**Logic:**

1.  Extracts the video ID from the YouTube URL.
2.  Calls the `/youtube` endpoint to fetch the video data.
3.  Calls the `/magic` endpoint to perform the AI analysis.
4.  Calls the `/ingest` endpoint to ingest the data into BigQuery.
5.  Streams status updates for each step to the web UI.
