package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	database "movie-server-backend/db"
	"movie-server-backend/models"
	"movie-server-backend/utils"
)

func ListShows(c *gin.Context) {
	var shows []models.Show
	database.DB.Order("id desc").Find(&shows)
	c.JSON(http.StatusOK, shows)
}

func GetShow(c *gin.Context) {
	id := c.Param("id")
	var show models.Show
	if err := database.DB.
		Preload("Seasons", func(db *gorm.DB) *gorm.DB {
			return db.Order("season_number asc")
		}).
		Preload("Seasons.Episodes", func(db *gorm.DB) *gorm.DB {
			return db.Order("episode_number asc")
		}).
		First(&show, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "show not found"})
		return
	}
	c.JSON(http.StatusOK, show)
}

func CreateShow(c *gin.Context) {
	var input models.ShowInput
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

	show := models.Show{
		Title:              input.Title,
		CoverImageFilePath: coverPath,
	}
	if err := database.DB.Create(&show).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create show"})
		return
	}
	c.JSON(http.StatusCreated, show)
}

func UpdateShow(c *gin.Context) {
	id := c.Param("id")
	var show models.Show
	if err := database.DB.First(&show, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "show not found"})
		return
	}
	var input models.ShowInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	show.Title = input.Title

	coverFile, coverHeader, coverErr := c.Request.FormFile("cover_image")
	if coverErr == nil {
		defer coverFile.Close()
		storage := utils.LocalStorage{BasePath: "./movies"}
		path, err := storage.UploadFile(coverFile, coverHeader)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload cover"})
			return
		}
		if show.CoverImageFilePath != "" {
			os.Remove(show.CoverImageFilePath)
		}
		show.CoverImageFilePath = path
	}

	database.DB.Save(&show)
	c.JSON(http.StatusOK, show)
}

func DeleteShow(c *gin.Context) {
	id := c.Param("id")
	var show models.Show
	if err := database.DB.
		Preload("Seasons.Episodes").
		First(&show, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "show not found"})
		return
	}
	for _, season := range show.Seasons {
		for _, ep := range season.Episodes {
			deleteEpisodeFiles(ep)
			database.DB.Delete(&ep)
		}
		database.DB.Delete(&season)
	}
	if show.CoverImageFilePath != "" {
		os.Remove(show.CoverImageFilePath)
	}
	database.DB.Delete(&show)
	c.Status(http.StatusNoContent)
}

func GetShowCover(c *gin.Context) {
	id := c.Param("id")
	var show models.Show
	if err := database.DB.First(&show, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "show not found"})
		return
	}
	if show.CoverImageFilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no cover image"})
		return
	}
	if _, err := os.Stat(show.CoverImageFilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "cover file missing"})
		return
	}
	c.File(show.CoverImageFilePath)
}

func CreateSeason(c *gin.Context) {
	showID := c.Param("id")
	var show models.Show
	if err := database.DB.First(&show, showID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "show not found"})
		return
	}
	var input models.SeasonInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	season := models.Season{
		ShowID:       show.ID,
		SeasonNumber: input.SeasonNumber,
		Title:        input.Title,
	}
	if err := database.DB.Create(&season).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create season"})
		return
	}
	c.JSON(http.StatusCreated, season)
}

// CreateEpisodeOnShow accepts season_number and auto-creates the season if needed.
func CreateEpisodeOnShow(c *gin.Context) {
	showID := c.Param("id")
	var show models.Show
	if err := database.DB.First(&show, showID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "show not found"})
		return
	}

	seasonNumStr := c.PostForm("season_number")
	seasonNum, err := strconv.Atoi(seasonNumStr)
	if err != nil || seasonNum < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "season_number is required"})
		return
	}

	season, err := EnsureOrCreateSeason(show.ID, seasonNum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create season"})
		return
	}

	createEpisodeForSeason(c, season)
}

