package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	database "movie-server-backend/db"
	"movie-server-backend/models"
)

func profileIDFromHeader(c *gin.Context) (uint, bool) {
	raw := c.GetHeader("X-Profile-Id")
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Profile-Id header required"})
		return 0, false
	}
	id64, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid X-Profile-Id"})
		return 0, false
	}
	var profile models.Profile
	if err := database.DB.First(&profile, id64).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "profile not found"})
		return 0, false
	}
	return profile.ID, true
}

func UpsertProgress(c *gin.Context) {
	profileID, ok := profileIDFromHeader(c)
	if !ok {
		return
	}
	var input models.ProgressInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.ContentType != "movie" && input.ContentType != "episode" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content_type must be movie or episode"})
		return
	}

	completed := false
	if input.DurationSeconds > 0 && input.PositionSeconds/input.DurationSeconds >= 0.9 {
		completed = true
	}

	var progress models.WatchProgress
	err := database.DB.Where(
		"profile_id = ? AND content_type = ? AND content_id = ?",
		profileID, input.ContentType, input.ContentID,
	).First(&progress).Error

	now := time.Now().Unix()
	if err != nil {
		progress = models.WatchProgress{
			ProfileID:       profileID,
			ContentType:     input.ContentType,
			ContentID:       input.ContentID,
			PositionSeconds: input.PositionSeconds,
			DurationSeconds: input.DurationSeconds,
			Completed:       completed,
			UpdatedAt:       now,
		}
		if err := database.DB.Create(&progress).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save progress"})
			return
		}
	} else {
		progress.PositionSeconds = input.PositionSeconds
		progress.DurationSeconds = input.DurationSeconds
		progress.Completed = completed
		progress.UpdatedAt = now
		database.DB.Save(&progress)
	}
	c.JSON(http.StatusOK, progress)
}

func GetProgress(c *gin.Context) {
	profileID, ok := profileIDFromHeader(c)
	if !ok {
		return
	}
	contentType := c.Param("type")
	contentID := c.Param("id")
	if contentType != "movie" && contentType != "episode" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
		return
	}

	var progress models.WatchProgress
	if err := database.DB.Where(
		"profile_id = ? AND content_type = ? AND content_id = ?",
		profileID, contentType, contentID,
	).First(&progress).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"position_seconds": 0,
			"duration_seconds": 0,
			"completed":        false,
		})
		return
	}
	c.JSON(http.StatusOK, progress)
}

func ListContinueWatching(c *gin.Context) {
	profileID, ok := profileIDFromHeader(c)
	if !ok {
		return
	}

	var rows []models.WatchProgress
	database.DB.
		Where("profile_id = ? AND completed = ? AND position_seconds > ?", profileID, false, 5).
		Order("updated_at desc").
		Limit(24).
		Find(&rows)

	items := make([]models.ContinueItem, 0, len(rows))
	for _, p := range rows {
		item, ok := enrichProgress(p)
		if ok {
			items = append(items, item)
		}
	}
	c.JSON(http.StatusOK, items)
}

func enrichProgress(p models.WatchProgress) (models.ContinueItem, bool) {
	progressPct := 0.0
	if p.DurationSeconds > 0 {
		progressPct = p.PositionSeconds / p.DurationSeconds
	}

	switch p.ContentType {
	case "movie":
		var movie models.Movie
		if database.DB.First(&movie, p.ContentID).Error != nil {
			return models.ContinueItem{}, false
		}
		cover := ""
		if movie.CoverImageFilePath != "" {
			cover = fmt.Sprintf("/movies/%d/cover", movie.ID)
		}
		return models.ContinueItem{
			ContentType:     "movie",
			ContentID:       movie.ID,
			Title:           movie.Title,
			CoverURL:        cover,
			WatchURL:        fmt.Sprintf("/watch/%d", movie.ID),
			PositionSeconds: p.PositionSeconds,
			DurationSeconds: p.DurationSeconds,
			Progress:        progressPct,
			UpdatedAt:       p.UpdatedAt,
		}, true

	case "episode":
		var episode models.Episode
		if database.DB.First(&episode, p.ContentID).Error != nil {
			return models.ContinueItem{}, false
		}
		var season models.Season
		subtitle := fmt.Sprintf("S%d E%d", 0, episode.EpisodeNumber)
		showTitle := episode.Title
		if database.DB.First(&season, episode.SeasonID).Error == nil {
			subtitle = fmt.Sprintf("S%d E%d", season.SeasonNumber, episode.EpisodeNumber)
			var show models.Show
			if database.DB.First(&show, season.ShowID).Error == nil {
				showTitle = show.Title
				subtitle = fmt.Sprintf("%s · %s", subtitle, episode.Title)
			}
		}
		cover := fmt.Sprintf("/episodes/%d/cover", episode.ID)
		return models.ContinueItem{
			ContentType:     "episode",
			ContentID:       episode.ID,
			Title:           showTitle,
			Subtitle:        subtitle,
			CoverURL:        cover,
			WatchURL:        fmt.Sprintf("/watch/episode/%d", episode.ID),
			PositionSeconds: p.PositionSeconds,
			DurationSeconds: p.DurationSeconds,
			Progress:        progressPct,
			UpdatedAt:       p.UpdatedAt,
		}, true
	}
	return models.ContinueItem{}, false
}
