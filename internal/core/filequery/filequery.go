package filequery

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

const minSimilarity = 0.4

type SimilarityFunc func(a, b string) float64

func BestMatch(file string, files []string, similarity SimilarityFunc) (string, error) {
	if similarity == nil {
		return "", errors.New("similarity function is required")
	}

	//TODO: review this
	file = filepath.Base(file)

	var best string
	bestScore := 0.0

	for _, candidate := range files {
		score := similarity(candidate, file)
		if score > bestScore || strings.EqualFold(candidate, file) {
			bestScore = score
			best = candidate
		}
	}

	if best == "" || bestScore < minSimilarity {
		if !strings.EqualFold(best, file) {
			return "", errors.New("file not found in archive")
		}
	}

	return best, nil
}

func ParseRange(rangeHeader string, fileSize int64) (int64, int64, error) {
	if fileSize <= 0 {
		return 0, 0, fmt.Errorf("invalid file size")
	}

	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return 0, 0, fmt.Errorf("invalid range unit")
	}

	rangeHeader = strings.TrimPrefix(rangeHeader, "bytes=")
	if strings.Contains(rangeHeader, ",") {
		return 0, 0, fmt.Errorf("multiple ranges are not supported")
	}

	parts := strings.SplitN(rangeHeader, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format")
	}

	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	if startStr == "" && endStr == "" {
		return 0, 0, fmt.Errorf("invalid range format")
	}

	if startStr == "" {
		suffixLen, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		if suffixLen <= 0 {
			return 0, 0, fmt.Errorf("invalid suffix length")
		}

		if suffixLen >= fileSize {
			return 0, fileSize - 1, nil
		}

		start := fileSize - suffixLen
		return start, fileSize - 1, nil
	}

	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	if start < 0 {
		return 0, 0, fmt.Errorf("invalid range start")
	}

	var end int64
	if endStr == "" {
		end = fileSize - 1
	} else {
		end, err = strconv.ParseInt(endStr, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		if end < 0 {
			return 0, 0, fmt.Errorf("invalid range end")
		}
		if end >= fileSize {
			end = fileSize - 1
		}
	}

	if start >= fileSize || start > end {
		return 0, 0, fmt.Errorf("invalid range")
	}

	return start, end, nil
}
