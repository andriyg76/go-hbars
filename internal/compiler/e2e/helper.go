package e2e

import (
	"path/filepath"
	"testing"
)

// repoRoot returns the repository root (go-hbars). Tests run with cwd = internal/compiler/e2e.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	return dir
}
