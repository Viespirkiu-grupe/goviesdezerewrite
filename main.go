package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"goviesdeze/internal/config"
	"goviesdeze/internal/handlers"
	"goviesdeze/internal/middleware"
	"goviesdeze/internal/utils"

	"github.com/gin-gonic/gin"
)

func cleanTmp() {
	for {
		cmd := exec.Command("find", "/tmp", "-mindepth", "1", "-mmin", "+5", "-exec", "rm", "-rf", "{}", "+")
		if err := cmd.Run(); err != nil {
			log.Println("Error running find:", err)
		}
		time.Sleep(1 * time.Minute)
	}
}

func main() {
	go cleanTmp() // runs in background

	// Load configuration
	cfg := config.Load()
	if cfg.AppDebug {
		log.Printf("APP_DEBUG enabled: verbose logging is active")
		log.Printf("config: port=%s storage_path=%s require_api_key=%v", cfg.Port, cfg.StoragePath, cfg.RequireAPIKey)
	}

	// Initialize storage usage
	if err := utils.LoadUsage(); err != nil {
		log.Printf("Warning: Failed to load usage: %v", err)
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(cfg.StoragePath, 0755); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	// Setup Gin router
	router := gin.Default()

	// Add middleware
	router.Use(middleware.RequestLogger(cfg.AppDebug))
	if cfg.RequireAPIKey {
		router.Use(middleware.APIKeyAuth(cfg.APIKey, cfg.AppDebug))
	}

	// Register routes
	handlers.RegisterRoutes(router, cfg)

	// Start server
	log.Printf("Server running on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
