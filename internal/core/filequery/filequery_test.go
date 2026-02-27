package filequery

import (
	"math"
	"testing"
)

func TestParseRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rangeValue string
		size       int64
		wantStart  int64
		wantEnd    int64
		wantErr    bool
	}{
		{name: "full range", rangeValue: "bytes=0-9", size: 100, wantStart: 0, wantEnd: 9},
		{name: "tail from start", rangeValue: "bytes=10-", size: 100, wantStart: 10, wantEnd: 99},
		{name: "invalid format", rangeValue: "items=0-9", size: 100, wantErr: true},
		{name: "start greater than end", rangeValue: "bytes=10-1", size: 100, wantErr: true},
		{name: "start outside file", rangeValue: "bytes=101-200", size: 100, wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotStart, gotEnd, err := ParseRange(tc.rangeValue, tc.size)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseRange() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseRange() error = %v, want nil", err)
			}

			if gotStart != tc.wantStart || gotEnd != tc.wantEnd {
				t.Fatalf("ParseRange() = (%d,%d), want (%d,%d)", gotStart, gotEnd, tc.wantStart, tc.wantEnd)
			}
		})
	}
}

func TestBestMatch(t *testing.T) {
	t.Parallel()

	similarity := func(a, b string) float64 {
		if a == b {
			return 1
		}
		if a == "docs/report.pdf" && b == "report.pdf" {
			return 0.8
		}
		if a == "docs/r.pdf" && b == "report.pdf" {
			return 0.2
		}
		return math.SmallestNonzeroFloat64
	}

	t.Run("chooses best similarity", func(t *testing.T) {
		t.Parallel()

		got, err := BestMatch("report.pdf", []string{"docs/r.pdf", "docs/report.pdf"}, similarity)
		if err != nil {
			t.Fatalf("BestMatch() error = %v, want nil", err)
		}
		if got != "docs/report.pdf" {
			t.Fatalf("BestMatch() = %q, want %q", got, "docs/report.pdf")
		}
	})

	t.Run("returns exact match even if similarity lower", func(t *testing.T) {
		t.Parallel()

		got, err := BestMatch("docs/report.pdf", []string{"docs/r.pdf", "docs/report.pdf"}, func(a, b string) float64 {
			if a == b {
				return 0.1
			}
			return 0.2
		})
		if err != nil {
			t.Fatalf("BestMatch() error = %v, want nil", err)
		}
		if got != "docs/report.pdf" {
			t.Fatalf("BestMatch() = %q, want %q", got, "docs/report.pdf")
		}
	})

	t.Run("returns error when no viable candidate", func(t *testing.T) {
		t.Parallel()

		_, err := BestMatch("report.pdf", []string{"docs/r.pdf"}, func(_, _ string) float64 { return 0.1 })
		if err == nil {
			t.Fatalf("BestMatch() error = nil, want error")
		}
	})
}
