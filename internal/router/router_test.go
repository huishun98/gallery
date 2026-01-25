package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetupRoutesWithoutAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	SetupRoutes(r, nil, "/tmp", nil, true)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSetupRoutesWithAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	SetupRoutes(r, nil, "/tmp", gin.Accounts{"user": "pass"}, true)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSetupRoutesWithoutDanmu(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	SetupRoutes(r, nil, "/tmp", nil, false)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/comment", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/comments", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSetupRoutesWithAdminWithoutDanmu(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	SetupRoutes(r, nil, "/tmp", gin.Accounts{"user": "pass"}, false)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/comments", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/admin/comments/delete", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
