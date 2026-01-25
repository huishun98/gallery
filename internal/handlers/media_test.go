package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mediaResponse struct {
	Items []string `json:"items"`
	Page  int      `json:"page"`
	Size  int      `json:"size"`
	Total int      `json:"total"`
}

func TestMediaListPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()

	files := []string{"b.jpg", "a.png", "c.mp4", "skip.txt"}
	for _, name := range files {
		assert.NoError(t, os.WriteFile(filepath.Join(tmp, name), []byte("x"), 0644))
	}

	r := gin.New()
	r.GET("/media", MediaList(tmp))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/media?page=1&size=2", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp mediaResponse
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, 3, resp.Total)
	assert.Len(t, resp.Items, 2)
	assert.Equal(t, "/mediafiles/a.png", resp.Items[0])
}

func TestParsePositiveInt(t *testing.T) {
	assert.Equal(t, 7, parsePositiveInt("-1", 7))
	assert.Equal(t, 9, parsePositiveInt("abc", 9))
	assert.Equal(t, 4, parsePositiveInt("4", 1))
}
