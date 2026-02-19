package routes

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"movie-server-backend/handlers"
)

func RegisterMovieRoutes(r *gin.Engine) {
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Cache-Control", "no-cache")
	})

	r.POST("/movies", handlers.CreateMovie)
	r.PUT("/movies/:id", handlers.UpdateMovie)
	r.DELETE("/movies/:id", handlers.DeleteMovie)
	r.GET("/movies/:id/stream", handlers.StreamMovie)
	r.GET("/movies", handlers.GetAllMovies)
	r.GET("/movies/:id/hls", handlers.StreamHLS)
	r.Static("/movies/hls", "./movies/hls")
	//adding a health check for later testing
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

}
