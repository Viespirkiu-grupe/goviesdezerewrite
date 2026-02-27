package ziparchive

import (
	"context"
	"io"

	"goviesdeze/internal/ziputil"
)

type Service struct{}

func (Service) ListFiles(ctx context.Context, archiveBytes []byte) ([]string, error) {
	return ziputil.IdentityFilesV2(ctx, archiveBytes)
}

func (Service) OpenFile(ctx context.Context, archiveBytes []byte, filename string) (io.ReadCloser, error) {
	return ziputil.GetFileFromArchiveV2(ctx, archiveBytes, filename)
}
