package filequery

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const minSimilarity = 0.4

type SimilarityFunc func(a, b string) float64

func BestMatch(file string, files []string, similarity SimilarityFunc) (string, error) {
	if similarity == nil {
		return "", errors.New("similarity function is required")
	}

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
