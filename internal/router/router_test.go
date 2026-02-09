package router

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetupRoutesWithoutApprovals(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	SetupRoutes(r, nil, "/tmp", gin.Accounts{"user": "pass"}, false, true, true)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/review", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSetupRoutesUploadsDisabledUnregistersUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	SetupRoutes(r, nil, "/tmp", gin.Accounts{"user": "pass"}, false, true, false)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("test"))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSetupRoutesTemplateDataIncludesMediaDirAndUploadsEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.SetHTMLTemplate(template.Must(template.New("slideshow.html").Parse(`{{.MediaDir}} {{printf "UploadsEnabled=%v" .UploadsEnabled}}`)))

	dataDir := "/tmp/gallery"
	SetupRoutes(r, nil, dataDir, gin.Accounts{"user": "pass"}, false, true, false)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/slideshow", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, dataDir+"/media/media")
	assert.Contains(t, body, "UploadsEnabled=false")
}

func TestSetupRoutesWithAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	SetupRoutes(r, nil, "/tmp", gin.Accounts{"user": "pass"}, true, true, true)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/review", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/comments", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSetupRoutesWithoutDanmu(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	SetupRoutes(r, nil, "/tmp", gin.Accounts{"user": "pass"}, true, false, true)

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

	SetupRoutes(r, nil, "/tmp", gin.Accounts{"user": "pass"}, true, false, true)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/comments", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/admin/comments/delete", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
