package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jsndz/authforge/internal/handler"
)

func AuthRouter(router *gin.RouterGroup, userHandler *handler.UserHandler, tokenHandler *handler.TokenHandler) {

	router.POST("/signup", userHandler.Register)
	router.POST("/login", userHandler.Login)
	router.GET("/email/verify", userHandler.VerifyEmail)
}
