package bootstrap

import (
	"github.com/jsndz/authforge/internal/handler"
	"github.com/jsndz/authforge/internal/repository"
	"github.com/jsndz/authforge/internal/services"
	"github.com/jsndz/authforge/pkg/email"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AppContainer struct {
	UserHandler  *handler.UserHandler
	TokenHandler *handler.TokenHandler
}

func InitApp(db *gorm.DB, redis *redis.Client, jwtSecret string) *AppContainer {
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)

	tokenService := services.NewTokenService(tokenRepo)
	sessionService := services.NewSessionService(jwtSecret, redis)
	emailService := email.NewEmailService()
	userService := services.NewUserService(userRepo, tokenService, sessionService, emailService, redis)

	userHandler := handler.NewUserHandler(userService)
	tokenHandler := handler.NewTokenHandler(tokenService)

	return &AppContainer{
		UserHandler:  userHandler,
		TokenHandler: tokenHandler,
	}
}
