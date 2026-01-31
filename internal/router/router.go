package router

import (
	"database/sql"
	"gallery/internal/handlers"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, db *sql.DB, dataDir string, adminAccount gin.Accounts, approvalsEnabled, danmuEnabled bool) {
	mediaDir := filepath.Join(dataDir, "media")
	uploadDir := filepath.Join(mediaDir, "media")
	templateData := gin.H{"DanmuEnabled": danmuEnabled, "ApprovalsEnabled": approvalsEnabled}

	var admin *gin.RouterGroup
	if approvalsEnabled || danmuEnabled {
		admin = r.Group("/admin", gin.BasicAuth(adminAccount))
	}

	if approvalsEnabled {
		admin.GET("/review", handlers.Page("decision.html", templateData))
		admin.GET("/media/pending", handlers.Pending(filepath.Join(mediaDir, "pending")))
		admin.POST("/media/approve", handlers.MoveMedia(filepath.Join(mediaDir, "pending"), filepath.Join(mediaDir, "media")))
		admin.POST("/media/reject", handlers.MoveMedia(filepath.Join(mediaDir, "pending"), filepath.Join(mediaDir, "rejected")))

		uploadDir = filepath.Join(mediaDir, "pending")
	}

	r.GET("/", handlers.Page("upload.html", templateData))
	r.POST("/upload", handlers.Upload(uploadDir))
	r.GET("/slideshow", handlers.Page("slideshow.html", templateData))
	r.GET("/media", handlers.MediaList(filepath.Join(mediaDir, "media")))
	r.GET("/qr", handlers.QR)

	if danmuEnabled {
		admin.GET("/comments", handlers.Page("comments.html", templateData))
		admin.POST("/comments/delete", handlers.DeleteComment(db))
		admin.GET("/comments/download", handlers.DownloadComments(db))

		r.GET("/comments", handlers.GetComments(db))
		r.POST("/comment", handlers.SaveComment(db))
	}

	r.Static("/mediafiles", filepath.Join(mediaDir, "media"))
	r.Static("/files", filepath.Join(mediaDir, "pending"))
}
