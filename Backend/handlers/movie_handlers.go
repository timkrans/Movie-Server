package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"movie-server-backend/db"
	"movie-server-backend/models"
	"movie-server-backend/utils"
)

func CreateMovie(c *gin.Context) {
	var input models.MovieInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	path, err := utils.ValidateAndResolvePath(input.VideoFilePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movie := models.Movie{
		Title:    input.Title,
		VideoFilePath: path,
	}

	if err := database.DB.Create(&movie).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create movie"})
		return
	}

	c.JSON(http.StatusCreated, movie)
}

func UpdateMovie(c *gin.Context) {
	id := c.Param("id")

	var movie models.Movie
	if err := database.DB.First(&movie, id).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	var input models.MovieInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	path, err := utils.ValidateAndResolvePath(input.VideoFilePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movie.Title = input.Title
	movie.VideoFilePath = path

	database.DB.Save(&movie)
	c.JSON(http.StatusOK, movie)
}

func DeleteMovie(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&models.Movie{}, id)
	c.Status(http.StatusNoContent)
}

func StreamMovie(c *gin.Context) {
	id := c.Param("id")

	var movie models.Movie
	if err := database.DB.First(&movie, id).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	file, err := os.Open(movie.VideoFilePath)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, _ := file.Stat()

	c.Header("Content-Type", "video/mp4")
	c.Header("Accept-Ranges", "bytes")

	http.ServeContent(
		c.Writer,
		c.Request,
		stat.Name(),
		stat.ModTime(),
		file,
	)
}