func CreateEpisode(c *gin.Context) {
	seasonID := c.Param("id")
	var season models.Season
	if err := database.DB.First(&season, seasonID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "season not found"})
		return
	}
	createEpisodeForSeason(c, season)
}

func createEpisodeForSeason(c *gin.Context, season models.Season) {
	var input models.EpisodeInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	videoFile, videoHeader, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video or zip file is required"})
		return
	}
	defer videoFile.Close()

	ext := strings.ToLower(filepath.Ext(videoHeader.Filename))
	allowed := map[string]bool{".mp4": true, ".mov": true, ".zip": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported video file type"})
		return
	}

	storage := utils.LocalStorage{BasePath: "./movies"}
	uploadedVideoPath, err := storage.UploadFile(videoFile, videoHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload video"})
		return
	}

	finalVideoPath := uploadedVideoPath
	finalCoverPath := ""

	if ext == ".zip" {
		videoPath, imagePath, err := utils.ExtractZip(uploadedVideoPath, "./movies/extracted")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "zip uploaded but no video found"})
			return
		}
		finalVideoPath = videoPath
		finalCoverPath = imagePath
	}

	coverFile, coverHeader, coverErr := c.Request.FormFile("cover_image")
	if coverErr == nil {
		defer coverFile.Close()
		uploadedCoverPath, err := storage.UploadFile(coverFile, coverHeader)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload cover image"})
			return
		}
		finalCoverPath = uploadedCoverPath
	}

	episode := models.Episode{
		SeasonID:           season.ID,
		EpisodeNumber:      input.EpisodeNumber,
		Title:              input.Title,
		VideoFilePath:      finalVideoPath,
		CoverImageFilePath: finalCoverPath,
		Status:             "processing",
	}
	if err := database.DB.Create(&episode).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create episode"})
		return
	}

	go generateEpisodeHLSAsync(episode.ID, finalVideoPath)
	c.JSON(http.StatusCreated, episode)
}

func generateEpisodeHLSAsync(episodeID uint, videoPath string) {
	hlsDir := fmt.Sprintf("./movies/hls/ep_%d", time.Now().UnixNano())
	fmt.Println("Starting HLS for episode:", episodeID)

	hlsPath, err := utils.GenerateHLS(videoPath, hlsDir)
	if err != nil {
		fmt.Println("Episode HLS generation failed:", err)
		database.DB.Model(&models.Episode{}).
			Where("id = ?", episodeID).
			Updates(map[string]interface{}{"status": "failed"})
		return
	}

	database.DB.Model(&models.Episode{}).
		Where("id = ?", episodeID).
		Updates(map[string]interface{}{
			"hls_path": hlsPath,
			"status":   "ready",
		})
	fmt.Println("Finished HLS for episode:", episodeID)
}

func GetEpisode(c *gin.Context) {
	id := c.Param("id")
	var episode models.Episode
	if err := database.DB.First(&episode, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "episode not found"})
		return
	}

	showID := uint(0)
	showTitle := ""
	seasonNumber := 0
	var season models.Season
	if database.DB.First(&season, episode.SeasonID).Error == nil {
		seasonNumber = season.SeasonNumber
		var show models.Show
		if database.DB.First(&show, season.ShowID).Error == nil {
			showID = show.ID
			showTitle = show.Title
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                    episode.ID,
		"season_id":             episode.SeasonID,
		"episode_number":        episode.EpisodeNumber,
		"title":                 episode.Title,
		"video_file_path":       episode.VideoFilePath,
		"cover_image_file_path": episode.CoverImageFilePath,
		"hls_path":              episode.HLSPath,
		"status":                episode.Status,
		"show_id":               showID,
		"show_title":            showTitle,
		"season_number":         seasonNumber,
	})
}

