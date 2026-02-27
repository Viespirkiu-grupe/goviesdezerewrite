package fileid

import "testing"

func TestIsNumericOrMD5(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		id    string
		valid bool
	}{
		{name: "numeric", id: "123456", valid: true},
		{name: "md5 lowercase", id: "d41d8cd98f00b204e9800998ecf8427e", valid: true},
		{name: "md5 uppercase", id: "D41D8CD98F00B204E9800998ECF8427E", valid: true},
		{name: "short hex", id: "3bb17dea17c7665749b8584ed311", valid: false},
		{name: "alpha", id: "file-name", valid: false},
		{name: "empty", id: "", valid: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := IsNumericOrMD5(tc.id); got != tc.valid {
				t.Fatalf("IsNumericOrMD5(%q) = %v, want %v", tc.id, got, tc.valid)
			}
		})
	}
}
