package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNextMergeID(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".nodex", "merges"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".nodex", "merges", "merge-0001.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".nodex", "merges", "merge-0012.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := NextMergeID(root)
	if got != "merge-0013" {
		t.Fatalf("expected merge-0013, got %s", got)
	}
}
