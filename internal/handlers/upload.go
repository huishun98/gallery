package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gallery/internal/binary"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Upload(dir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("picture")
		if err != nil {
			c.String(http.StatusBadRequest, "No file uploaded")
			return
		}

		ts := time.Now().Format("20060102-150405")
		id := uuid.New().String()
		ext := filepath.Ext(file.Filename)

		filename := ts + "_" + id + ext
		dst := filepath.Join(dir, filename)

		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.String(http.StatusInternalServerError, "Failed to save file")
			return
		}

		if ext == ".mp4" || ext == ".webm" || ext == ".mov" {
			seconds, err := videoDurationSeconds(dst)
			if err != nil {
				os.Remove(dst)
				c.String(http.StatusBadRequest, "Unable to read video metadata, %v", err)
				return
			}

			if seconds > 30 {
				os.Remove(dst)
				c.String(http.StatusBadRequest, "Video should not exceed 30 seconds")
				return
			}
		}

		c.String(http.StatusOK, "Uploaded successfully")
	}
}

func videoDurationSeconds(path string) (float64, error) {

	ffprobePath, _, err := binary.WhereIs("ffprobe")
	if err != nil {
		return 0, err
	}

	cmd := exec.Command(
		ffprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	var errOut bytes.Buffer
	cmd.Stderr = &errOut

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errOut.String())
		if msg != "" {
			return 0, fmt.Errorf("ffprobe failed: %w: %s", err, msg)
		}
		return 0, err
	}

	return strconv.ParseFloat(strings.TrimSpace(out.String()), 64)
}
