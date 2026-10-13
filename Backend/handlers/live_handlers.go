package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	database "movie-server-backend/db"
	"movie-server-backend/models"
	"movie-server-backend/utils"
)

func newStreamKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func init() {
	utils.OnLiveEnded = func(streamID uint) {
		database.DB.Model(&models.LiveStream{}).
			Where("id = ? AND status = ?", streamID, "live").
			Update("status", "ended")
	}
}

// reconcileLiveStatus marks DB "live" rows as ended when FFmpeg is no longer running them.
func reconcileLiveStatus(stream *models.LiveStream) {
	if stream.Status != "live" {
		return
	}
	if utils.Live.ActiveID() == stream.ID && utils.Live.IsBusy() {
		return
	}
	stream.Status = "ended"
	database.DB.Model(stream).Update("status", "ended")
}

func CreateLiveStream(c *gin.Context) {
	var input models.LiveStreamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := newStreamKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate stream key"})
		return
	}

	stream := models.LiveStream{
		Title:     input.Title,
		StreamKey: key,
		Status:    "idle",
		HLSDir:    "",
		CreatedAt: time.Now().Unix(),
	}

	if err := database.DB.Create(&stream).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create live stream"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          stream.ID,
		"title":       stream.Title,
		"stream_key":  stream.StreamKey,
		"status":      stream.Status,
		"created_at":  stream.CreatedAt,
		"obs_server":  "rtmp://localhost:1935/live",
		"obs_stream_key": stream.StreamKey,
	})
}

func ListLiveStreams(c *gin.Context) {
	var streams []models.LiveStream
	database.DB.Order("id desc").Find(&streams)
	for i := range streams {
		reconcileLiveStatus(&streams[i])
	}
	c.JSON(http.StatusOK, streams)
}

func GetLiveStream(c *gin.Context) {
	id := c.Param("id")
	var stream models.LiveStream
	if err := database.DB.First(&stream, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "live stream not found"})
		return
	}
	reconcileLiveStatus(&stream)
	c.JSON(http.StatusOK, gin.H{
		"id":             stream.ID,
		"title":          stream.Title,
		"stream_key":     stream.StreamKey,
		"status":         stream.Status,
		"hls_dir":        stream.HLSDir,
		"created_at":     stream.CreatedAt,
		"obs_server":     "rtmp://localhost:1935/live",
		"obs_stream_key": stream.StreamKey,
	})
}

func StartLiveStream(c *gin.Context) {
	id := c.Param("id")
	var stream models.LiveStream
	if err := database.DB.First(&stream, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "live stream not found"})
		return
	}

	if stream.Status == "ended" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stream has ended; create a new one"})
		return
	}

	if utils.Live.IsBusy() && utils.Live.ActiveID() != stream.ID {
		c.JSON(http.StatusConflict, gin.H{
			"error": fmt.Sprintf("another live session is already active (id=%d)", utils.Live.ActiveID()),
		})
		return
	}

	if utils.Live.ActiveID() == stream.ID {
		c.JSON(http.StatusOK, gin.H{"status": "live", "id": stream.ID})
		return
	}

	hlsDir := fmt.Sprintf("./movies/live/%d", stream.ID)
	os.RemoveAll(hlsDir)

	if err := utils.Live.StartListen(stream.ID, stream.StreamKey, hlsDir); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	stream.Status = "live"
	stream.HLSDir = hlsDir
	database.DB.Save(&stream)

	c.JSON(http.StatusOK, gin.H{
		"status":         "live",
		"id":             stream.ID,
		"obs_server":     "rtmp://localhost:1935/live",
		"obs_stream_key": stream.StreamKey,
		"hls_url":        fmt.Sprintf("/live/%d/hls/index.m3u8", stream.ID),
	})
}

func StopLiveStream(c *gin.Context) {
	id := c.Param("id")
	var stream models.LiveStream
	if err := database.DB.First(&stream, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "live stream not found"})
		return
	}

	_ = utils.Live.Stop(stream.ID)

	stream.Status = "ended"
	database.DB.Save(&stream)

	c.JSON(http.StatusOK, gin.H{"status": "ended", "id": stream.ID})
}

func StreamLiveHLS(c *gin.Context) {
	id := c.Param("id")
	fp := strings.TrimPrefix(c.Param("filepath"), "/")

	var stream models.LiveStream
	if err := database.DB.First(&stream, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "live stream not found"})
		return
	}

	reconcileLiveStatus(&stream)
	if stream.Status != "live" {
		c.JSON(http.StatusGone, gin.H{"error": "stream is not live"})
		return
	}

	if stream.HLSDir == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "HLS not available yet"})
		return
	}

	if fp == "" || fp == "hls" {
		fp = "index.m3u8"
	}

	fullPath := filepath.Join(stream.HLSDir, fp)

	// Prevent path traversal
	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(stream.HLSDir)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid path"})
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found — is OBS connected?"})
		return
	}

	if strings.HasSuffix(fullPath, ".m3u8") {
		c.Header("Content-Type", "application/vnd.apple.mpegurl")
	} else if strings.HasSuffix(fullPath, ".ts") {
		c.Header("Content-Type", "video/mp2t")
	}

	c.File(fullPath)
}
