CREATE TABLE your_dataset_name.videos (
id STRING,
channel_id STRING,
channel_title STRING,
tracking_id STRING,
run_date DATE,
title STRING,
description STRING,
thumbnail_url STRING,
duration STRING,
category_id STRING,
view_count INTEGER,
like_count INTEGER,
favorite_count INTEGER,
comment_count INTEGER
);

CREATE TABLE your_dataset_name.comments (
video_id STRING,
id STRING,
parent_id STRING,
channel_id STRING,
text STRING,
like_count INTEGER,
reply_count INTEGER,
tracking_id STRING,
run_date DATE
);

CREATE TABLE your_dataset_name.analyzed (
    tracking_id STRING,
    run_date DATE,
    executive_summary STRING,
    performance_metrics STRUCT<
        video_statistics STRUCT<
            view_count INT64,
            like_count INT64,
            comment_count INT64
        >,
        engagement_ratios STRUCT<
            like_to_view_ratio FLOAT64,
            comment_to_view_ratio FLOAT64
        >,
        interpretation STRING
    >,
    audience_analysis STRUCT<
        sentiment_label STRING,
        summary STRING,
        positive_comments INT64,
        negative_comments INT64,
        neutral_comments INT64,
        audience_persona STRING
    >,
    content_feedback STRUCT<
        positive_feedback ARRAY<STRUCT<point STRING, representative_comment STRING>>,
        constructive_criticism ARRAY<STRUCT<point STRING, representative_comment STRING>>,
        unanswered_questions ARRAY<STRUCT<question STRING, representative_comment STRING>>
    >,
    key_themes ARRAY<STRUCT<
        theme_title STRING,
        summary STRING,
        representative_comment STRING
    >>,
    engagement_highlights ARRAY<STRUCT<
        comment_text STRING,
        engagement_count INT64,
        reason_for_engagement STRING
    >>,
    swot_analysis STRUCT<
        strengths STRING,
        weaknesses STRING,
        opportunities STRING,
        threats STRING
    >,
    actionable_recommendations STRUCT<
        content_strategy ARRAY<STRUCT<idea STRING, reason STRING>>,
        video_improvements ARRAY<STRUCT<suggestion STRING, reason STRING>>,
        community_management STRING,
        monetization_opportunities ARRAY<STRUCT<category STRING, products ARRAY<STRING>>>
    >
);
