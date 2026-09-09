package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint64    `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"not null;uniqueIndex"`
	Name      string    `json:"name" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type userRepo struct {
	db *gorm.DB
}
