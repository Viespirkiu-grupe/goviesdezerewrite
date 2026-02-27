package file

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"goviesdeze/internal/config"
	"goviesdeze/internal/utils"

	"github.com/gin-gonic/gin"
)

// UploadFile handles file uploads for both local filesystem and S3
func UploadFile(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.Param("filename")
		filePath := utils.ShardPath(filename, cfg.StoragePath)
		var existingSize int64

		// Local filesystem upload logic
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
			return
		}

		// Check if file exists
		if stat, err := os.Stat(filePath); err == nil {
			existingSize = stat.Size()
		}

		// Create file
		file, err := os.Create(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
			return
		}
		defer file.Close()

		// Copy request body to file
		byteCount, err := io.Copy(file, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
			return
		}

		totalSize := utils.GetUsage() - existingSize + byteCount
		utils.SetUsage(totalSize)

		c.JSON(http.StatusOK, gin.H{
			"uploaded":  filename,
			"replaced":  existingSize > 0,
			"oldSize":   existingSize,
			"newSize":   byteCount,
			"totalSize": totalSize,
		})
	}

}
