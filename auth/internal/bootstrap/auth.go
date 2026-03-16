package bootstrap

import (
	"github.com/jsndz/authforge/internal/handler"
	"github.com/jsndz/authforge/internal/repository"
	"github.com/jsndz/authforge/internal/services"
	"gorm.io/gorm"
)

func InitAuthModule(db *gorm.DB) *handler.UserHandler {
	userRepo := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	return handler.NewUserHandler(userService)
}
