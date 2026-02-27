package file

import (
	"io"
	"log"
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
		if cfg.AppDebug {
			log.Printf("debug upload start: filename=%q path=%q", filename, filePath)
		}

		// Local filesystem upload logic
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			if cfg.AppDebug {
				log.Printf("debug upload mkdir failed: path=%q err=%v", filepath.Dir(filePath), err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
			return
		}

		// Check if file exists
		if stat, err := os.Stat(filePath); err == nil {
			existingSize = stat.Size()
			if cfg.AppDebug {
				log.Printf("debug upload existing file: size=%d", existingSize)
			}
		}

		// Create file
		file, err := os.Create(filePath)
		if err != nil {
			if cfg.AppDebug {
				log.Printf("debug upload create failed: path=%q err=%v", filePath, err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
			return
		}
		defer file.Close()

		// Copy request body to file
		byteCount, err := io.Copy(file, c.Request.Body)
		if err != nil {
			if cfg.AppDebug {
				log.Printf("debug upload copy failed: filename=%q err=%v", filename, err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
			return
		}

		totalSize := utils.GetUsage() - existingSize + byteCount
		if err := utils.SetUsage(totalSize); err != nil {
			if cfg.AppDebug {
				log.Printf("debug upload usage update failed: err=%v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update usage"})
			return
		}
		if cfg.AppDebug {
			log.Printf("debug upload success: filename=%q replaced=%v old_size=%d new_size=%d total=%d", filename, existingSize > 0, existingSize, byteCount, totalSize)
		}

		c.JSON(http.StatusOK, gin.H{
			"uploaded":  filename,
			"replaced":  existingSize > 0,
			"oldSize":   existingSize,
			"newSize":   byteCount,
			"totalSize": totalSize,
		})
	}

}
