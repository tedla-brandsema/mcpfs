package fs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tedla-brandsema/mcpfs/internal/config"
)

func TestPatchRejectsReadRoot(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world")
	svc := newTestService(t, testRootConfig("repo", dir))

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "hello", New: "goodbye"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchReplacesExactText(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	result, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "hello", New: "goodbye"},
		},
	})
	if err != nil {
		t.Fatalf("Patch returned error: %v", err)
	}

	if !result.Changed {
		t.Fatal("Changed = false, want true")
	}
	if result.EditsApplied != 1 {
		t.Fatalf("EditsApplied = %d, want 1", result.EditsApplied)
	}
	if result.BytesBefore != len("hello world\n") {
		t.Fatalf("BytesBefore = %d, want %d", result.BytesBefore, len("hello world\n"))
	}
	if result.BytesAfter != len("goodbye world\n") {
		t.Fatalf("BytesAfter = %d, want %d", result.BytesAfter, len("goodbye world\n"))
	}
	if !strings.Contains(result.Diff, "-hello world\n") {
		t.Fatalf("Diff = %q, want removed line", result.Diff)
	}
	if !strings.Contains(result.Diff, "+goodbye world\n") {
		t.Fatalf("Diff = %q, want added line", result.Diff)
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "goodbye world\n" {
		t.Fatalf("file content = %q, want goodbye world", data)
	}
}

func TestPatchDryRunDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	result, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		DryRun: true,
		Edits: []PatchEdit{
			{Old: "hello", New: "goodbye"},
		},
	})
	if err != nil {
		t.Fatalf("Patch returned error: %v", err)
	}

	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	if !result.Changed {
		t.Fatal("Changed = false, want true")
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world\n" {
		t.Fatalf("file content = %q, want unchanged", data)
	}
}

func TestPatchWithExpectedSHA256AppliesMatchingPatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID:         "repo",
		Path:           "hello.txt",
		ExpectedSHA256: testSHA256("hello world\n"),
		Edits: []PatchEdit{
			{Old: "hello", New: "goodbye"},
		},
	})
	if err != nil {
		t.Fatalf("Patch returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "goodbye world\n" {
		t.Fatalf("file content = %q, want goodbye world", data)
	}
}

func TestPatchWithExpectedSHA256RejectsMismatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID:         "repo",
		Path:           "hello.txt",
		ExpectedSHA256: testSHA256("different"),
		Edits: []PatchEdit{
			{Old: "hello", New: "goodbye"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world\n" {
		t.Fatalf("file content = %q, want unchanged", data)
	}
}

func TestPatchDryRunWithExpectedSHA256RejectsMismatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID:         "repo",
		Path:           "hello.txt",
		DryRun:         true,
		ExpectedSHA256: testSHA256("different"),
		Edits: []PatchEdit{
			{Old: "hello", New: "goodbye"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world\n" {
		t.Fatalf("file content = %q, want unchanged", data)
	}
}

func TestPatchAppliesMultipleEdits(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "one two three\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	result, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "one", New: "ONE"},
			{Old: "three", New: "THREE"},
		},
	})
	if err != nil {
		t.Fatalf("Patch returned error: %v", err)
	}
	if result.EditsApplied != 2 {
		t.Fatalf("EditsApplied = %d, want 2", result.EditsApplied)
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "ONE two THREE\n" {
		t.Fatalf("file content = %q, want patched", data)
	}
}

func TestPatchRejectsZeroMatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "missing", New: "replacement"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchRejectsMultipleMatches(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello hello\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "hello", New: "goodbye"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchRejectsEmptyOldText(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "", New: "replacement"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchRejectsEmptyEdits(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchRejectsExcludedFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "secret.txt", "secret")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite
	cfg.Exclude = []string{"secret.txt"}

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "secret.txt",
		Edits: []PatchEdit{
			{Old: "secret", New: "public"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchRejectsContentOverMaxFileBytes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite
	cfg.MaxFileBytes = 5

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "hello", New: "hello world"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchRejectsDirectoryTarget(t *testing.T) {
	dir := t.TempDir()
	mkdir(t, dir, "target")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "target",
		Edits: []PatchEdit{
			{Old: "hello", New: "goodbye"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchMultiEditFailureIsAtomic(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "one two three\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "one", New: "ONE"},
			{Old: "missing", New: "MISSING"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "one two three\n" {
		t.Fatalf("file content = %q, want unchanged", data)
	}
}

func TestPatchDiffUsesLocalizedHunks(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", strings.Join([]string{
		"line 1",
		"line 2",
		"line 3",
		"line 4",
		"line 5",
		"line 6",
		"line 7",
		"line 8",
		"line 9",
		"line 10",
		"",
	}, "\n"))

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	result, err := svc.Patch(context.Background(), PatchArgs{
		RootID:           "repo",
		Path:             "hello.txt",
		DryRun:           true,
		DiffContextLines: 1,
		Edits: []PatchEdit{
			{Old: "line 5", New: "LINE 5"},
		},
	})
	if err != nil {
		t.Fatalf("Patch returned error: %v", err)
	}

	if result.DiffContextLines != 1 {
		t.Fatalf("DiffContextLines = %d, want 1", result.DiffContextLines)
	}
	if !strings.Contains(result.Diff, "@@ -4,3 +4,3 @@\n") {
		t.Fatalf("Diff = %q, want localized hunk header", result.Diff)
	}
	if !strings.Contains(result.Diff, " line 4\n-line 5\n+LINE 5\n line 6\n") {
		t.Fatalf("Diff = %q, want one context line around change", result.Diff)
	}
	if strings.Contains(result.Diff, "line 1") || strings.Contains(result.Diff, "line 10") {
		t.Fatalf("Diff = %q, want distant unchanged lines omitted", result.Diff)
	}
}

func TestPatchDiffDefaultsToThreeContextLines(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", strings.Join([]string{
		"line 1",
		"line 2",
		"line 3",
		"line 4",
		"line 5",
		"line 6",
		"line 7",
		"line 8",
		"line 9",
		"line 10",
		"",
	}, "\n"))

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	result, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		DryRun: true,
		Edits: []PatchEdit{
			{Old: "line 5", New: "LINE 5"},
		},
	})
	if err != nil {
		t.Fatalf("Patch returned error: %v", err)
	}

	if result.DiffContextLines != 3 {
		t.Fatalf("DiffContextLines = %d, want 3", result.DiffContextLines)
	}
	if !strings.Contains(result.Diff, "@@ -2,7 +2,7 @@\n") {
		t.Fatalf("Diff = %q, want default context hunk header", result.Diff)
	}
	if !strings.Contains(result.Diff, " line 2\n line 3\n line 4\n-line 5\n+LINE 5\n line 6\n line 7\n line 8\n") {
		t.Fatalf("Diff = %q, want three context lines around change", result.Diff)
	}
}

func TestPatchDiffMergesNearbyHunks(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", strings.Join([]string{
		"line 1",
		"line 2",
		"line 3",
		"line 4",
		"line 5",
		"line 6",
		"line 7",
		"",
	}, "\n"))

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	result, err := svc.Patch(context.Background(), PatchArgs{
		RootID:           "repo",
		Path:             "hello.txt",
		DryRun:           true,
		DiffContextLines: 1,
		Edits: []PatchEdit{
			{Old: "line 3", New: "LINE 3"},
			{Old: "line 5", New: "LINE 5"},
		},
	})
	if err != nil {
		t.Fatalf("Patch returned error: %v", err)
	}

	if got := strings.Count(result.Diff, "@@ "); got != 1 {
		t.Fatalf("hunk count = %d, want 1; diff = %q", got, result.Diff)
	}
	if !strings.Contains(result.Diff, "-line 3\n+LINE 3\n") || !strings.Contains(result.Diff, "-line 5\n+LINE 5\n") {
		t.Fatalf("Diff = %q, want both changes", result.Diff)
	}
}

func TestPatchTruncatesDiff(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	result, err := svc.Patch(context.Background(), PatchArgs{
		RootID:       "repo",
		Path:         "hello.txt",
		DryRun:       true,
		MaxDiffBytes: 10,
		Edits: []PatchEdit{
			{Old: "hello", New: "goodbye"},
		},
	})
	if err != nil {
		t.Fatalf("Patch returned error: %v", err)
	}
	if !result.DiffTruncated {
		t.Fatal("DiffTruncated = false, want true")
	}
	if len(result.Diff) > 10 {
		t.Fatalf("len(Diff) = %d, want <= 10", len(result.Diff))
	}
}
