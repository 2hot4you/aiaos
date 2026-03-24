package domain

import "time"

type Season struct {
	ID        int64     `json:"id,string" gorm:"primaryKey"`
	ProjectID int64     `json:"project_id,string" gorm:"not null"`
	Title     string    `json:"title" gorm:"size:255;not null"`
	SortOrder int       `json:"sort_order" gorm:"not null;default:0"`
	CreatedBy int64     `json:"created_by,string" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`

	Episodes []Episode `json:"episodes,omitempty" gorm:"foreignKey:SeasonID"`
}

func (Season) TableName() string {
	return "seasons"
}
