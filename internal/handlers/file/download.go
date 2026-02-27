package file

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"goviesdeze/internal/adapters/archive/ziparchive"
	"goviesdeze/internal/config"
	"goviesdeze/internal/core/archivequery"
	"goviesdeze/internal/core/fileid"
	"goviesdeze/internal/core/filequery"
	"goviesdeze/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/h2non/filetype"
)

// GetFile handles file downloads with optional extraction and range support
func GetFile(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := strings.TrimSpace(c.Param("filename"))
		if !fileid.IsNumericOrMD5(filename) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a number or MD5"})
			return
		}

		extractTarget := strings.TrimSpace(c.Query("extract"))
		convertTo := strings.TrimSpace(c.Query("convertTo"))
		if convertTo != "" {
			extractTarget = convertTo
		}
		basePath := utils.ShardPath(filename, cfg.StoragePath)

		// Find the local file
		candidates := utils.GenerateCandidatePaths(basePath)
		var filePath string
		var fileInfo os.FileInfo

		for _, candidate := range candidates {
			// Log the candidate being checked
			log.Printf("checking candidate path: %s", candidate)
			if info, err := os.Stat(candidate); err == nil {
				filePath = candidate
				fileInfo = info
				break
			}
		}

		if filePath == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		var rdr io.Reader
		var size int64
		var contentType string

		if extractTarget == "" {
			// Normal file serving
			f, err := os.Open(filePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
				return
			}
			defer f.Close()

			rdr = f
			size = fileInfo.Size()
			contentType = getContentType(filePath)
		} else {
			// Extraction requested
			buf, err := os.ReadFile(filePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read archive"})
				return
			}

			extracted, best, err := archivequery.ReadBestMatch(
				c.Request.Context(),
				ziparchive.Service{},
				buf,
				extractTarget,
				filequery.SimilarityFunc(utils.Similarity),
			)
			if err != nil {
				switch {
				case errors.Is(err, archivequery.ErrInvalidArchive):
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid archive"})
				case errors.Is(err, archivequery.ErrFileNotFound):
					c.JSON(http.StatusNotFound, gin.H{"error": "File not found in archive"})
				default:
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error extracting file"})
				}
				return
			}

			// Wrap in a bytes.Reader so it implements both io.Reader and io.Seeker
			rdr = bytes.NewReader(extracted)
			size = int64(len(extracted)) // size is just the buffer length
			contentType = getContentType(best)
		}

		// Range support
		rangeHeader := c.GetHeader("Range")
		if rangeHeader != "" {
			start, end, err := filequery.ParseRange(rangeHeader, size)
			if err != nil {
				c.Header("Content-Range", fmt.Sprintf("bytes */%d", size))
				c.Status(http.StatusRequestedRangeNotSatisfiable)
				return
			}

			var limitedReader io.Reader
			if seeker, ok := rdr.(io.Seeker); ok && rdr != nil {
				if _, err := seeker.Seek(start, io.SeekStart); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seek file"})
					return
				}
				limitedReader = io.LimitReader(rdr, end-start+1)
			} else {
				// For non-seekable readers (rare), read into buffer
				buf, _ := io.ReadAll(rdr)
				limitedReader = bytes.NewReader(buf[start : end+1])
			}

			c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
			c.Header("Accept-Ranges", "bytes")
			c.Header("Content-Length", strconv.FormatInt(end-start+1, 10))
			c.Header("Content-Type", contentType)
			c.Status(http.StatusPartialContent)
			if _, err := io.Copy(c.Writer, limitedReader); err != nil {
				log.Printf("copy range response failed: %v", err)
			}
			return
		}

		// Full content
		c.Header("Content-Length", strconv.FormatInt(size, 10))
		c.Header("Content-Type", contentType)
		c.Header("Accept-Ranges", "bytes")
		c.Status(http.StatusOK)
		if _, err := io.Copy(c.Writer, rdr); err != nil {
			log.Printf("copy full response failed: %v", err)
		}
	}
}

// getContentType determines the content type based on file extension
func getContentType(filePath string) string {
	kind, _ := filetype.MatchFile(filePath)
	if kind != filetype.Unknown {
		return kind.MIME.Value
	}
	return "application/octet-stream"
}

func writeResponse(w http.ResponseWriter, r *http.Request, rdr io.ReadCloser, upRes *http.Response, name string) bool {
	defer rdr.Close()

	// Determine content type from extension
	contentType := contentTypeFromExt(name)

	w.Header().Set("Content-Type", contentType)

	// Content-Disposition: inline; filename*=UTF-8''...
	if name != "" {
		nameOnly := path.Base(name) // get the last part of the path
		w.Header().Set("Content-Disposition",
			fmt.Sprintf("inline; filename*=UTF-8''%s", url.PathEscape(nameOnly)))
	}

	// Cache-Control: same as before
	w.Header().Set("Cache-Control", "public, max-age=2592000, immutable")

	// Forward byte ranges if upstream provided (optional)
	if rng := upRes.Header.Get("Accept-Ranges"); rng != "" {
		w.Header().Set("Accept-Ranges", rng)
	}
	if cr := upRes.Header.Get("Content-Range"); cr != "" {
		w.Header().Set("Content-Range", cr)
	}

	// Status code
	w.WriteHeader(upRes.StatusCode)

	// Copy body
	if _, err := io.Copy(w, rdr); err != nil {
		log.Printf("writing response body error: %v", err)
		return true
	}

	return false
}

func contentTypeFromExt(filename string) string {
	ext := filepath.Ext(filename)
	if ext != "" {
		if ct := mime.TypeByExtension(ext); ct != "" {
			return ct
		}
	}
	return "application/octet-stream"
}
