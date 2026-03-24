package domain

import (
	"encoding/json"
	"time"
)

type Episode struct {
	ID            int64           `json:"id,string" gorm:"primaryKey"`
	SeasonID      int64           `json:"season_id,string" gorm:"not null"`
	Title         string          `json:"title" gorm:"size:255;not null"`
	ScriptContent *string         `json:"script_content,omitempty"`
	Config        json.RawMessage `json:"config" gorm:"type:jsonb;not null;default:{}"`
	SortOrder     int             `json:"sort_order" gorm:"not null;default:0"`
	CreatedBy     int64           `json:"created_by,string" gorm:"not null"`
	CreatedAt     time.Time       `json:"created_at" gorm:"not null"`
	UpdatedAt     time.Time       `json:"updated_at" gorm:"not null"`
}

func (Episode) TableName() string {
	return "episodes"
}
