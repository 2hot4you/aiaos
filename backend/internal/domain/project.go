package domain

import "time"

type Project struct {
	ID        int64     `json:"id,string" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:255;not null"`
	CreatedBy int64     `json:"created_by,string" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}

func (Project) TableName() string {
	return "projects"
}
