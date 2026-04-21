package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jsndz/authforge/internal/handler"
	"github.com/jsndz/authforge/internal/middleware"
)

func AuthRouter(router *gin.RouterGroup, userHandler *handler.UserHandler, tokenHandler *handler.TokenHandler, oauthHandler *handler.OauthHandler, jwtSecret string) {

	router.POST("/signup", userHandler.Register)
	router.POST("/login", userHandler.Login)
	router.GET("/email/verify", userHandler.VerifyEmail)
	router.PATCH("/update/username", middleware.AuthenticateUser(jwtSecret), userHandler.UpdateUsername)
	router.POST("/reset/password", userHandler.RequestPasswordReset)
	router.GET("/logout", middleware.AuthenticateUser(jwtSecret), userHandler.Logout)
	router.GET("/logout/all", middleware.AuthenticateUser(jwtSecret), userHandler.CompleteLogout)
	router.POST("/refresh", middleware.AuthenticateUser(jwtSecret), userHandler.RefreshToken)
	router.GET("/oauth/authorize", oauthHandler.Authorize)
	router.POST("/oauth/token", oauthHandler.Token)
}
