package handlers

import (
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

var NgrokURL atomic.Value

func QR(c *gin.Context) {
	url, ok := NgrokURL.Load().(string)
	if !ok || url == "" {
		c.JSON(500, gin.H{"error": "public URL not ready yet"})
		return
	}
	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate QR"})
		return
	}
	c.Data(http.StatusOK, "image/png", png)
}
