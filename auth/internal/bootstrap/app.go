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
	OauthHandler *handler.OauthHandler
}

func InitApp(db *gorm.DB, redis *redis.Client, jwtSecret string, GOOGLE_CLIENT_ID string, GOOGLE_CALLBACK_URL string, GOOGLE_CLIENT_SECRET string) *AppContainer {
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	oauthRepo := repository.NewOauthRepo(db)

	tokenService := services.NewTokenService(tokenRepo)
	sessionService := services.NewSessionService(jwtSecret, redis)
	emailService := email.NewEmailService()
	userService := services.NewUserService(userRepo, tokenService, sessionService, emailService, redis)
	oauthService := services.NewOAuthService(oauthRepo, sessionService, userService, redis, jwtSecret, GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, GOOGLE_CALLBACK_URL)

	userHandler := handler.NewUserHandler(userService)
	tokenHandler := handler.NewTokenHandler(tokenService)
	oauthHandler := handler.NewOauthHandler(oauthService)

	return &AppContainer{
		UserHandler:  userHandler,
		TokenHandler: tokenHandler,
		OauthHandler: oauthHandler,
	}
}
