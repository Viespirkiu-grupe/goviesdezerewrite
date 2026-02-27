package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Levenshtein returns the Levenshtein distance between two strings
func Levenshtein(a, b string) int {
	la := len(a)
	lb := len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	cur := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		cur[0] = i
		for j := 1; j <= lb; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			ins := cur[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			// min
			min := ins
			if del < min {
				min = del
			}
			if sub < min {
				min = sub
			}
			cur[j] = min
		}
		copy(prev, cur)
	}
	return cur[lb]
}

func Similarity(a, b string) float64 {
	na := Normalize(a)
	nb := Normalize(b)
	if len(na) == 0 && len(nb) == 0 {
		return 1.0
	}
	dist := Levenshtein(na, nb)
	maxLen := len(na)
	if len(nb) > maxLen {
		maxLen = len(nb)
	}
	if maxLen == 0 {
		return 0.0
	}
	return 1.0 - float64(dist)/float64(maxLen)
}

var nonAlnumRe = regexp.MustCompile(`[^a-z0-9\.\-_]+`)

func Normalize(s string) string {
	s = strings.ToLower(s)
	t := norm.NFKD.String(s)
	b := make([]rune, 0, len(t))
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b = append(b, r)
	}
	out := string(b)
	out = nonAlnumRe.ReplaceAllString(out, "")
	return out
}
