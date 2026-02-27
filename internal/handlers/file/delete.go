package file

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"goviesdeze/internal/config"
	"goviesdeze/internal/utils"

	"github.com/gin-gonic/gin"
)

// DeleteFile handles file deletion for both local filesystem and S3
func DeleteFile(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.Param("filename")
		basePath := utils.ShardPath(filename, cfg.StoragePath)
		if cfg.AppDebug {
			log.Printf("debug delete start: filename=%q base_path=%q", filename, basePath)
		}

		// Generate candidate paths for the file
		candidates := utils.GenerateCandidatePaths(basePath)

		// Local filesystem deletion logic
		var filePath string
		var fileInfo os.FileInfo

		// Check each candidate path for existence
		for _, candidate := range candidates {
			if cfg.AppDebug {
				log.Printf("debug delete candidate: %q", candidate)
			}
			if info, err := os.Stat(candidate); err == nil {
				filePath = candidate
				fileInfo = info
				break
			}
		}

		// If no file found, return 404
		if filePath == "" {
			if cfg.AppDebug {
				log.Printf("debug delete not found: filename=%q", filename)
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		// Delete the file
		if err := os.Remove(filePath); err != nil {
			if cfg.AppDebug {
				log.Printf("debug delete remove failed: path=%q err=%v", filePath, err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
			return
		}

		// Update usage
		if err := utils.SetUsage(utils.GetUsage() - fileInfo.Size()); err != nil {
			if cfg.AppDebug {
				log.Printf("debug delete usage update failed: err=%v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update usage"})
			return
		}
		if cfg.AppDebug {
			log.Printf("debug delete success: path=%q size_freed=%d", filePath, fileInfo.Size())
		}

		c.JSON(http.StatusOK, gin.H{
			"deleted":   filepath.Base(filePath),
			"sizeFreed": fileInfo.Size(),
		})
	}

}
