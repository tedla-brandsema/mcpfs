package fs

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/tedla-brandsema/mcpfs/internal/config"
	"github.com/tedla-brandsema/mcpfs/internal/core"
)

func newTestService(t *testing.T, configs ...config.RootConfig) *Service {
	t.Helper()

	roots := make([]*core.Root, 0, len(configs))
	for _, cfg := range configs {
		root, err := core.NewRoot(cfg, discardLogger())
		if err != nil {
			t.Fatalf("NewRoot(%q) returned error: %v", cfg.ID, err)
		}
		roots = append(roots, root)
	}

	return New(roots, discardLogger())
}

func testRootConfig(id string, dir string) config.RootConfig {
	return config.RootConfig{
		ID:           id,
		Path:         dir,
		Mode:         config.ModeRead,
		Include:      []string{"**/*"},
		Exclude:      nil,
		UseGitignore: false,
		MaxFileBytes: 262144,
	}
}

func writeFile(t *testing.T, root string, rel string, content string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) returned error: %v", path, err)
	}
}

func mkdir(t *testing.T, root string, rel string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", path, err)
	}
}

func assertTreeEntryPaths(t *testing.T, entries []TreeEntry, want []string) {
	t.Helper()

	if len(entries) != len(want) {
		t.Fatalf("len(entries) = %d, want %d; entries = %#v", len(entries), len(want), entries)
	}

	for i := range want {
		if entries[i].Path != want[i] {
			t.Fatalf("entries[%d].Path = %q, want %q; entries = %#v", i, entries[i].Path, want[i], entries)
		}
	}
}

func assertEntryPaths(t *testing.T, entries []Entry, want []string) {
	t.Helper()

	if len(entries) != len(want) {
		t.Fatalf("len(entries) = %d, want %d; entries = %#v", len(entries), len(want), entries)
	}

	for i := range want {
		if entries[i].Path != want[i] {
			t.Fatalf("entries[%d].Path = %q, want %q; entries = %#v", i, entries[i].Path, want[i], entries)
		}
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
