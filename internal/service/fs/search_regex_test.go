package fs

import (
	"context"
	"testing"
)

func TestSearchRegexReturnsMatches(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "auth.go", "package auth\n\nfunc Authenticate() {}\n")
	writeFile(t, dir, "other.go", "package other\n\nfunc SomethingElse() {}\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.SearchRegex(context.Background(), SearchRegexArgs{
		RootID:     "repo",
		Query:      `func .*Authenticate`,
		Glob:       "**/*.go",
		MaxResults: 10,
	})
	if err != nil {
		t.Fatalf("SearchRegex returned error: %v", err)
	}

	if result.RootID != "repo" {
		t.Fatalf("RootID = %q, want repo", result.RootID)
	}
	if result.Query != `func .*Authenticate` {
		t.Fatalf("Query = %q", result.Query)
	}
	if result.Glob != "**/*.go" {
		t.Fatalf("Glob = %q", result.Glob)
	}
	if !result.CaseSensitive {
		t.Fatal("CaseSensitive = false, want true")
	}
	if len(result.Matches) != 1 {
		t.Fatalf("len(Matches) = %d, want 1; matches = %#v", len(result.Matches), result.Matches)
	}
	if result.Matches[0].Path != "auth.go" {
		t.Fatalf("Path = %q, want auth.go", result.Matches[0].Path)
	}
	if result.Matches[0].Line != 3 {
		t.Fatalf("Line = %d, want 3", result.Matches[0].Line)
	}
	if result.Matches[0].Preview != "func Authenticate() {}" {
		t.Fatalf("Preview = %q", result.Matches[0].Preview)
	}
	if result.Truncated {
		t.Fatal("Truncated = true, want false")
	}
}

func TestSearchRegexCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "auth.go", "func Authenticate() {}\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	caseSensitive := false
	result, err := svc.SearchRegex(context.Background(), SearchRegexArgs{
		RootID:        "repo",
		Query:         `func authenticate`,
		CaseSensitive: &caseSensitive,
	})
	if err != nil {
		t.Fatalf("SearchRegex returned error: %v", err)
	}

	if result.CaseSensitive {
		t.Fatal("CaseSensitive = true, want false")
	}
	if len(result.Matches) != 1 {
		t.Fatalf("len(Matches) = %d, want 1", len(result.Matches))
	}
}

func TestSearchRegexRejectsInvalidRegex(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "auth.go", "func Authenticate() {}\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	_, err := svc.SearchRegex(context.Background(), SearchRegexArgs{
		RootID: "repo",
		Query:  `[`,
	})
	if err == nil {
		t.Fatal("SearchRegex returned nil error")
	}
}

func TestSearchRegexHonorsMaxResults(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "match one\nmatch two\n")
	writeFile(t, dir, "b.txt", "match three\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.SearchRegex(context.Background(), SearchRegexArgs{
		RootID:     "repo",
		Query:      `match`,
		MaxResults: 2,
	})
	if err != nil {
		t.Fatalf("SearchRegex returned error: %v", err)
	}

	if result.MaxResults != 2 {
		t.Fatalf("MaxResults = %d, want 2", result.MaxResults)
	}
	if len(result.Matches) != 2 {
		t.Fatalf("len(Matches) = %d, want 2", len(result.Matches))
	}
	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
}

func TestSearchRegexHonorsExclude(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "visible.txt", "secret\n")
	writeFile(t, dir, "secret.txt", "secret\n")

	cfg := testRootConfig("repo", dir)
	cfg.Exclude = []string{"secret.txt"}

	svc := newTestService(t, cfg)

	result, err := svc.SearchRegex(context.Background(), SearchRegexArgs{
		RootID: "repo",
		Query:  `secret`,
	})
	if err != nil {
		t.Fatalf("SearchRegex returned error: %v", err)
	}

	if len(result.Matches) != 1 {
		t.Fatalf("len(Matches) = %d, want 1", len(result.Matches))
	}
	if result.Matches[0].Path != "visible.txt" {
		t.Fatalf("Path = %q, want visible.txt", result.Matches[0].Path)
	}
}
