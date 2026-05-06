package fs

import (
	"context"
	"testing"
)

func TestListHidesExcludedFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "visible.txt", "visible")
	writeFile(t, dir, "secret.txt", "secret")

	cfg := testRootConfig("repo", dir)
	cfg.Exclude = []string{"secret.txt"}

	svc := newTestService(t, cfg)

	result, err := svc.List(context.Background(), ListArgs{
		RootID: "repo",
		Path:   ".",
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	assertEntryPaths(t, result.Entries, []string{"visible.txt"})
}

func TestListRecursiveHonorsMaxEntries(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "a")
	writeFile(t, dir, "b.txt", "b")
	writeFile(t, dir, "nested/c.txt", "c")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.List(context.Background(), ListArgs{
		RootID:     "repo",
		Path:       ".",
		Recursive:  true,
		MaxEntries: 2,
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if len(result.Entries) != 2 {
		t.Fatalf("len(Entries) = %d, want 2", len(result.Entries))
	}
	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
}
