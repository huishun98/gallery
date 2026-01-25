package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPageRendersTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	tmpl := template.Must(template.New("test.html").Parse("OK"))
	r.SetHTMLTemplate(tmpl)
	r.GET("/", Page("test.html", nil))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}
