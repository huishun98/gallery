package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type Media struct {
	Path string `json:"path"`
	URL  string `json:"url"`
	Type string `json:"type"`
}

func Pending(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		entries, _ := os.ReadDir(target)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}

			ext := strings.ToLower(filepath.Ext(e.Name()))
			mediaType := "image"
			if ext == ".mp4" || ext == ".webm" || ext == ".mov" {
				mediaType = "video"
			}

			c.JSON(http.StatusOK, Media{
				Path: e.Name(),
				URL:  "/files/" + e.Name(),
				Type: mediaType,
			})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func MoveMedia(source, target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Path string `json:"path"`
		}
		if err := c.BindJSON(&body); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		src := filepath.Join(source, body.Path)
		dst := filepath.Join(target, body.Path)

		_ = os.MkdirAll(target, 0755)
		_ = os.Rename(src, dst)

		c.Status(http.StatusOK)
	}
}
