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

	"goviesdeze/internal/config"
	"goviesdeze/internal/utils"
	"goviesdeze/internal/ziputil"

	"github.com/gin-gonic/gin"
	"github.com/h2non/filetype"
)

// GetFile handles file downloads with optional extraction and range support
func GetFile(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.Param("filename")
		extractTarget := c.Query("extract") // ?extract=/path/to/file
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

			files, err := ziputil.IdentityFilesV2(c.Request.Context(), buf)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid archive"})
				return
			}

			best, err := bestMatch(extractTarget, files)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "File not found in archive"})
				return
			}

			fileReader, err := ziputil.GetFileFromArchiveV2(c.Request.Context(), buf, best)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error extracting file"})
				return
			}
			defer fileReader.Close()

			// Wrap in a bytes.Reader so it implements both io.Reader and io.Seeker
			buf2, _ := io.ReadAll(fileReader)
			rdr = bytes.NewReader(buf2)
			size = int64(len(buf2)) // size is just the buffer length
			contentType = getContentType(best)
		}

		// Range support
		rangeHeader := c.GetHeader("Range")
		if rangeHeader != "" {
			start, end, err := parseRange(rangeHeader, size)
			if err != nil {
				c.Header("Content-Range", fmt.Sprintf("bytes */%d", size))
				c.Status(http.StatusRequestedRangeNotSatisfiable)
				return
			}

			var limitedReader io.Reader
			if seeker, ok := rdr.(io.Seeker); ok && rdr != nil {
				// rdr is already a *bytes.Reader (implements io.Reader + io.Seeker)
				seeker.Seek(start, io.SeekStart)
				limitedReader := io.LimitReader(rdr, end-start+1)
				io.Copy(c.Writer, limitedReader)
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
			io.Copy(c.Writer, limitedReader)
			return
		}

		// Full content
		c.Header("Content-Length", strconv.FormatInt(size, 10))
		c.Header("Content-Type", contentType)
		c.Header("Accept-Ranges", "bytes")
		c.Status(http.StatusOK)
		io.Copy(c.Writer, rdr)
	}
}

// parseRange parses the Range header and returns start and end positions
func parseRange(rangeHeader string, fileSize int64) (int64, int64, error) {
	rangeHeader = strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangeHeader, "-")

	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format")
	}

	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	var end int64
	if parts[1] == "" {
		end = fileSize - 1
	} else {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, 0, err
		}
	}

	if start >= fileSize || end >= fileSize || start > end {
		return 0, 0, fmt.Errorf("invalid range")
	}

	return start, end, nil
}

// getContentType determines the content type based on file extension
func getContentType(filePath string) string {
	kind, _ := filetype.MatchFile(filePath)
	if kind != filetype.Unknown {
		return kind.MIME.Value
	}
	return "application/octet-stream"
}

func bestMatch(file string, files []string) (string, error) {
	var bestMatch string
	bestSim := 0.0
	for _, f := range files {
		sim := utils.Similarity(f, file)
		log.Printf("considering file %q %v %s with similarity %.3f", f, f, f, sim)
		if sim > bestSim || strings.EqualFold(f, file) {
			bestSim = sim
			bestMatch = f
		}
	}
	if bestSim < 0.4 {
		return "", errors.New("file not found in archive")
	}
	log.Printf("best match: %q with similarity %.3f", bestMatch, bestSim)
	return bestMatch, nil
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
