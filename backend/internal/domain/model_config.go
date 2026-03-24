package domain

import "time"

type AIModelConfig struct {
	ID              int64     `json:"id,string" gorm:"primaryKey"`
	Name            string    `json:"name" gorm:"size:255;not null"`
	ModelType       string    `json:"model_type" gorm:"size:32;not null"`
	Provider        string    `json:"provider" gorm:"size:64;not null"`
	Endpoint        string    `json:"endpoint" gorm:"type:text;not null"`
	APIKeyEnc       string    `json:"-" gorm:"column:api_key_enc;type:text;not null"`
	ModelIdentifier string    `json:"model_identifier" gorm:"size:128;not null"`
	MaxConcurrency  int       `json:"max_concurrency" gorm:"not null;default:5"`
	TimeoutSeconds  int       `json:"timeout_seconds" gorm:"not null;default:60"`
	IsDefault       bool      `json:"is_default" gorm:"not null;default:false"`
	Enabled         bool      `json:"enabled" gorm:"not null;default:true"`
	CreatedAt       time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"not null"`
}

func (AIModelConfig) TableName() string {
	return "ai_model_configs"
}

const (
	ModelTypeText  = "text"
	ModelTypeImage = "image"
	ModelTypeVideo = "video"
)
