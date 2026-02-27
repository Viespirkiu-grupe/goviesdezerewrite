package file

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"syscall"

	"goviesdeze/internal/config"
	"goviesdeze/internal/utils"

	"github.com/gin-gonic/gin"
)

// DownloadURLRequest represents the request body for download-url endpoint
type DownloadURLRequest struct {
	URL string `json:"url" binding:"required"`
}

// DownloadURL handles downloading files from URLs and storing them
func DownloadURL(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DownloadURLRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			if cfg.AppDebug {
				log.Printf("debug download-url invalid payload: err=%v", err)
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing url field"})
			return
		}
		if cfg.AppDebug {
			log.Printf("debug download-url start: url=%q", req.URL)
		}

		// Download the file from URL
		resp, err := http.Get(req.URL)
		if err != nil {
			if cfg.AppDebug {
				log.Printf("debug download-url fetch failed: url=%q err=%v", req.URL, err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch URL"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if cfg.AppDebug {
				log.Printf("debug download-url non-200: url=%q status=%d", req.URL, resp.StatusCode)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch %s: %s", req.URL, resp.Status)})
			return
		}

		// Local filesystem storage logic
		if err := os.MkdirAll(cfg.StoragePath, 0o755); err != nil {
			if cfg.AppDebug {
				log.Printf("debug download-url storage mkdir failed: path=%q err=%v", cfg.StoragePath, err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create storage directory"})
			return
		}

		// Create temporary file inside storage path to avoid cross-device rename issues.
		tmpFile, err := os.CreateTemp(cfg.StoragePath, "tmp_*")
		if err != nil {
			if cfg.AppDebug {
				log.Printf("debug download-url temp create failed: path=%q err=%v", cfg.StoragePath, err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary file"})
			return
		}
		defer os.Remove(tmpFile.Name())

		// Calculate MD5 hash while writing
		hash := md5.New()
		multiWriter := io.MultiWriter(tmpFile, hash)

		if _, err := io.Copy(multiWriter, resp.Body); err != nil {
			tmpFile.Close()
			if cfg.AppDebug {
				log.Printf("debug download-url copy failed: url=%q err=%v", req.URL, err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to temporary file"})
			return
		}
		tmpFile.Close()

		md5sum := fmt.Sprintf("%x", hash.Sum(nil))
		filename := md5sum
		finalPath := utils.ShardPath(filename, cfg.StoragePath)
		if cfg.AppDebug {
			log.Printf("debug download-url digest computed: md5=%s temp=%q final=%q", md5sum, tmpFile.Name(), finalPath)
		}

		// Check if file already exists
		if _, err := os.Stat(finalPath); err == nil {
			// File already exists, remove temp file
			os.Remove(tmpFile.Name())
			if stat, err := os.Stat(finalPath); err == nil {
				if cfg.AppDebug {
					log.Printf("debug download-url deduplicated existing file: md5=%s size=%d", md5sum, stat.Size())
				}
				c.JSON(http.StatusOK, gin.H{
					"md5":  md5sum,
					"size": stat.Size(),
				})
				return
			}
		}

		// Create directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
			os.Remove(tmpFile.Name())
			if cfg.AppDebug {
				log.Printf("debug download-url final dir mkdir failed: dir=%q err=%v", filepath.Dir(finalPath), err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
			return
		}

		// Move temp file to final location
		if err := moveFile(tmpFile.Name(), finalPath); err != nil {
			os.Remove(tmpFile.Name())
			if cfg.AppDebug {
				log.Printf("debug download-url move failed: src=%q dst=%q err=%v", tmpFile.Name(), finalPath, err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move file"})
			return
		}

		// Update usage
		if stat, err := os.Stat(finalPath); err == nil {
			// utils.SetUsage(utils.GetUsage() + stat.Size())
			if err := utils.AddUsage(stat.Size()); err != nil {
				if cfg.AppDebug {
					log.Printf("debug download-url usage update failed: md5=%s err=%v", md5sum, err)
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update usage"})
				return
			}
			if cfg.AppDebug {
				log.Printf("debug download-url success: md5=%s size=%d", md5sum, stat.Size())
			}
			c.JSON(http.StatusOK, gin.H{
				"md5":  md5sum,
				"size": stat.Size(),
			})
		} else {
			if cfg.AppDebug {
				log.Printf("debug download-url final stat failed: path=%q err=%v", finalPath, err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file stats"})
		}
	}

}

func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	} else if !errors.Is(err, syscall.EXDEV) {
		return err
	}

	if err := copyFile(src, dst); err != nil {
		return err
	}

	return os.Remove(src)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}
