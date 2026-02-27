package file

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"goviesdeze/internal/config"
	"goviesdeze/internal/utils"

	"github.com/gin-gonic/gin"
)

type trackingReadCloser struct {
	reader io.Reader
	closed bool
}

func (t *trackingReadCloser) Read(p []byte) (int, error) {
	return t.reader.Read(p)
}

func (t *trackingReadCloser) Close() error {
	t.closed = true
	return nil
}

func TestWriteResponseClosesReader(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/file/report.pdf", nil)
	rdr := &trackingReadCloser{reader: strings.NewReader("content")}

	upstream := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
	}

	hasErr := writeResponse(rr, req, rdr, upstream, "report.pdf")
	if hasErr {
		t.Fatalf("writeResponse() returned error flag, want false")
	}

	if !rdr.closed {
		t.Fatalf("writeResponse() did not close reader")
	}
}

func TestGetFileSupportsSuffixAndClampedRanges(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storage := t.TempDir()
	filename := "abcdef"
	filePath := utils.ShardPath(filename, storage)

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := os.WriteFile(filePath, []byte("0123456789"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg := &config.Config{StoragePath: storage}
	router := gin.New()
	router.GET("/file/:filename", GetFile(cfg))

	t.Run("suffix range bytes=-4", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/file/%s", filename), nil)
		req.Header.Set("Range", "bytes=-4")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusPartialContent {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusPartialContent)
		}
		if rr.Body.String() != "6789" {
			t.Fatalf("body = %q, want %q", rr.Body.String(), "6789")
		}
		if got := rr.Header().Get("Content-Range"); got != "bytes 6-9/10" {
			t.Fatalf("Content-Range = %q, want %q", got, "bytes 6-9/10")
		}
	})

	t.Run("clamped end bytes=8-1000", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/file/%s", filename), nil)
		req.Header.Set("Range", "bytes=8-1000")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusPartialContent {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusPartialContent)
		}
		if rr.Body.String() != "89" {
			t.Fatalf("body = %q, want %q", rr.Body.String(), "89")
		}
		if got := rr.Header().Get("Content-Range"); got != "bytes 8-9/10" {
			t.Fatalf("Content-Range = %q, want %q", got, "bytes 8-9/10")
		}
	})
}
