package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestQRWhenURLNotReady(t *testing.T) {
	gin.SetMode(gin.TestMode)
	NgrokURL.Store("")

	r := gin.New()
	r.GET("/qr", QR)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/qr", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestQRReturnsPNG(t *testing.T) {
	gin.SetMode(gin.TestMode)
	NgrokURL.Store("https://example.com")
	defer NgrokURL.Store("")

	r := gin.New()
	r.GET("/qr", QR)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/qr", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
	assert.NotZero(t, w.Body.Len())
}
