package routes

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"movie-server-backend/handlers"
)

func RegisterMovieRoutes(r *gin.Engine) {
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Profile-Id")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	r.POST("/movies", handlers.CreateMovie)
	r.PUT("/movies/:id", handlers.UpdateMovie)
	r.DELETE("/movies/:id", handlers.DeleteMovie)
	r.GET("/movies/:id/stream", handlers.StreamMovie)
	r.GET("/movies/:id/cover", handlers.GetMovieCover)
	r.GET("/movies", handlers.GetAllMovies)
	r.GET("/movies/:id/hls", cacheAwareHLS(handlers.StreamHLS))
	r.GET("/movies/:id/hls/*filepath", cacheAwareHLS(handlers.StreamHLS))
	r.GET("/movies/:id", handlers.GetMovie)

	r.POST("/live", handlers.CreateLiveStream)
	r.GET("/live", handlers.ListLiveStreams)
	r.GET("/live/:id", handlers.GetLiveStream)
	r.POST("/live/:id/start", handlers.StartLiveStream)
	r.POST("/live/:id/stop", handlers.StopLiveStream)
	r.GET("/live/:id/hls", cacheAwareHLS(handlers.StreamLiveHLS))
	r.GET("/live/:id/hls/*filepath", cacheAwareHLS(handlers.StreamLiveHLS))

	r.GET("/profiles", handlers.ListProfiles)
	r.POST("/profiles", handlers.CreateProfile)
	r.PUT("/profiles/:id", handlers.UpdateProfile)
	r.DELETE("/profiles/:id", handlers.DeleteProfile)

	r.GET("/collections", handlers.ListCollections)
	r.POST("/collections", handlers.CreateCollection)
	r.GET("/collections/:id", handlers.GetCollection)
	r.PUT("/collections/:id", handlers.UpdateCollection)
	r.DELETE("/collections/:id", handlers.DeleteCollection)
	r.GET("/collections/:id/cover", handlers.GetCollectionCover)
	r.POST("/collections/:id/movies", handlers.AttachMovieToCollection)

	r.GET("/shows", handlers.ListShows)
	r.POST("/shows", handlers.CreateShow)
	r.GET("/shows/:id", handlers.GetShow)
	r.PUT("/shows/:id", handlers.UpdateShow)
	r.DELETE("/shows/:id", handlers.DeleteShow)
	r.GET("/shows/:id/cover", handlers.GetShowCover)
	r.POST("/shows/:id/seasons", handlers.CreateSeason)
	r.POST("/shows/:id/episodes", handlers.CreateEpisodeOnShow)

	r.POST("/seasons/:id/episodes", handlers.CreateEpisode)

	r.GET("/episodes/:id", handlers.GetEpisode)
	r.DELETE("/episodes/:id", handlers.DeleteEpisode)
	r.GET("/episodes/:id/cover", handlers.GetEpisodeCover)
	r.GET("/episodes/:id/next", handlers.GetNextEpisode)
	r.GET("/episodes/:id/hls", cacheAwareHLS(handlers.StreamEpisodeHLS))
	r.GET("/episodes/:id/hls/*filepath", cacheAwareHLS(handlers.StreamEpisodeHLS))

	r.PUT("/progress", handlers.UpsertProgress)
	r.GET("/progress/continue", handlers.ListContinueWatching)
	r.GET("/progress/:type/:id", handlers.GetProgress)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func cacheAwareHLS(next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		fp := c.Param("filepath")
		ext := strings.ToLower(filepath.Ext(fp))
		switch ext {
		case ".ts":
			c.Header("Cache-Control", "public, max-age=86400")
		case ".m3u8":
			c.Header("Cache-Control", "public, max-age=2")
		default:
			c.Header("Cache-Control", "no-cache")
		}
		next(c)
	}
}
