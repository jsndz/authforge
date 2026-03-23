package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBConnectURL string
	JWTSecret    string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	return &Config{
		JWTSecret:    os.Getenv("JWT_SECRET"),
		DBConnectURL: os.Getenv("DB_CONNECT_URL"),
	}
}
