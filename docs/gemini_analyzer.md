# Gemini Analyzer

**Package:** `pkgs/gemini_magic`
**File:** `analyzer.go`

This service uses the Google Gemini AI to perform a deep analysis of YouTube video data.

## Functions

### `AnalyzeData(cfg *models.AppConfig) http.HandlerFunc`

This HTTP handler takes a `trackingId`, retrieves the corresponding raw data from GCS, sends it to the Gemini API for analysis, and stores the result back in GCS.

**Endpoint:** `/magic`

**Query Parameters:**

*   `trackingId` (required): The unique identifier for the analysis job.

**Logic:**

1.  **Fetch from GCS**: Retrieves the raw data file (`<trackingId>.json`) from the GCS bucket.
2.  **Chunking**: Splits the comments into smaller chunks to fit within the token limits of the Gemini API.
3.  **Map-Reduce Analysis**: Performs a map-reduce style analysis:
    *   **Map**: Each chunk is sent to the Gemini API in parallel for a partial analysis.
    *   **Reduce**: The partial analyses are combined and sent to the Gemini API in a final call to generate a comprehensive report.
4.  **Store in GCS**: The final analysis is saved as a new JSON file (`<trackingId>_analyzed.json`) in the GCS bucket.

## Prompts

The analysis is guided by two main prompts defined as constants in the code:

*   `mapPrompt`: Instructs the AI to perform a partial analysis on a chunk of comments.
*   `reducePrompt`: Instructs the AI to synthesize the partial analyses into a final, comprehensive report.

For easier maintenance, these prompts could be externalized into separate `.txt` or `.md` files and read by the application at runtime.

## Usage

```bash
curl "http://localhost:8080/magic?trackingId=<your-tracking-id>"
```

## Error Handling

Errors are wrapped with context using `fmt.Errorf` for detailed logging. The function includes a retry mechanism for the Gemini API calls to handle transient network issues. Common errors include:

*   Failure to create the Gemini client.
*   Errors during the Gemini API calls.
*   Failure to parse the JSON response from the AI.
*   GCS access errors.