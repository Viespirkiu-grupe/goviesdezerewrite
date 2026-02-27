package file

import (
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

		// Generate candidate paths for the file
		candidates := utils.GenerateCandidatePaths(basePath)

		// Local filesystem deletion logic
		var filePath string
		var fileInfo os.FileInfo

		// Check each candidate path for existence
		for _, candidate := range candidates {
			if info, err := os.Stat(candidate); err == nil {
				filePath = candidate
				fileInfo = info
				break
			}
		}

		// If no file found, return 404
		if filePath == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		// Delete the file
		if err := os.Remove(filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
			return
		}

		// Update usage
		utils.SetUsage(utils.GetUsage() - fileInfo.Size())

		c.JSON(http.StatusOK, gin.H{
			"deleted":   filepath.Base(filePath),
			"sizeFreed": fileInfo.Size(),
		})
	}

}
