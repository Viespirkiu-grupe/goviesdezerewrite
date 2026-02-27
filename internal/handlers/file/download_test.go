package file

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
