# UI Application for YouTube Analysis

## Overview

The YouTube Analysis Service includes an integrated web-based UI. It allows a user to initiate the entire three-step analysis pipeline (fetch, analyze, ingest) by simply providing a YouTube video URL. The interface displays the progress of each step in real-time.

## How to Use

### 1. Run the Main Application

Follow the instructions in the main `README.md` to build and run the service. No separate UI server is needed.

### 2. Access the UI

Once the main service is running (either locally or on Cloud Run), open your web browser and navigate to the `/ui` endpoint.

- **Locally**: `http://localhost:8080/ui`
- **On Cloud Run**: `https://your-backend-service-url.run.app/ui`

-   Paste a full YouTube video URL into the input field.
-   Click the "Run" button.
-   The "Live Status" section will display real-time updates as the backend processes each step.

## Key Functions

The UI functionality is handled by the `pkgs/ui_handler` package within the main service.

-   **`ProcessHandler(...)`**: The core handler that manages the Server-Sent Events (SSE) connection. It receives the YouTube URL, extracts the video ID, and orchestrates the calls to the other service endpoints (`/youtube`, `/magic`, `/ingest`) in sequence, streaming status updates back to the client.
-   **`ServeUI(...)`**: Serves the `web/ui.html` file.