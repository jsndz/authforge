package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jsndz/authforge/internal/bootstrap"
	"gorm.io/gorm"
)

func AuthRouter(router *gin.RouterGroup, db *gorm.DB) {
	userhandler := bootstrap.InitAuthModule(db)

	router.POST("/signup", userhandler.Register)
	router.POST("/login", userhandler.Login)
	router.GET("/email/verify", userhandler.VerifyEmail)
}
