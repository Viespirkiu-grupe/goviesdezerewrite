package handlers

import (
	"goviesdeze/internal/config"
	"goviesdeze/internal/handlers/file"
	"goviesdeze/internal/handlers/storage"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all the routes for the application
func RegisterRoutes(router *gin.Engine, cfg *config.Config) {
	// Storage usage endpoint
	router.GET("/storage-usage", storage.GetStorageUsage)

	// File operations
	router.PUT("/file/:filename", file.UploadFile(cfg))
	router.GET("/file/:filename", file.GetFile(cfg))
	router.DELETE("/file/:filename", file.DeleteFile(cfg))

	// Download URL endpoint
	router.POST("/download-url", file.DownloadURL(cfg))
}
