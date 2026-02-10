package router

import (
	"database/sql"
	"gallery/internal/handlers"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, db *sql.DB, dataDir string, adminAccount gin.Accounts, approvalsEnabled, danmuEnabled, uploadsEnabled bool) {
	mediaDir := filepath.Join(dataDir, "media")

	approvedDir := filepath.Join(mediaDir, "media")
	pendingDir := filepath.Join(mediaDir, "pending")
	rejectedDir := filepath.Join(mediaDir, "rejected")

	uploadDir := approvedDir
	templateData := gin.H{
		"DanmuEnabled":     danmuEnabled,
		"ApprovalsEnabled": approvalsEnabled,
		"UploadsEnabled":   uploadsEnabled,
		"MediaDir":         approvedDir,
		"Year":             time.Now().Year(),
	}

	var admin *gin.RouterGroup
	if approvalsEnabled || danmuEnabled {
		admin = r.Group("/admin", gin.BasicAuth(adminAccount))
	}

	if approvalsEnabled {
		admin.GET("/review", handlers.Page("decision.html", templateData))
		admin.GET("/media/pending", handlers.Pending(pendingDir))
		admin.POST("/media/approve", handlers.MoveMedia(pendingDir, approvedDir))
		admin.POST("/media/reject", handlers.MoveMedia(pendingDir, rejectedDir))

		uploadDir = pendingDir
	}

	r.GET("/", handlers.Page("upload.html", templateData))
	if uploadsEnabled {
		r.POST("/upload", handlers.Upload(uploadDir))
	}
	r.GET("/slideshow", handlers.Page("slideshow.html", templateData))
	r.GET("/media", handlers.MediaList(approvedDir))
	r.GET("/qr", handlers.QR)

	if danmuEnabled {
		admin.GET("/comments", handlers.Page("comments.html", templateData))
		admin.POST("/comments/delete", handlers.DeleteComment(db))
		admin.GET("/comments/download", handlers.DownloadComments(db))

		r.GET("/comments", handlers.GetComments(db))
		r.POST("/comment", handlers.SaveComment(db))
	}

	r.Static("/mediafiles", approvedDir)
	if approvalsEnabled {
		r.Static("/files", pendingDir)
	}
}
