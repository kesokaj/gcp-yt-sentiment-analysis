# Go YouTube Sentiment Analysis AI Developer

This document provides the necessary context for an AI assistant to contribute to this project effectively.

## 🚀 Project Overview

This project is a Go-based application that analyzes the sentiment of YouTube video comments. It fetches video and comment data from the YouTube API, uses Google's Gemini AI to analyze the sentiment of the comments, and stores the results in a BigQuery database. The application also provides a simple web interface to view the analysis results.

## 🏗️ Architecture

The application is designed with a package-by-domain structure, separating concerns into distinct packages. The data flows as follows:

1.  **YouTube Fetcher**: Fetches video and comment data from the YouTube API.
2.  **Gemini Analyzer**: Analyzes the sentiment of the comments using the Gemini AI.
3.  **BigQuery Ingestor**: Ingests the raw data and the analysis results into BigQuery.
4.  **UI Handler**: Provides a web interface to view the analysis results.

## 📦 Package Breakdown

*   `main.go`: The entry point of the application. It initializes the services, sets up logging, and starts the server.
*   `/pkgs/yt_video/fetcher.go`: Fetches video and comment data from the YouTube API.
*   `/pkgs/gemini_magic/analyzer.go`: Analyzes the sentiment of the comments using the Gemini AI.
*   `/pkgs/bq_ingest/ingestor.go`: Ingests the raw data and the analysis results into BigQuery.
*   `/pkgs/ui_handler/handler.go`: Provides a web interface to view the analysis results.
*   `/pkgs/models/models.go`: Contains the struct definitions for the data models used in the application.
*   `/pkgs/shared/`: Contains shared and reusable functions, such as configuration management, logging, and HTTP responses.
*   `/web/`: Contains the HTML, CSS, and client-side JavaScript for the web interface.
*   `/docs/`: Contains Markdown documentation for the project.
*   `/examples/`: Contains example files for reference.

## ⚙️ Configuration

The application is configured using environment variables. The following environment variables are required:

*   `GOOGLE_CLOUD_PROJECT`: The Google Cloud project ID.
*   `YOUTUBE_API_KEY`: The API key for the YouTube Data API.
*   `GEMINI_API_KEY`: The API key for the Gemini API.
*   `BIGQUERY_DATASET`: The name of the BigQuery dataset.

## 🏃‍♀️ Running the Application

To run the application, set the required environment variables and run the following command:

```bash
go run main.go
```

## 🗃️ Database Schema

The application uses BigQuery to store the data. The database schema is defined in the `schemas.sql` file and consists of the following tables:

*   `videos`: Stores the video data.
*   `comments`: Stores the comment data.
*   `analyzed`: Stores the sentiment analysis results.

For detailed information about the table schemas, please refer to the `schemas.sql` file.

## 🌐 API Endpoints

The application exposes the following API endpoints:

*   `GET /`: Serves the web interface.
*   `POST /analyze`: Triggers the sentiment analysis for a given YouTube video URL.

## 🖥️ Web UI

The web interface is a simple HTML page that allows users to enter a YouTube video URL and view the sentiment analysis results. The UI is served by the `ui_handler` package and the HTML file is located in the `/web/` directory.