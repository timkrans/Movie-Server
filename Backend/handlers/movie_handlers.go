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

    //fenerate HLS
    hlsDir := fmt.Sprintf("./movies/hls/%d", time.Now().UnixNano())
    hlsPath, err := utils.GenerateHLS(finalVideoPath, hlsDir)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate HLS"})
        return
    }

    movie := models.Movie{
        Title:              input.Title,
        VideoFilePath:      finalVideoPath,
        CoverImageFilePath: finalCoverPath,
        HLSPath:            hlsPath,
    }

    if err := database.DB.Create(&movie).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create movie"})
        return
    }

    c.JSON(http.StatusCreated, movie)
}

func StreamHLS(c *gin.Context) {
    id := c.Param("id")

    var movie models.Movie
    if err := database.DB.First(&movie, id).Error; err != nil {
        c.Status(http.StatusNotFound)
        return
    }

    if movie.HLSPath == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "HLS not available for this movie"})
        return
    }

    publicPath := strings.TrimPrefix(movie.HLSPath, ".")
    c.Redirect(http.StatusTemporaryRedirect, publicPath)
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
            //ensure permissions allow deletion 
            os.Chmod(hlsDir, 0777) 
            os.RemoveAll(hlsDir)
        }

        hlsDir := fmt.Sprintf("./movies/hls/%d", time.Now().UnixNano())
        hlsPath, err := utils.GenerateHLS(finalVideoPath, hlsDir)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate HLS"})
            return
        }

        movie.VideoFilePath = finalVideoPath
        movie.CoverImageFilePath = finalCoverPath
        movie.HLSPath = hlsPath
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
