package fs

import (
	"context"
	"testing"
)

func TestStatReturnsFileMetadata(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Stat(context.Background(), StatArgs{
		RootID: "repo",
		Path:   "hello.txt",
	})
	if err != nil {
		t.Fatalf("Stat returned error: %v", err)
	}

	if result.RootID != "repo" {
		t.Fatalf("RootID = %q, want %q", result.RootID, "repo")
	}
	if result.Path != "hello.txt" {
		t.Fatalf("Path = %q, want %q", result.Path, "hello.txt")
	}
	if result.Type != "file" {
		t.Fatalf("Type = %q, want %q", result.Type, "file")
	}
	if result.Size != 5 {
		t.Fatalf("Size = %d, want 5", result.Size)
	}
	if result.MTime == "" {
		t.Fatal("MTime is empty")
	}
	if result.Mode == "" {
		t.Fatal("Mode is empty")
	}
}

func TestStatReturnsDirectoryMetadata(t *testing.T) {
	dir := t.TempDir()
	mkdir(t, dir, "nested")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Stat(context.Background(), StatArgs{
		RootID: "repo",
		Path:   "nested",
	})
	if err != nil {
		t.Fatalf("Stat returned error: %v", err)
	}

	if result.Type != "dir" {
		t.Fatalf("Type = %q, want %q", result.Type, "dir")
	}
}
