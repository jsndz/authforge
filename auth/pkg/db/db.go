package db

import (
	"log"

	"github.com/jsndz/authforge/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DBConnectURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Coudn't run postgres")
	}
	return db, nil
}
