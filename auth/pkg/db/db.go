package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(URL string) (*gorm.DB, error) {

	db, err := gorm.Open(postgres.Open(URL), &gorm.Config{})
	if err != nil {
		log.Fatal("Coudn't run postgres")
	}
	return db, nil
}
