package db

import (
	"github.com/jsndz/authforge/internal/model"
	"gorm.io/gorm"
)

func MigrateDB(db *gorm.DB) error {
	return db.AutoMigrate(&model.User{}, &model.Token{})
}
