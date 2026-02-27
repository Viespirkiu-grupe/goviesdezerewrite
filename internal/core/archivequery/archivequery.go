package archivequery

import (
	"context"
	"errors"
	"fmt"
	"io"

	"goviesdeze/internal/core/filequery"
)

var (
	ErrInvalidArchive = errors.New("invalid archive")
	ErrFileNotFound   = errors.New("file not found in archive")
	ErrExtractFailed  = errors.New("extract file from archive")
)

type Service interface {
	ListFiles(ctx context.Context, archiveBytes []byte) ([]string, error)
	OpenFile(ctx context.Context, archiveBytes []byte, filename string) (io.ReadCloser, error)
}

type SimilarityFunc = filequery.SimilarityFunc

func ReadBestMatch(ctx context.Context, svc Service, archiveBytes []byte, target string, similarity SimilarityFunc) ([]byte, string, error) {
	if svc == nil {
		return nil, "", errors.New("archive service is required")
	}

	files, err := svc.ListFiles(ctx, archiveBytes)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrInvalidArchive, err)
	}

	best, err := filequery.BestMatch(target, files, similarity)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrFileNotFound, err)
	}

	rdr, err := svc.OpenFile(ctx, archiveBytes, best)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrExtractFailed, err)
	}
	defer rdr.Close()

	body, err := io.ReadAll(rdr)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrExtractFailed, err)
	}

	return body, best, nil
}
