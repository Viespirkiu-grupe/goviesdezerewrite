package storage

import (
	"net/http"

	"goviesdeze/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetStorageUsage returns the current storage usage
func GetStorageUsage(c *gin.Context) {
	totalSize := utils.GetUsage()
	c.JSON(http.StatusOK, gin.H{
		"totalSizeBytes": totalSize,
	})
}
