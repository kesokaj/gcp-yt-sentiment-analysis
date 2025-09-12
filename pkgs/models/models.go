package models

type APIResponse struct {
	TrackingID     string `json:"tracking_id"`
	ProcessingTime string `json:"processing_time"`
	Status         string `json:"status"`
	Message        string `json:"message,omitempty"`
	NextActionURI  string `json:"next_action_uri,omitempty"`
}

type AppConfig struct {
	YTApiKey           string
	GEMINIApiKey       string
	GCSBucketName      string
	GCPProject         string
	GCPLocation        string
	BQDataset          string
	GEMINIModel        string
	Port               string
	MaxCommentsToFetch int
}

type VideoData struct {
	ID            string     `json:"id"`
	ChannelID     string     `json:"channel_id"`
	ChannelTitle  string     `json:"channel_title"`
	TrackingID    string     `json:"tracking_id"`
	RunDate       string     `json:"run_date"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	ThumbnailURL  string     `json:"thumbnail_url"`
	Duration      string     `json:"duration"`
	CategoryID    string     `json:"category_id"`
	ViewCount     int64      `json:"view_count"`
	LikeCount     int64      `json:"like_count"`
	FavoriteCount int64      `json:"favorite_count"`
	CommentCount  int64      `json:"comment_count"`
	Comments      []*Comment `json:"comments"`
}

type Comment struct {
	VideoID    string `json:"-" bigquery:"video_id"`
	ID         string `json:"id" bigquery:"id"`
	ParentID   string `json:"parent_id,omitempty" bigquery:"parent_id"`
	ChannelID  string `json:"channel_id" bigquery:"channel_id"`
	Text       string `json:"text" bigquery:"text"`
	LikeCount  int64  `json:"like_count" bigquery:"like_count"`
	ReplyCount int64  `json:"reply_count" bigquery:"reply_count"`
	TrackingID string `json:"tracking_id" bigquery:"tracking_id"`
	RunDate    string `json:"run_date" bigquery:"run_date"`
}

type VideoRecord struct {
	ID            string `bigquery:"id"`
	ChannelID     string `bigquery:"channel_id"`
	ChannelTitle  string `bigquery:"channel_title"`
	TrackingID    string `bigquery:"tracking_id"`
	RunDate       string `bigquery:"run_date"`
	Title         string `bigquery:"title"`
	Description   string `bigquery:"description"`
	ThumbnailURL  string `bigquery:"thumbnail_url"`
	Duration      string `bigquery:"duration"`
	CategoryID    string `bigquery:"category_id"`
	ViewCount     int64  `bigquery:"view_count"`
	LikeCount     int64  `bigquery:"like_count"`
	FavoriteCount int64  `bigquery:"favorite_count"`
	CommentCount  int64  `bigquery:"comment_count"`
}

type AnalysisRecord struct {
	TrackingID                string                    `json:"tracking_id" bigquery:"tracking_id"`
	RunDate                   string                    `json:"run_date" bigquery:"run_date"`
	ExecutiveSummary          string                    `json:"executive_summary" bigquery:"executive_summary"`
	PerformanceMetrics        PerformanceMetrics        `json:"performance_metrics" bigquery:"performance_metrics"`
	AudienceAnalysis          AudienceAnalysis          `json:"audience_analysis" bigquery:"audience_analysis"`
	ContentFeedback           ContentFeedback           `json:"content_feedback" bigquery:"content_feedback"`
	KeyThemes                 []KeyTheme                `json:"key_themes" bigquery:"key_themes"`
	EngagementHighlights      []EngagementHighlight     `json:"engagement_highlights" bigquery:"engagement_highlights"`
	SWOTAnalysis              SWOTAnalysis              `json:"swot_analysis" bigquery:"swot_analysis"`
	ActionableRecommendations ActionableRecommendations `json:"actionable_recommendations" bigquery:"actionable_recommendations"`
}

type PerformanceMetrics struct {
	VideoStatistics  VideoStatistics  `json:"video_statistics" bigquery:"video_statistics"`
	EngagementRatios EngagementRatios `json:"engagement_ratios" bigquery:"engagement_ratios"`
	Interpretation   string           `json:"interpretation" bigquery:"interpretation"`
}

type VideoStatistics struct {
	ViewCount    int64 `json:"view_count" bigquery:"view_count"`
	LikeCount    int64 `json:"like_count" bigquery:"like_count"`
	CommentCount int64 `json:"comment_count" bigquery:"comment_count"`
}

type EngagementRatios struct {
	LikeToViewRatio    float64 `json:"like_to_view_ratio" bigquery:"like_to_view_ratio"`
	CommentToViewRatio float64 `json:"comment_to_view_ratio" bigquery:"comment_to_view_ratio"`
}

type AudienceAnalysis struct {
	SentimentLabel   string `json:"sentiment_label" bigquery:"sentiment_label"`
	Summary          string `json:"summary" bigquery:"summary"`
	PositiveComments int64  `json:"positive_comments" bigquery:"positive_comments"`
	NegativeComments int64  `json:"negative_comments" bigquery:"negative_comments"`
	NeutralComments  int64  `json:"neutral_comments" bigquery:"neutral_comments"`
	AudiencePersona  string `json:"audience_persona" bigquery:"audience_persona"`
}

type ContentFeedback struct {
	PositiveFeedback      []FeedbackPoint `json:"positive_feedback" bigquery:"positive_feedback"`
	ConstructiveCriticism []FeedbackPoint `json:"constructive_criticism" bigquery:"constructive_criticism"`
	UnansweredQuestions   []QuestionPoint `json:"unanswered_questions" bigquery:"unanswered_questions"`
}

type FeedbackPoint struct {
	Point                 string `json:"point" bigquery:"point"`
	RepresentativeComment string `json:"representative_comment" bigquery:"representative_comment"`
}

type KeyTheme struct {
	ThemeTitle            string `json:"theme_title" bigquery:"theme_title"`
	Summary               string `json:"summary" bigquery:"summary"`
	RepresentativeComment string `json:"representative_comment" bigquery:"representative_comment"`
}

type EngagementHighlight struct {
	CommentText         string `json:"comment_text" bigquery:"comment_text"`
	EngagementCount     int64  `json:"engagement_count" bigquery:"engagement_count"`
	ReasonForEngagement string `json:"reason_for_engagement" bigquery:"reason_for_engagement"`
}

type SWOTAnalysis struct {
	Strengths     string `json:"strengths" bigquery:"strengths"`
	Weaknesses    string `json:"weaknesses" bigquery:"weaknesses"`
	Opportunities string `json:"opportunities" bigquery:"opportunities"`
	Threats       string `json:"threats" bigquery:"threats"`
}

type ContentStrategyRecommendation struct {
	Idea   string `json:"idea" bigquery:"idea"`
	Reason string `json:"reason" bigquery:"reason"`
}

type VideoImprovementRecommendation struct {
	Suggestion string `json:"suggestion" bigquery:"suggestion"`
	Reason     string `json:"reason" bigquery:"reason"`
}

type QuestionPoint struct {
	Question              string `json:"question" bigquery:"question"`
	RepresentativeComment string `json:"representative_comment" bigquery:"representative_comment"`
}

type MonetizationOpportunity struct {
	Category string   `json:"category" bigquery:"category"`
	Products []string `json:"products" bigquery:"products"`
}

type ActionableRecommendations struct {
	ContentStrategy           []ContentStrategyRecommendation  `json:"content_strategy" bigquery:"content_strategy"`
	VideoImprovements         []VideoImprovementRecommendation `json:"video_improvements" bigquery:"video_improvements"`
	CommunityManagement       string                           `json:"community_management" bigquery:"community_management"`
	MonetizationOpportunities []MonetizationOpportunity        `json:"monetization_opportunities" bigquery:"monetization_opportunities"`
}
