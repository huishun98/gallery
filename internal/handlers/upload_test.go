package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUploadMissingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()

	r := gin.New()
	r.POST("/upload", Upload(tmp))

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadSavesFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()

	r := gin.New()
	r.POST("/upload", Upload(tmp))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("picture", "photo.jpg")
	if !assert.NoError(t, err) {
		return
	}
	_, err = part.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	entries, err := os.ReadDir(tmp)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
}

func TestUploadRejectsLongVideo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()
	binDir := t.TempDir()
	ffprobe := filepath.Join(binDir, "ffprobe")
	assert.NoError(t, os.WriteFile(ffprobe, []byte("#!/bin/sh\necho 45\n"), 0755))
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	r := gin.New()
	r.POST("/upload", Upload(tmp))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("picture", "clip.mp4")
	if !assert.NoError(t, err) {
		return
	}
	_, err = part.Write([]byte("fake"))
	assert.NoError(t, err)
	assert.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	entries, _ := os.ReadDir(tmp)
	assert.Len(t, entries, 0)
}

func TestVideoDurationSeconds(t *testing.T) {
	binDir := t.TempDir()
	ffprobe := filepath.Join(binDir, "ffprobe")
	assert.NoError(t, os.WriteFile(ffprobe, []byte("#!/bin/sh\necho 12.5\n"), 0755))
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	dur, err := videoDurationSeconds("/tmp/does-not-matter")
	assert.NoError(t, err)
	assert.Equal(t, 12.5, dur)
}
