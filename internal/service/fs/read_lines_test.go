package fs

import (
	"context"
	"testing"
)

func TestReadLinesReturnsLineRange(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "data.txt", "one\ntwo\nthree\nfour\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.ReadLines(context.Background(), ReadLinesArgs{
		RootID:    "repo",
		Path:      "data.txt",
		StartLine: 2,
		EndLine:   3,
	})
	if err != nil {
		t.Fatalf("ReadLines returned error: %v", err)
	}

	if result.RootID != "repo" {
		t.Fatalf("RootID = %q, want %q", result.RootID, "repo")
	}
	if result.Path != "data.txt" {
		t.Fatalf("Path = %q, want %q", result.Path, "data.txt")
	}
	if result.StartLine != 2 {
		t.Fatalf("StartLine = %d, want 2", result.StartLine)
	}
	if result.EndLine != 3 {
		t.Fatalf("EndLine = %d, want 3", result.EndLine)
	}
	if len(result.Lines) != 2 {
		t.Fatalf("len(Lines) = %d, want 2", len(result.Lines))
	}
	if result.Lines[0].Number != 2 || result.Lines[0].Text != "two" {
		t.Fatalf("Lines[0] = %#v, want line 2 two", result.Lines[0])
	}
	if result.Lines[1].Number != 3 || result.Lines[1].Text != "three" {
		t.Fatalf("Lines[1] = %#v, want line 3 three", result.Lines[1])
	}
	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
}

func TestReadLinesDefaultsStartLine(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "data.txt", "one\ntwo\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.ReadLines(context.Background(), ReadLinesArgs{
		RootID:  "repo",
		Path:    "data.txt",
		EndLine: 1,
	})
	if err != nil {
		t.Fatalf("ReadLines returned error: %v", err)
	}

	if len(result.Lines) != 1 {
		t.Fatalf("len(Lines) = %d, want 1", len(result.Lines))
	}
	if result.Lines[0].Number != 1 || result.Lines[0].Text != "one" {
		t.Fatalf("Lines[0] = %#v, want line 1 one", result.Lines[0])
	}
}

func TestReadLinesRejectsInvalidRange(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "data.txt", "one\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	_, err := svc.ReadLines(context.Background(), ReadLinesArgs{
		RootID:    "repo",
		Path:      "data.txt",
		StartLine: 3,
		EndLine:   2,
	})
	if err == nil {
		t.Fatal("ReadLines returned nil error")
	}
}

func TestReadLinesRejectsExcludedFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "secret.txt", "secret\n")

	cfg := testRootConfig("repo", dir)
	cfg.Exclude = []string{"secret.txt"}

	svc := newTestService(t, cfg)

	_, err := svc.ReadLines(context.Background(), ReadLinesArgs{
		RootID: "repo",
		Path:   "secret.txt",
	})
	if err == nil {
		t.Fatal("ReadLines returned nil error")
	}
}

func TestReadLinesRejectsFileOverMaxFileBytes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "large.txt", "abcdef\n")

	cfg := testRootConfig("repo", dir)
	cfg.MaxFileBytes = 3

	svc := newTestService(t, cfg)

	_, err := svc.ReadLines(context.Background(), ReadLinesArgs{
		RootID: "repo",
		Path:   "large.txt",
	})
	if err == nil {
		t.Fatal("ReadLines returned nil error")
	}
}

func TestReadLinesReturnsEmptyPastEOF(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "data.txt", "one\ntwo\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.ReadLines(context.Background(), ReadLinesArgs{
		RootID:    "repo",
		Path:      "data.txt",
		StartLine: 10,
		EndLine:   20,
	})
	if err != nil {
		t.Fatalf("ReadLines returned error: %v", err)
	}

	if len(result.Lines) != 0 {
		t.Fatalf("len(Lines) = %d, want 0", len(result.Lines))
	}
	if result.Truncated {
		t.Fatal("Truncated = true, want false")
	}
}
