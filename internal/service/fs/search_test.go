package fs

import (
	"context"
	"testing"
)

func TestSearchFindsMatchingLines(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "alpha\nneedle here\nomega\n")
	writeFile(t, dir, "b.txt", "nothing\nneedle again\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Search(context.Background(), SearchArgs{
		RootID: "repo",
		Query:  "needle",
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if len(result.Matches) != 2 {
		t.Fatalf("len(Matches) = %d, want 2", len(result.Matches))
	}
	if result.Matches[0].Line != 2 {
		t.Fatalf("Matches[0].Line = %d, want 2", result.Matches[0].Line)
	}
}

func TestSearchHonorsGlob(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "needle\n")
	writeFile(t, dir, "b.md", "needle\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Search(context.Background(), SearchArgs{
		RootID: "repo",
		Query:  "needle",
		Glob:   "**/*.md",
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if len(result.Matches) != 1 {
		t.Fatalf("len(Matches) = %d, want 1", len(result.Matches))
	}
	if result.Matches[0].Path != "b.md" {
		t.Fatalf("Matches[0].Path = %q, want %q", result.Matches[0].Path, "b.md")
	}
}

func TestSearchRequiresQuery(t *testing.T) {
	dir := t.TempDir()

	svc := newTestService(t, testRootConfig("repo", dir))

	_, err := svc.Search(context.Background(), SearchArgs{
		RootID: "repo",
		Query:  "",
	})
	if err == nil {
		t.Fatal("Search returned nil error")
	}
}

func TestSearchTruncatesAtMaxResults(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "needle 1\nneedle 2\nneedle 3\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Search(context.Background(), SearchArgs{
		RootID:     "repo",
		Query:      "needle",
		MaxResults: 2,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if len(result.Matches) != 2 {
		t.Fatalf("len(Matches) = %d, want 2", len(result.Matches))
	}
	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
}
