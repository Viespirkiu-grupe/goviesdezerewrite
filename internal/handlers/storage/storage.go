package storage

import (
	"log"
	"net/http"

	"goviesdeze/internal/config"
	"goviesdeze/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetStorageUsage returns the current storage usage
func GetStorageUsage(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		totalSize := utils.GetUsage()
		if cfg.AppDebug {
			log.Printf("debug storage-usage: total_bytes=%d", totalSize)
		}

		c.JSON(http.StatusOK, gin.H{
			"totalSizeBytes": totalSize,
		})
	}
}
