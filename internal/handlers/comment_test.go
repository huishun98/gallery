package handlers

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

type commentsResponse struct {
	Filename string    `json:"filename"`
	Comments []Comment `json:"comments"`
	Count    int       `json:"count"`
	Page     int       `json:"page"`
	Size     int       `json:"size"`
	Total    int       `json:"total"`
}

func TestSaveCommentAndGetComments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	r := gin.New()
	r.POST("/comment", SaveComment(db))
	r.GET("/comments", GetComments(db))

	form := url.Values{}
	form.Set("filename", "pic.jpg")
	form.Set("comment", "nice")

	req := httptest.NewRequest(http.MethodPost, "/comment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/comments?filename=pic.jpg", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp commentsResponse
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, 1, resp.Count)
	assert.Equal(t, 1, resp.Total)
}

func TestSaveCommentValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	r := gin.New()
	r.POST("/comment", SaveComment(db))

	req := httptest.NewRequest(http.MethodPost, "/comment", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCommentsAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	r := gin.New()
	r.GET("/comments", GetComments(db))

	_, err := db.Exec(`INSERT INTO comments (filename, comment) VALUES (?, ?)`, "pic1.jpg", "a")
	assert.NoError(t, err)
	_, err = db.Exec(`INSERT INTO comments (filename, comment) VALUES (?, ?)`, "pic2.jpg", "b")
	assert.NoError(t, err)
	_, err = db.Exec(`INSERT INTO comments (filename, comment, removed_at) VALUES (?, ?, CURRENT_TIMESTAMP)`, "pic3.jpg", "c")
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/comments", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp commentsResponse
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, 2, resp.Total)
	assert.Equal(t, 2, resp.Count)
	assert.Equal(t, 1, resp.Page)
}

func TestDeleteComment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	res, err := db.Exec(`INSERT INTO comments (filename, comment) VALUES (?, ?)`, "pic.jpg", "to delete")
	assert.NoError(t, err)
	id, _ := res.LastInsertId()

	r := gin.New()
	r.POST("/comments/delete", DeleteComment(db))

	form := url.Values{}
	form.Set("id", strconv.FormatInt(id, 10))

	req := httptest.NewRequest(http.MethodPost, "/comments/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM comments WHERE id = ? AND removed_at IS NULL`, id).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDownloadComments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	_, err := db.Exec(`INSERT INTO comments (filename, comment) VALUES (?, ?)`, "pic1.jpg", "first")
	assert.NoError(t, err)
	_, err = db.Exec(`INSERT INTO comments (filename, comment) VALUES (?, ?)`, "pic2.jpg", "second")
	assert.NoError(t, err)
	_, err = db.Exec(`INSERT INTO comments (filename, comment, removed_at) VALUES (?, ?, CURRENT_TIMESTAMP)`, "pic3.jpg", "removed")
	assert.NoError(t, err)

	r := gin.New()
	r.GET("/comments/download", DownloadComments(db))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/comments/download", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/csv")

	reader := csv.NewReader(w.Body)
	rows, err := reader.ReadAll()
	assert.NoError(t, err)

	if assert.GreaterOrEqual(t, len(rows), 1) {
		assert.Equal(t, []string{"id", "filename", "comment", "created_at", "removed_at"}, rows[0])
	}
	assert.Equal(t, 4, len(rows))
}

func testDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if !assert.NoError(t, err) {
		return nil
	}
	_, err = db.Exec(`
		CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			filename TEXT NOT NULL,
			comment TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			removed_at DATETIME
		)
	`)
	if !assert.NoError(t, err) {
		_ = db.Close()
	}
	return db
}
