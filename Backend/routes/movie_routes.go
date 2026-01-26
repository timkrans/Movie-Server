package routes

import (
	"github.com/gin-gonic/gin"
	"movie-server-backend/handlers"
)

func RegisterMovieRoutes(r *gin.Engine) {
	r.POST("/movies", handlers.CreateMovie)
	r.PUT("/movies/:id", handlers.UpdateMovie)
	r.DELETE("/movies/:id", handlers.DeleteMovie)
	r.GET("/movies/:id", handlers.StreamMovie)
}
