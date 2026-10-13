package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"movie-server-backend/db"
	"movie-server-backend/models"
	"movie-server-backend/routes"
	"movie-server-backend/loadenv"
)

func main() {
	// .env is optional — FFMPEG_PATH falls back to "ffmpeg" on PATH
	if err := loadenv.LoadEnv(""); err != nil {
		log.Printf("No .env loaded (%v); continuing with process environment", err)
	}

	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}

	database.DB.AutoMigrate(
		&models.Movie{},
		&models.LiveStream{},
		&models.Profile{},
		&models.Collection{},
		&models.Show{},
		&models.Season{},
		&models.Episode{},
		&models.WatchProgress{},
	)
	// Backfill status for movies created before the status column existed.
	database.DB.Model(&models.Movie{}).
		Where("status = '' OR status IS NULL").
		Where("hls_path != '' AND hls_path IS NOT NULL").
		Update("status", "ready")
	database.DB.Model(&models.Movie{}).
		Where("status = '' OR status IS NULL").
		Update("status", "processing")

	r := gin.Default()
	r.SetTrustedProxies(nil)
	routes.RegisterMovieRoutes(r)

	r.Run(":8080")
}
