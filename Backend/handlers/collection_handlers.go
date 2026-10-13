package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	database "movie-server-backend/db"
	"movie-server-backend/models"
	"movie-server-backend/utils"
)

func ListCollections(c *gin.Context) {
	var collections []models.Collection
	database.DB.Order("id desc").Find(&collections)
	c.JSON(http.StatusOK, collections)
}

func GetCollection(c *gin.Context) {
	id := c.Param("id")
	var collection models.Collection
	if err := database.DB.
		Preload("Movies", func(db *gorm.DB) *gorm.DB {
			return db.Order("collection_order asc, id asc")
		}).
		First(&collection, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		return
	}
	c.JSON(http.StatusOK, collection)
}

func CreateCollection(c *gin.Context) {
	var input models.CollectionInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coverPath := ""
	coverFile, coverHeader, coverErr := c.Request.FormFile("cover_image")
	if coverErr == nil {
		defer coverFile.Close()
		storage := utils.LocalStorage{BasePath: "./movies"}
		path, err := storage.UploadFile(coverFile, coverHeader)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload cover"})
			return
		}
		coverPath = path
	}

	collection := models.Collection{
		Title:              input.Title,
		CoverImageFilePath: coverPath,
	}
	if err := database.DB.Create(&collection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create collection"})
		return
	}
	c.JSON(http.StatusCreated, collection)
}

func UpdateCollection(c *gin.Context) {
	id := c.Param("id")
	var collection models.Collection
	if err := database.DB.First(&collection, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		return
	}
	var input models.CollectionInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	collection.Title = input.Title

	coverFile, coverHeader, coverErr := c.Request.FormFile("cover_image")
	if coverErr == nil {
		defer coverFile.Close()
		storage := utils.LocalStorage{BasePath: "./movies"}
		path, err := storage.UploadFile(coverFile, coverHeader)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload cover"})
			return
		}
		if collection.CoverImageFilePath != "" {
			os.Remove(collection.CoverImageFilePath)
		}
		collection.CoverImageFilePath = path
	}

	database.DB.Save(&collection)
	c.JSON(http.StatusOK, collection)
}

func DeleteCollection(c *gin.Context) {
	id := c.Param("id")
	var collection models.Collection
	if err := database.DB.First(&collection, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		return
	}
	// Detach movies, don't delete them
	database.DB.Model(&models.Movie{}).
		Where("collection_id = ?", collection.ID).
		Updates(map[string]interface{}{"collection_id": nil, "collection_order": nil})
	if collection.CoverImageFilePath != "" {
		os.Remove(collection.CoverImageFilePath)
	}
	database.DB.Delete(&collection)
	c.Status(http.StatusNoContent)
}

func GetCollectionCover(c *gin.Context) {
	id := c.Param("id")
	var collection models.Collection
	if err := database.DB.First(&collection, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		return
	}
	if collection.CoverImageFilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no cover image"})
		return
	}
	if _, err := os.Stat(collection.CoverImageFilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "cover file missing"})
		return
	}
	c.File(collection.CoverImageFilePath)
}

func AttachMovieToCollection(c *gin.Context) {
	id := c.Param("id")
	var collection models.Collection
	if err := database.DB.First(&collection, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		return
	}
	var input models.AttachMovieInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var movie models.Movie
	if err := database.DB.First(&movie, input.MovieID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		return
	}
	order := input.Order
	cid := collection.ID
	movie.CollectionID = &cid
	movie.CollectionOrder = &order
	database.DB.Save(&movie)
	c.JSON(http.StatusOK, movie)
}
