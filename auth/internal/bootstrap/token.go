package bootstrap

import (
	"github.com/jsndz/authforge/internal/handler"
	"github.com/jsndz/authforge/internal/repository"
	"github.com/jsndz/authforge/internal/services"
	"gorm.io/gorm"
)

func InitTokenModule(db *gorm.DB) *handler.TokenHandler {
	tokenRepo := repository.NewTokenRepository(db)
	tokenService := services.NewTokenService(tokenRepo)
	return handler.NewTokenHandler(tokenService)
}
