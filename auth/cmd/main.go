package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/authforge/internal/routes"
	"github.com/jsndz/authforge/pkg/db"
)

func main() {
	router := gin.Default()

	database, err := db.InitDB()
	db.MigrateDB(database)
	if err != nil {
		panic(err)
	}

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	api := router.Group("/api/v1/auth")
	routes.AuthRouter(api, database)
	router.Run()
}
