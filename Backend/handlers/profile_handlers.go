package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	database "movie-server-backend/db"
	"movie-server-backend/models"
)

func ListProfiles(c *gin.Context) {
	var profiles []models.Profile
	database.DB.Order("id asc").Find(&profiles)
	c.JSON(http.StatusOK, profiles)
}

func CreateProfile(c *gin.Context) {
	var input models.ProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	color := input.AvatarColor
	if color == "" {
		color = "#e50914"
	}
	profile := models.Profile{
		Name:        input.Name,
		AvatarColor: color,
		CreatedAt:   time.Now().Unix(),
	}
	if err := database.DB.Create(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create profile"})
		return
	}
	c.JSON(http.StatusCreated, profile)
}

func UpdateProfile(c *gin.Context) {
	id := c.Param("id")
	var profile models.Profile
	if err := database.DB.First(&profile, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}
	var input models.ProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	profile.Name = input.Name
	if input.AvatarColor != "" {
		profile.AvatarColor = input.AvatarColor
	}
	database.DB.Save(&profile)
	c.JSON(http.StatusOK, profile)
}

func DeleteProfile(c *gin.Context) {
	id := c.Param("id")
	var profile models.Profile
	if err := database.DB.First(&profile, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}
	database.DB.Where("profile_id = ?", profile.ID).Delete(&models.WatchProgress{})
	database.DB.Delete(&profile)
	c.Status(http.StatusNoContent)
}
