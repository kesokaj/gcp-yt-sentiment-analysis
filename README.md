# YouTube Sentiment Analysis

This service provides a multi-step pipeline to fetch YouTube video data, analyze it with AI, and ingest the results into BigQuery.

## Google Cloud Deployment Guide

This guide provides a streamlined approach to deploying the service on Google Cloud Run.

### Step 1: Prerequisites & Initial Setup

1.  **Install Tools**: Make sure you have the `gcloud` CLI installed and authenticated.

2.  **Set Environment Variables**: Open your terminal and set the following variables. **You must change `your-unique-bucket-name`**.

    ```bash
    export GCP_PROJECT="your-unique-project"
    export GCP_LOCATION="us-central1"       # Or your preferred region
    export AR_REPO_NAME="yt-sentiment-repo"
    export SERVICE_NAME="yt-sentiment"
    export GCS_BUCKET_NAME="your-unique-yt-sentiment-bucket" # <-- CHANGE THIS
    export BQ_DATASET="yt_sentiment_data"
    export SERVICE_ACCOUNT_NAME="yt-sentiment-sa"
    ```

3.  **Enable APIs**: Enable all necessary Google Cloud services.

    ```bash
    gcloud services enable \
        run.googleapis.com \
        cloudbuild.googleapis.com \
        artifactregistry.googleapis.com \
        iam.googleapis.com \
        bigquery.googleapis.com \
        storage-component.googleapis.com \
        aiplatform.googleapis.com \
        iap.googleapis.com --project=${GCP_PROJECT}
    ```

4.  **Set default project**

    ```bash
    gcloud config set project ${GCP_PROJECT}
    ```

### Step 2: Build and Push Docker Container

1.  **Create an Artifact Registry Repository**: This is where your container image will be stored.

    ```bash
    gcloud artifacts repositories create $AR_REPO_NAME \
        --repository-format=docker \
        --location=$GCP_LOCATION --project=${GCP_PROJECT}
    ```

2.  **Build the Container**: Use Cloud Build to build the container from your source code and push it to Artifact Registry.

    ```bash
    gcloud builds submit --tag ${GCP_LOCATION}-docker.pkg.dev/${GCP_PROJECT}/${AR_REPO_NAME}/${SERVICE_NAME}:latest
    ```

### Step 3: Create GCS and BigQuery Resources

1.  **Create GCS Bucket**: This bucket will store the intermediate JSON files.

    ```bash
    gsutil mb -l $GCP_LOCATION gs://$GCS_BUCKET_NAME
    ```

2.  **Create BigQuery Dataset and Tables**: This one-liner creates the dataset and then uses your `schemas.sql` file to create the necessary tables.

    ```bash
    bq mk --location=$GCP_LOCATION $BQ_DATASET && sed "s/your_dataset_name/$BQ_DATASET/g" schemas.sql | bq query --use_legacy_sql=false
    ```

### Step 4: Create Service Account and Deploy

1.  **Create Service Account**: Create a dedicated identity for your Cloud Run service.

    ```bash
    gcloud iam service-accounts create $SERVICE_ACCOUNT_NAME \
        --display-name="YouTube Sentiment Analysis Service Account"
    ```

2.  **Grant Permissions**: Grant the service account permissions to access GCS, BigQuery, Vertex AI (for Gemini), and Secret Manager.

    ```bash
    # GCS: To read/write JSON files
    gsutil iam ch serviceAccount:${SERVICE_ACCOUNT_NAME}@${GCP_PROJECT}.iam.gserviceaccount.com:objectAdmin gs://${GCS_BUCKET_NAME}

    # BigQuery: To write data and run ingestion jobs
    gcloud projects add-iam-policy-binding $GCP_PROJECT --member="serviceAccount:${SERVICE_ACCOUNT_NAME}@${GCP_PROJECT}.iam.gserviceaccount.com" --role="roles/bigquery.dataEditor"
    gcloud projects add-iam-policy-binding $GCP_PROJECT --member="serviceAccount:${SERVICE_ACCOUNT_NAME}@${GCP_PROJECT}.iam.gserviceaccount.com" --role="roles/bigquery.jobUser"

    # Vertex AI: To call the Gemini model
    gcloud projects add-iam-policy-binding $GCP_PROJECT --member="serviceAccount:${SERVICE_ACCOUNT_NAME}@${GCP_PROJECT}.iam.gserviceaccount.com" --role="roles/aiplatform.user"
    ```

