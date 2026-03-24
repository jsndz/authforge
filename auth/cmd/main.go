package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/authforge/internal/bootstrap"
	"github.com/jsndz/authforge/internal/config"
	"github.com/jsndz/authforge/internal/routes"
	"github.com/jsndz/authforge/pkg/db"
	"github.com/jsndz/authforge/pkg/redis"
)

func main() {
	router := gin.Default()
	cfg := config.Load()
	database, err := db.InitDB(cfg.DBConnectURL)
	db.MigrateDB(database)
	if err != nil {
		panic(err)
	}
	redis := redis.NewRedisClient()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	app := bootstrap.InitApp(database, redis, cfg.JWTSecret)
	api := router.Group("/api/v1/auth")
	routes.AuthRouter(api, app.UserHandler, app.TokenHandler, cfg.JWTSecret)
	router.Run()
}
