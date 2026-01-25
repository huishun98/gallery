package handlers

import (
	"database/sql"
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Comment struct {
	ID        int64  `json:"id"`
	Filename  string `json:"filename"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"created_at"`
}

func SaveComment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.PostForm("filename")
		comment := strings.TrimSpace(c.PostForm("comment"))

		if filename == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
			return
		}
		if comment == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "comment is required"})
			return
		}

		// Insert into SQLite
		res, err := db.Exec(
			`INSERT INTO comments (filename, comment) VALUES (?, ?)`,
			filename, comment,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save comment"})
			return
		}

		id, _ := res.LastInsertId()

		c.JSON(http.StatusOK, gin.H{
			"id":       id,
			"filename": filename,
			"comment":  comment,
		})
	}
}

func GetComments(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := strings.TrimSpace(c.Query("filename"))
		page := parsePositiveInt(c.Query("page"), 1)
		size := parsePositiveInt(c.Query("size"), 25)
		if size > 100 {
			size = 100
		}
		offset := (page - 1) * size

		where := "removed_at IS NULL"
		args := []any{}

		if filename != "" {
			where += " AND filename = ?"
			args = append(args, filename)
		}

		// count
		var total int
		err := db.QueryRow(
			"SELECT COUNT(*) FROM comments WHERE "+where,
			args...,
		).Scan(&total)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query comments"})
			return
		}

		// data
		query := `
			SELECT id, filename, comment, created_at
			FROM comments
			WHERE ` + where + `
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`
		rows, err := db.Query(query, append(args, size, offset)...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query comments"})
			return
		}
		defer rows.Close()

		comments := []Comment{}
		for rows.Next() {
			var cm Comment
			if err := rows.Scan(&cm.ID, &cm.Filename, &cm.Comment, &cm.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read comment"})
				return
			}
			comments = append(comments, cm)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read comments"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"filename": filename,
			"comments": comments,
			"count":    len(comments),
			"page":     page,
			"size":     size,
			"total":    total,
		})
	}
}

func DeleteComment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := strings.TrimSpace(c.PostForm("id"))
		if idStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		res, err := db.Exec(`UPDATE comments SET removed_at = CURRENT_TIMESTAMP WHERE id = ? AND removed_at IS NULL`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete comment"})
			return
		}

		rows, err := res.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete comment"})
			return
		}
		if rows == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"deleted": id})
	}
}

func DownloadComments(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		var (
			rows *sql.Rows
			err  error
		)

		rows, err = db.Query(
			`SELECT id, filename, comment, created_at, removed_at
			FROM comments
			ORDER BY created_at DESC`,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query comments"})
			return
		}
		defer rows.Close()

		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=comments.csv")

		writer := csv.NewWriter(c.Writer)
		if err := writer.Write([]string{"id", "filename", "comment", "created_at", "removed_at"}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write csv"})
			return
		}

		for rows.Next() {
			var id int64
			var name string
			var comment string
			var createdAt string
			var removedAt sql.NullString
			if err := rows.Scan(&id, &name, &comment, &createdAt, &removedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read comment"})
				return
			}
			removedValue := ""
			if removedAt.Valid {
				removedValue = removedAt.String
			}
			if err := writer.Write([]string{
				strconv.FormatInt(id, 10),
				name,
				comment,
				createdAt,
				removedValue,
			}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write csv"})
				return
			}
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read comments"})
			return
		}

		writer.Flush()
		if err := writer.Error(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write csv"})
			return
		}
	}
}
