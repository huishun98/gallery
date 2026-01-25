package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Page(templateName string, data any) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, templateName, data)
	}
}
