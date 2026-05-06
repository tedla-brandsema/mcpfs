package fs

import (
	"context"
	"testing"
)

func TestReadReturnsContentAndMetadata(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Read(context.Background(), ReadArgs{
		RootID: "repo",
		Path:   "hello.txt",
	})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	if result.RootID != "repo" {
		t.Fatalf("RootID = %q, want %q", result.RootID, "repo")
	}
	if result.Path != "hello.txt" {
		t.Fatalf("Path = %q, want %q", result.Path, "hello.txt")
	}
	if result.Content != "hello world" {
		t.Fatalf("Content = %q, want %q", result.Content, "hello world")
	}
	if result.Bytes != len("hello world") {
		t.Fatalf("Bytes = %d, want %d", result.Bytes, len("hello world"))
	}
	if result.Size != int64(len("hello world")) {
		t.Fatalf("Size = %d, want %d", result.Size, len("hello world"))
	}
	if result.Truncated {
		t.Fatal("Truncated = true, want false")
	}
}

func TestReadHonorsOffsetAndLimit(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "data.txt", "abcdef")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Read(context.Background(), ReadArgs{
		RootID: "repo",
		Path:   "data.txt",
		Offset: 2,
		Limit:  3,
	})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	if result.Limit != 3 {
		t.Fatalf("Limit = %d, want 3", result.Limit)
	}
	if result.Content != "cde" {
		t.Fatalf("Content = %q, want %q", result.Content, "cde")
	}
	if result.Offset != 2 {
		t.Fatalf("Offset = %d, want 2", result.Offset)
	}
	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
}

func TestReadOffsetPastEOFReturnsEmptyContent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "data.txt", "abc")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Read(context.Background(), ReadArgs{
		RootID: "repo",
		Path:   "data.txt",
		Offset: 99,
	})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	if result.Content != "" {
		t.Fatalf("Content = %q, want empty", result.Content)
	}
	if result.Bytes != 0 {
		t.Fatalf("Bytes = %d, want 0", result.Bytes)
	}
	if result.Truncated {
		t.Fatal("Truncated = true, want false")
	}
}

func TestReadRejectsExcludedFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "secret.txt", "secret")

	cfg := testRootConfig("repo", dir)
	cfg.Exclude = []string{"secret.txt"}

	svc := newTestService(t, cfg)

	_, err := svc.Read(context.Background(), ReadArgs{
		RootID: "repo",
		Path:   "secret.txt",
	})
	if err == nil {
		t.Fatal("Read returned nil error")
	}
}

func TestReadRejectsFileOverMaxFileBytes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "large.txt", "abcdef")

	cfg := testRootConfig("repo", dir)
	cfg.MaxFileBytes = 3

	svc := newTestService(t, cfg)

	_, err := svc.Read(context.Background(), ReadArgs{
		RootID: "repo",
		Path:   "large.txt",
	})
	if err == nil {
		t.Fatal("Read returned nil error")
	}
}

func TestReadRejectsNegativeOffset(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "data.txt", "abc")

	svc := newTestService(t, testRootConfig("repo", dir))

	_, err := svc.Read(context.Background(), ReadArgs{
		RootID: "repo",
		Path:   "data.txt",
		Offset: -1,
	})
	if err == nil {
		t.Fatal("Read returned nil error")
	}
}
