# BigQuery Ingestor

**Package:** `pkgs/bq_ingest`
**File:** `ingestor.go`

This service is responsible for ingesting both raw and analyzed YouTube data into Google BigQuery.

## Functions

### `IngestData(cfg *models.AppConfig) http.HandlerFunc`

This is the main HTTP handler for the service. It reads data from Google Cloud Storage (GCS) and ingests it into the appropriate BigQuery tables.

**Endpoint:** `/ingest`

**Query Parameters:**

*   `trackingId` (required): The unique identifier for the analysis job.

**Logic:**

1.  **Check for Existing Data**: It first checks if data for the given `trackingId` already exists in the `videos` and `analyzed` tables to prevent duplicates.
2.  **Fetch from GCS**: If the data is new, it fetches the corresponding raw data (`<trackingId>.json`) and analyzed data (`<trackingId>_analyzed.json`) from the GCS bucket.
3.  **Ingest Raw Data**: It ingests the video metadata into the `videos` table and the comments into the `comments` table.
4.  **Ingest Analyzed Data**: It ingests the Gemini analysis report into the `analyzed` table.

## Usage

```bash
curl "http://localhost:8080/ingest?trackingId=<your-tracking-id>"
```

## Error Handling

The `IngestData` function is designed to be robust. It wraps errors with context using `fmt.Errorf` before logging them, which provides a detailed trace in the logs. Common errors include:

*   Failure to connect to BigQuery.
*   Failure to retrieve data files from GCS.
*   Errors unmarshaling JSON data.
*   Failures during the data insertion into BigQuery tables.