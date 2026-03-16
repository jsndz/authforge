package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID uint `gorm:"primaryKey"`

	UserName string `gorm:"size:50;uniqueIndex;not null"`
	Email    string `gorm:"size:100;uniqueIndex;not null"`
	Password string `gorm:"size:255;not null"`

	IsActive      bool `gorm:"default:true"`
	EmailVerified bool `gorm:"default:false"`

	LastLoginAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