func GetEpisodeCover(c *gin.Context) {
	id := c.Param("id")
	var episode models.Episode
	if err := database.DB.First(&episode, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "episode not found"})
		return
	}
	if episode.CoverImageFilePath == "" {
		// Fall back to show cover
		var season models.Season
		if database.DB.First(&season, episode.SeasonID).Error == nil {
			var show models.Show
			if database.DB.First(&show, season.ShowID).Error == nil && show.CoverImageFilePath != "" {
				c.File(show.CoverImageFilePath)
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "no cover image"})
		return
	}
	if _, err := os.Stat(episode.CoverImageFilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "cover file missing"})
		return
	}
	c.File(episode.CoverImageFilePath)
}

func StreamEpisodeHLS(c *gin.Context) {
	id := c.Param("id")
	fp := strings.TrimPrefix(c.Param("filepath"), "/")

	var episode models.Episode
	if err := database.DB.First(&episode, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "episode not found"})
		return
	}
	if episode.HLSPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "HLS not available"})
		return
	}

	baseDir := filepath.Dir(episode.HLSPath)
	if fp == "" || fp == "hls" {
		fp = filepath.Base(episode.HLSPath)
	}
	fullPath := filepath.Join(baseDir, fp)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	if strings.HasSuffix(fullPath, ".m3u8") {
		c.Header("Content-Type", "application/vnd.apple.mpegurl")
	} else if strings.HasSuffix(fullPath, ".ts") {
		c.Header("Content-Type", "video/mp2t")
	}
	c.File(fullPath)
}

func DeleteEpisode(c *gin.Context) {
	id := c.Param("id")
	var episode models.Episode
	if err := database.DB.First(&episode, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "episode not found"})
		return
	}
	deleteEpisodeFiles(episode)
	database.DB.Where("content_type = ? AND content_id = ?", "episode", episode.ID).
		Delete(&models.WatchProgress{})
	database.DB.Delete(&episode)
	c.Status(http.StatusNoContent)
}

func GetNextEpisode(c *gin.Context) {
	id := c.Param("id")
	var episode models.Episode
	if err := database.DB.First(&episode, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "episode not found"})
		return
	}

	var next models.Episode
	// Next in same season
	err := database.DB.
		Where("season_id = ? AND episode_number > ?", episode.SeasonID, episode.EpisodeNumber).
		Order("episode_number asc").
		First(&next).Error
	if err == nil {
		c.JSON(http.StatusOK, next)
		return
	}

	// First episode of next season
	var season models.Season
	if database.DB.First(&season, episode.SeasonID).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no next episode"})
		return
	}
	var nextSeason models.Season
	if database.DB.
		Where("show_id = ? AND season_number > ?", season.ShowID, season.SeasonNumber).
		Order("season_number asc").
		First(&nextSeason).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no next episode"})
		return
	}
	if database.DB.
		Where("season_id = ?", nextSeason.ID).
		Order("episode_number asc").
		First(&next).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no next episode"})
		return
	}
	c.JSON(http.StatusOK, next)
}

func deleteEpisodeFiles(ep models.Episode) {
	if ep.VideoFilePath != "" {
		os.Remove(ep.VideoFilePath)
	}
	if ep.CoverImageFilePath != "" {
		os.Remove(ep.CoverImageFilePath)
	}
	if ep.HLSPath != "" {
		hlsDir := filepath.Dir(ep.HLSPath)
		os.Chmod(hlsDir, 0777)
		os.RemoveAll(hlsDir)
	}
}

// EnsureOrCreateSeason finds or creates a season for a show by number.
func EnsureOrCreateSeason(showID uint, seasonNumber int) (models.Season, error) {
	var season models.Season
	err := database.DB.Where("show_id = ? AND season_number = ?", showID, seasonNumber).First(&season).Error
	if err == nil {
		return season, nil
	}
	season = models.Season{
		ShowID:       showID,
		SeasonNumber: seasonNumber,
		Title:        "Season " + strconv.Itoa(seasonNumber),
	}
	if err := database.DB.Create(&season).Error; err != nil {
		return season, err
	}
	return season, nil
}
