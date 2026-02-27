package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

// ShardPath creates a sharded path by using the first two characters of the filename as a subdirectory
func ShardPath(filename, storagePath string) string {
	if len(filename) < 2 {
		return filepath.Join(storagePath, filename)
	}
	shard := filename[:2]
	return filepath.Join(storagePath, shard, filename)
}

// GenerateCandidatePaths generates an array of candidate paths for file lookup
func GenerateCandidatePaths(basePath string) []string {
	candidates := []string{basePath}

	// Add .jpeg/.jpg variants
	if strings.HasSuffix(strings.ToLower(basePath), ".jpeg") {
		candidates = append(candidates, strings.TrimSuffix(basePath, ".jpeg")+".jpg")
	} else if strings.HasSuffix(strings.ToLower(basePath), ".jpg") {
		candidates = append(candidates, strings.TrimSuffix(basePath, ".jpg")+".jpeg")
	}

	ext := filepath.Ext(basePath)

	// If the extension has uppercase letters, add lowercase variant
	if ext != "" && ext != strings.ToLower(ext) {
		lowerExtPath := strings.TrimSuffix(basePath, ext) + strings.ToLower(ext)
		candidates = append(candidates, lowerExtPath)
	}

	// Add extensionless version
	if ext != "" {
		withoutExt := strings.TrimSuffix(basePath, ext)
		candidates = append(candidates, withoutExt)
		candidates = append(candidates, withoutExt+".bin")
		candidates = append(candidates, withoutExt+".php")
		candidates = append(candidates, withoutExt+".null")
	}

	return candidates
}

// UsageData represents the structure of the usage.json file
type UsageData struct {
	TotalSize int64 `json:"totalSize"`
}

var totalSize int64

// GetUsage returns the current total disk usage in bytes
func GetUsage() int64 {
	return atomic.LoadInt64(&totalSize)
}

// SetUsage sets the total disk usage to a specific value
func SetUsage(size int64) error {
	atomic.SwapInt64(&totalSize, size)
	return saveUsage()
}

// AddUsage adds a specific value to the total disk usage
func AddUsage(size int64) error {
	atomic.AddInt64(&totalSize, size)
	return saveUsage()
}

// LoadUsage loads disk usage from the usage.json file
func LoadUsage() error {
	data, err := os.ReadFile("./usage.json")
	if err != nil {
		if os.IsNotExist(err) {
			totalSize = 0
			return nil
		}
		return err
	}

	var usageData UsageData
	if err := json.Unmarshal(data, &usageData); err != nil {
		return err
	}

	atomic.SwapInt64(&totalSize, usageData.TotalSize)
	return nil
}

// saveUsage saves the current totalSize to the usage.json file
func saveUsage() error {
	usageData := UsageData{TotalSize: atomic.LoadInt64(&totalSize)}
	data, err := json.Marshal(usageData)
	if err != nil {
		return err
	}
	return os.WriteFile("./usage.json", data, 0644)
}
