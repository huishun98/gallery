package handlers

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	defaultPage = 1
	defaultSize = 24
)

var imageExts = map[string]struct{}{
	".heic": {},
	".jpg":  {},
	".jpeg": {},
	".png":  {},
}

func MediaList(mediaDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr, hasPage := c.GetQuery("page")
		sizeStr, hasSize := c.GetQuery("size")

		var mediaFiles []string

		filepath.WalkDir(mediaDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".heic" || ext == ".jpg" || ext == ".png" || ext == ".jpeg" || ext == ".mp4" || ext == ".webm" || ext == ".mov" {
				mediaFiles = append(mediaFiles, "/mediafiles/"+filepath.Base(path))
			}
			return nil
		})

		total := len(mediaFiles)

		// Default: return everything as one page
		page := 1
		size := total

		if hasPage || hasSize {
			page = parsePositiveInt(pageStr, defaultPage)
			size = parsePositiveInt(sizeStr, defaultSize)
		}

		sort.Strings(mediaFiles)
		items := paginate(mediaFiles, page, size)

		c.JSON(http.StatusOK, gin.H{
			"items": items,
			"page":  page,
			"size":  size,
			"total": total,
		})
	}
}

func paginate(items []string, page, size int) []string {
	start := (page - 1) * size
	if start >= len(items) {
		return []string{}
	}

	end := start + size
	if end > len(items) {
		end = len(items)
	}

	return items[start:end]
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}
