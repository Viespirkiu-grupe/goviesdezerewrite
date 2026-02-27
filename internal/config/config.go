package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	StoragePath   string
	APIKey        string
	RequireAPIKey bool
	AppDebug      bool
}

func Load() *Config {
	cfg := &Config{
		Port:          getEnv("PORT", "3000"),
		StoragePath:   getEnv("STORAGE_PATH", "/storage"),
		APIKey:        getEnv("API_KEY", "super-secret-key"),
		RequireAPIKey: getEnvBool("REQUIRE_API_KEY", true),
		AppDebug:      getEnvBool("APP_DEBUG", false),
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
