package archivequery

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

type fakeArchiveService struct {
	listFiles []string
	listErr   error
	openErr   error
	reader    io.ReadCloser
	opened    string
}

func (f *fakeArchiveService) ListFiles(context.Context, []byte) ([]string, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listFiles, nil
}

func (f *fakeArchiveService) OpenFile(_ context.Context, _ []byte, filename string) (io.ReadCloser, error) {
	f.opened = filename
	if f.openErr != nil {
		return nil, f.openErr
	}
	if f.reader != nil {
		return f.reader, nil
	}
	return io.NopCloser(strings.NewReader("ok")), nil
}

type trackingReader struct {
	r      io.Reader
	closed bool
}

func (t *trackingReader) Read(p []byte) (int, error) {
	return t.r.Read(p)
}

func (t *trackingReader) Close() error {
	t.closed = true
	return nil
}

func TestReadBestMatch(t *testing.T) {
	t.Parallel()

	t.Run("maps list error to invalid archive", func(t *testing.T) {
		t.Parallel()

		svc := &fakeArchiveService{listErr: errors.New("bad archive")}
		_, _, err := ReadBestMatch(context.Background(), svc, []byte("archive"), "report.pdf", func(a, b string) float64 {
			if a == b {
				return 1
			}
			return 0.1
		})
		if !errors.Is(err, ErrInvalidArchive) {
			t.Fatalf("ReadBestMatch() error = %v, want ErrInvalidArchive", err)
		}
	})

	t.Run("returns file-not-found for weak match", func(t *testing.T) {
		t.Parallel()

		svc := &fakeArchiveService{listFiles: []string{"doc.txt"}}
		_, _, err := ReadBestMatch(context.Background(), svc, []byte("archive"), "report.pdf", func(_, _ string) float64 { return 0.1 })
		if !errors.Is(err, ErrFileNotFound) {
			t.Fatalf("ReadBestMatch() error = %v, want ErrFileNotFound", err)
		}
	})

	t.Run("reads best match and closes reader", func(t *testing.T) {
		t.Parallel()

		tr := &trackingReader{r: strings.NewReader("payload")}
		svc := &fakeArchiveService{
			listFiles: []string{"docs/report.pdf"},
			reader:    tr,
		}

		payload, name, err := ReadBestMatch(context.Background(), svc, []byte("archive"), "report.pdf", func(a, b string) float64 {
			if a == "docs/report.pdf" && b == "report.pdf" {
				return 0.9
			}
			return 0.1
		})
		if err != nil {
			t.Fatalf("ReadBestMatch() error = %v, want nil", err)
		}
		if string(payload) != "payload" {
			t.Fatalf("ReadBestMatch() payload = %q, want %q", string(payload), "payload")
		}
		if name != "docs/report.pdf" {
			t.Fatalf("ReadBestMatch() name = %q, want %q", name, "docs/report.pdf")
		}
		if svc.opened != "docs/report.pdf" {
			t.Fatalf("OpenFile() called with %q, want %q", svc.opened, "docs/report.pdf")
		}
		if !tr.closed {
			t.Fatalf("ReadBestMatch() did not close reader")
		}
	})
}