3.  **Create and Set API Keys**:
    > **Note**: API Key creation for these services is a manual UI-based process. There are no `gcloud` commands to generate them directly.

    *   **YouTube API Key**:
        1.  Go to the Google Cloud Console Credentials page.
        2.  Click **+ CREATE CREDENTIALS** and select **API key**.
        3.  Copy the generated key. It's highly recommended to restrict the key to the "YouTube Data API v3".

    *   **Gemini API Key**:
        1.  Go to Google AI Studio.
        2.  Click **Create API key in new project**.
        3.  Copy the generated key.

        ```bash
        Gemini for Google Cloud API
        YouTube Data API v3
        Generative Language API
        ```

    *   **Set Environment Variables**: Once you have your keys, set them as environment variables in your terminal.
        > **Security Note**: For production workloads, it is highly recommended to use Secret Manager instead of passing keys as plain environment variables.

    ```bash
    export YOUTUBE_API_KEY="your-youtube-api-key" # <-- PASTE YOUR KEY HERE
    export GEMINI_API_KEY="your-gemini-api-key"   # <-- PASTE YOUR KEY HERE
    ```

4.  **Deploy to Cloud Run**: Deploy the container image, connecting it to the service account and all environment variables.

    ```bash
    gcloud run deploy $SERVICE_NAME \
        --image ${GCP_LOCATION}-docker.pkg.dev/${GCP_PROJECT}/${AR_REPO_NAME}/${SERVICE_NAME}:latest \
        --platform managed \
        --region $GCP_LOCATION \
        --allow-unauthenticated \
        --execution-environment=gen2 \
        --service-account ${SERVICE_ACCOUNT_NAME}@${GCP_PROJECT}.iam.gserviceaccount.com \
        --set-env-vars="YOUTUBE_API_KEY=$YOUTUBE_API_KEY,GEMINI_API_KEY=$GEMINI_API_KEY,GCS_BUCKET_NAME=$GCS_BUCKET_NAME,GCP_PROJECT=$GCP_PROJECT,GCP_LOCATION=$GCP_LOCATION,BQ_DATASET=$BQ_DATASET,GEMINI_MODEL=gemini-2.5-pro,MAX_COMMENTS_TO_FETCH=5000" \
        --cpu=1 \
        --memory=1Gi \
        --concurrency=80 \
        --timeout=3600
    ```

### Step 5: Access the UI

Once deployed, you can access the web interface to easily run the analysis pipeline.

1.  Find your service URL in the Cloud Run console or from the output of the deploy command.
2.  Navigate to `https://<your-service-url>/ui` in your web browser.

### Step 6: Visualize in Looker Studio (Optional)

After ingesting data, you can build a dashboard to visualize the AI-driven analysis.

1.  Go to Looker Studio and create a new **Blank Report**.
2.  When prompted to add data, select the **BigQuery** connector.
3.  Navigate to **My Projects** > `[Your Project]` > `yt_sentiment_data` > `analyzed` and click **Add**.
4.  You can now build charts using the fields from the `analyzed` table. For example:
    *   A **Pie Chart** with `audienceAnalysis.sentimentLabel` as the dimension.
    *   A **Table** showing `keyThemes.themeTitle` and `keyThemes.summary`.
    *   **Scorecards** for `performanceMetrics.videoStatistics.viewCount`.