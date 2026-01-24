package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const movieDir = "./movies"

func main() {
	r := gin.Default()

	r.GET("/movies/:name", streamMovie)

	r.Run(":8080")
}

func streamMovie(c *gin.Context) {
	name := filepath.Base(c.Param("name"))
	path := filepath.Join(movieDir, name)

	file, err := os.Open(path)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

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
