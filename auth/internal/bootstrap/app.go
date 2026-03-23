package bootstrap

import (
	"github.com/jsndz/authforge/internal/handler"
	"github.com/jsndz/authforge/internal/repository"
	"github.com/jsndz/authforge/internal/services"
	"github.com/jsndz/authforge/pkg/email"
	"gorm.io/gorm"
)

type AppContainer struct {
	UserHandler  *handler.UserHandler
	TokenHandler *handler.TokenHandler
}

func InitApp(db *gorm.DB, jwtSecret string) *AppContainer {
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)

	tokenService := services.NewTokenService(tokenRepo)
	sessionService := services.NewSessionService(jwtSecret)
	emailService := email.NewEmailService()
	userService := services.NewUserService(userRepo, tokenService, sessionService, emailService)

	userHandler := handler.NewUserHandler(userService)
	tokenHandler := handler.NewTokenHandler(tokenService)

	return &AppContainer{
		UserHandler:  userHandler,
		TokenHandler: tokenHandler,
	}
}
