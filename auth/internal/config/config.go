package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBConnectURL         string
	JWTSecret            string
	GOOGLE_CLIENT_ID     string
	GOOGLE_CLIENT_SECRET string
	GOOGLE_CALLBACK_URL  string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	return &Config{
		JWTSecret:            os.Getenv("JWT_SECRET"),
		DBConnectURL:         os.Getenv("DB_CONNECT_URL"),
		GOOGLE_CLIENT_ID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GOOGLE_CLIENT_SECRET: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GOOGLE_CALLBACK_URL:  os.Getenv("GOOGLE_CALLBACK_URL"),
	}
}
