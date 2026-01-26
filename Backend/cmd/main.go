package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"movie-server-backend/db"
	"movie-server-backend/models"
	"movie-server-backend/routes"
)

func main() {
	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}

	database.DB.AutoMigrate(&models.Movie{})
	r := gin.Default()
	r.SetTrustedProxies(nil)
	routes.RegisterMovieRoutes(r)

	r.Run(":8080")
}
