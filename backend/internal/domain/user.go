package domain

import (
	"time"
)

type User struct {
	ID           int64      `json:"id,string" gorm:"primaryKey"`
	Username     string     `json:"username" gorm:"size:64;not null"`
	DisplayName  string     `json:"display_name" gorm:"size:128;not null"`
	PasswordHash string     `json:"-" gorm:"size:255;not null"`
	Role         string     `json:"role" gorm:"size:16;not null;default:user"`
	Enabled      bool       `json:"enabled" gorm:"not null;default:true"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"not null"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (User) TableName() string {
	return "users"
}

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)
