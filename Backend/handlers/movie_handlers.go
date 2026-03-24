package handlers

import (
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/gin-gonic/gin"

    "movie-server-backend/db"
    "movie-server-backend/models"
    "movie-server-backend/utils"
)

func CreateMovie(c *gin.Context) {
    var input models.MovieInput
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

    coverFile, coverHeader, _ := c.Request.FormFile("cover_image")

    ext := strings.ToLower(filepath.Ext(videoHeader.Filename))
    allowed := map[string]bool{
        ".mp4": true,
        ".mov": true,
        ".zip": true,
    }

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

    if coverFile != nil {
        defer coverFile.Close()
        uploadedCoverPath, err := storage.UploadFile(coverFile, coverHeader)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload cover image"})
            return
        }
        finalCoverPath = uploadedCoverPath
    }
    movie := models.Movie{
        Title:              input.Title,
        VideoFilePath:      finalVideoPath,
        CoverImageFilePath: finalCoverPath,
        HLSPath:            "", //not ready yet
    }

    if err := database.DB.Create(&movie).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create movie"})
        return
    }

    //background HLS generation
    go func(movieID uint, videoPath string) {
        hlsDir := fmt.Sprintf("./movies/hls/%d", time.Now().UnixNano())

        fmt.Println("Starting HLS for movie:", movieID)

        hlsPath, err := utils.GenerateHLS(videoPath, hlsDir)
        if err != nil {
            fmt.Println("HLS generation failed:", err)
            return
        }

        database.DB.Model(&models.Movie{}).
            Where("id = ?", movieID).
            Update("hls_path", hlsPath)

        fmt.Println("Finished HLS for movie:", movieID)
    }(movie.ID, finalVideoPath)
    c.JSON(http.StatusCreated, movie)
}

func StreamHLS(c *gin.Context) {
    id := c.Param("id")
    fp := strings.TrimPrefix(c.Param("filepath"), "/")

    var movie models.Movie
    if err := database.DB.First(&movie, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
        return
    }

    if movie.HLSPath == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "HLS not available"})
        return
    }

    baseDir := filepath.Dir(movie.HLSPath)

    if fp == "" || fp == "hls" {
        fp = filepath.Base(movie.HLSPath)
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
    ext := strings.ToLower(filepath.Ext(movie.VideoFilePath))
    contentType := "video/mp4" 
    if ext == ".mov" { 
        contentType = "video/quicktime" 
    } 
    c.Header("Content-Type", contentType)
    c.Header("Accept-Ranges", "bytes")

    http.ServeContent(
        c.Writer,
        c.Request,
        stat.Name(),
        stat.ModTime(),
        file,
    )
}

func GetAllMovies(c *gin.Context) {
    var movies []models.Movie
    database.DB.Find(&movies)
    c.JSON(http.StatusOK, movies)
}

func DeleteMovie(c *gin.Context) {
    id := c.Param("id")

    var movie models.Movie
    if err := database.DB.First(&movie, id).Error; err != nil {
        c.Status(http.StatusNotFound)
        return
    }

    if movie.VideoFilePath != "" {
        os.Remove(movie.VideoFilePath)
    }
    if movie.HLSPath != "" { 
        hlsDir := filepath.Dir(movie.HLSPath)
         //ensure permissions allow deletion 
         os.Chmod(hlsDir, 0777) 
         os.RemoveAll(hlsDir)
     }
    database.DB.Delete(&models.Movie{}, id)
    c.Status(http.StatusNoContent)
}

func UpdateMovie(c *gin.Context) {
    id := c.Param("id")

    var movie models.Movie
    if err := database.DB.First(&movie, id).Error; err != nil {
        c.Status(http.StatusNotFound)
        return
    }

    var input models.MovieInput
    if err := c.ShouldBind(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    videoFile, videoHeader, videoErr := c.Request.FormFile("video")
    coverFile, coverHeader, coverErr := c.Request.FormFile("cover_image")

    storage := utils.LocalStorage{BasePath: "./movies"}

    if videoErr == nil {
        defer videoFile.Close()

        ext := strings.ToLower(filepath.Ext(videoHeader.Filename))
        allowed := map[string]bool{
            ".mp4": true,
            ".mov": true,
            ".zip": true,
        }

        if !allowed[ext] {
            c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported video file type"})
            return
        }

        uploadedVideoPath, err := storage.UploadFile(videoFile, videoHeader)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload video"})
            return
        }

        finalVideoPath := uploadedVideoPath
        finalCoverPath := movie.CoverImageFilePath

        if ext == ".zip" {
            videoPath, imagePath, err := utils.ExtractZip(uploadedVideoPath, "./movies/extracted")
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "zip uploaded but no video found"})
                return
            }
            finalVideoPath = videoPath
            if imagePath != "" {
                finalCoverPath = imagePath
            }
        }

        if movie.VideoFilePath != "" {
            os.Remove(movie.VideoFilePath)
        }
        if movie.HLSPath != "" {
            hlsDir := filepath.Dir(movie.HLSPath)
            os.Chmod(hlsDir, 0777)
            os.RemoveAll(hlsDir)
        }

        movie.VideoFilePath = finalVideoPath
        movie.CoverImageFilePath = finalCoverPath
        movie.HLSPath = ""
        database.DB.Save(&movie)

        go func(movieID uint, videoPath string) {
            hlsDir := fmt.Sprintf("./movies/hls/%d", time.Now().UnixNano())
            hlsPath, err := utils.GenerateHLS(videoPath, hlsDir)
            if err != nil {
                return
            }
            database.DB.Model(&models.Movie{}).
                Where("id = ?", movieID).
                Update("hls_path", hlsPath)
        }(movie.ID, finalVideoPath)
    }

    if coverErr == nil {
        defer coverFile.Close()

        uploadedCoverPath, err := storage.UploadFile(coverFile, coverHeader)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload cover image"})
            return
        }

        if movie.CoverImageFilePath != "" {
            os.Remove(movie.CoverImageFilePath)
        }

        movie.CoverImageFilePath = uploadedCoverPath
    }

    movie.Title = input.Title
    database.DB.Save(&movie)

    c.JSON(http.StatusOK, movie)
}