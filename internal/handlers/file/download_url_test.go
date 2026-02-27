package file

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMoveFile(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	src := filepath.Join(srcDir, "source.bin")
	dst := filepath.Join(dstDir, "target.bin")

	if err := os.WriteFile(src, []byte("payload"), 0o644); err != nil {
		t.Fatalf("WriteFile(src) error = %v", err)
	}

	if err := moveFile(src, dst); err != nil {
		t.Fatalf("moveFile() error = %v", err)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("source should not exist after move, err = %v", err)
	}

	b, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile(dst) error = %v", err)
	}
	if string(b) != "payload" {
		t.Fatalf("moved bytes = %q, want %q", string(b), "payload")
	}
}
