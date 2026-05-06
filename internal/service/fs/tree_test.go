package fs

import (
	"context"
	"testing"
)

func TestTreeReturnsStructuredTreeAndText(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "a")
	writeFile(t, dir, "nested/b.txt", "b")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Tree(context.Background(), TreeArgs{
		RootID:   "repo",
		Path:     ".",
		MaxDepth: 3,
	})
	if err != nil {
		t.Fatalf("Tree returned error: %v", err)
	}

	if result.RootID != "repo" {
		t.Fatalf("RootID = %q, want %q", result.RootID, "repo")
	}
	if result.Path != "." {
		t.Fatalf("Path = %q, want %q", result.Path, ".")
	}
	if result.Root.Path != "." {
		t.Fatalf("Root.Path = %q, want %q", result.Root.Path, ".")
	}
	if result.Root.Type != "dir" {
		t.Fatalf("Root.Type = %q, want %q", result.Root.Type, "dir")
	}

	assertTreeEntryPaths(t, result.Entries, []string{"a.txt", "nested", "nested/b.txt"})

	if result.Entries[0].Depth != 1 {
		t.Fatalf("Entries[0].Depth = %d, want 1", result.Entries[0].Depth)
	}
	if result.Entries[2].Depth != 2 {
		t.Fatalf("Entries[2].Depth = %d, want 2", result.Entries[2].Depth)
	}
	if result.Entries[2].ParentPath != "nested" {
		t.Fatalf("Entries[2].ParentPath = %q, want %q", result.Entries[2].ParentPath, "nested")
	}

	wantText := ".\n├── a.txt\n└── nested\n    └── b.txt"
	if result.Text != wantText {
		t.Fatalf("Text = %q, want %q", result.Text, wantText)
	}
	if result.Truncated {
		t.Fatal("Truncated = true, want false")
	}
}

func TestTreeHonorsMaxDepth(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "nested/deeper/file.txt", "x")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Tree(context.Background(), TreeArgs{
		RootID:   "repo",
		Path:     ".",
		MaxDepth: 1,
	})
	if err != nil {
		t.Fatalf("Tree returned error: %v", err)
	}

	assertTreeEntryPaths(t, result.Entries, []string{"nested"})
}

func TestTreeHonorsMaxEntries(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "a")
	writeFile(t, dir, "b.txt", "b")
	writeFile(t, dir, "c.txt", "c")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Tree(context.Background(), TreeArgs{
		RootID:     "repo",
		Path:       ".",
		MaxEntries: 2,
	})
	if err != nil {
		t.Fatalf("Tree returned error: %v", err)
	}

	assertTreeEntryPaths(t, result.Entries, []string{"a.txt", "b.txt"})
	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
}

func TestTreeCanExcludeFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "a")
	writeFile(t, dir, "nested/b.txt", "b")

	includeFiles := false
	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Tree(context.Background(), TreeArgs{
		RootID:       "repo",
		Path:         ".",
		IncludeFiles: &includeFiles,
	})
	if err != nil {
		t.Fatalf("Tree returned error: %v", err)
	}

	assertTreeEntryPaths(t, result.Entries, []string{"nested"})
}

func TestTreeHidesExcludedFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "visible.txt", "visible")
	writeFile(t, dir, "secret.txt", "secret")

	cfg := testRootConfig("repo", dir)
	cfg.Exclude = []string{"secret.txt"}

	svc := newTestService(t, cfg)

	result, err := svc.Tree(context.Background(), TreeArgs{
		RootID: "repo",
		Path:   ".",
	})
	if err != nil {
		t.Fatalf("Tree returned error: %v", err)
	}

	assertTreeEntryPaths(t, result.Entries, []string{"visible.txt"})
}
