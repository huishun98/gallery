package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPendingReturnsFirstFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(tmp, "a.jpg"), []byte("x"), 0644))
	assert.NoError(t, os.WriteFile(filepath.Join(tmp, "b.mp4"), []byte("x"), 0644))

	r := gin.New()
	r.GET("/media/pending", Pending(tmp))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/media/pending", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var got Media
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	assert.NotEmpty(t, got.Path)
	assert.NotEmpty(t, got.URL)
	assert.NotEmpty(t, got.Type)
}

func TestPendingNoFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()

	r := gin.New()
	r.GET("/media/pending", Pending(tmp))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/media/pending", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestMoveMediaMovesFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	src := t.TempDir()
	dst := t.TempDir()

	filename := "clip.mp4"
	assert.NoError(t, os.WriteFile(filepath.Join(src, filename), []byte("x"), 0644))

	r := gin.New()
	r.POST("/move", MoveMedia(src, dst))

	payload, _ := json.Marshal(map[string]string{"path": filename})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/move", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	_, err := os.Stat(filepath.Join(dst, filename))
	assert.NoError(t, err)
}

func TestMoveMediaBadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/move", MoveMedia("/tmp", "/tmp"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/move", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
