package ai

import "context"

// Provider is the unified interface for all AI model calls (mock for Phase 1)
type Provider interface {
	GenerateText(ctx context.Context, req *TextRequest) (*TextResponse, error)
	GenerateImage(ctx context.Context, req *ImageRequest) (*ImageResponse, error)
	GenerateVideo(ctx context.Context, req *VideoRequest) (*VideoResponse, error)
	HealthCheck(ctx context.Context) error
}

type TextRequest struct {
	Model        string `json:"model"`
	SystemPrompt string `json:"system_prompt"`
	UserPrompt   string `json:"user_prompt"`
	MaxTokens    int    `json:"max_tokens"`
}

type TextResponse struct {
	Content    string `json:"content"`
	TokensUsed int    `json:"tokens_used"`
}

type ImageRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type ImageResponse struct {
	ImageURL string `json:"image_url"`
}

type VideoRequest struct {
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	Duration int    `json:"duration"`
}

type VideoResponse struct {
	VideoURL string `json:"video_url"`
}
