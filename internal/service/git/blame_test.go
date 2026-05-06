package git

import (
	"context"
	"testing"
)

func TestParseBlame(t *testing.T) {
	input := "" +
		"abcdef1234567890 1 10 1\n" +
		"author Alice\n" +
		"author-mail <alice@example.com>\n" +
		"author-time 1714564800\n" +
		"summary Add first line\n" +
		"filename file.go\n" +
		"\tpackage main\n" +
		"123456abcdef7890 2 11 1\n" +
		"author Bob\n" +
		"author-mail <bob@example.com>\n" +
		"author-time 1714568400\n" +
		"summary Add second line\n" +
		"filename file.go\n" +
		"\tfunc main() {}\n"

	lines, err := ParseBlame(input)
	if err != nil {
		t.Fatal(err)
	}

	if len(lines) != 2 {
		t.Fatalf("len(lines) = %d, want 2", len(lines))
	}

	if lines[0].Line != 10 {
		t.Fatalf("lines[0].Line = %d, want 10", lines[0].Line)
	}
	if lines[0].Commit != "abcdef1234567890" {
		t.Fatalf("lines[0].Commit = %q", lines[0].Commit)
	}
	if lines[0].ShortCommit != "abcdef123456" {
		t.Fatalf("lines[0].ShortCommit = %q", lines[0].ShortCommit)
	}
	if lines[0].Author != "Alice" {
		t.Fatalf("lines[0].Author = %q", lines[0].Author)
	}
	if lines[0].AuthorEmail != "alice@example.com" {
		t.Fatalf("lines[0].AuthorEmail = %q", lines[0].AuthorEmail)
	}
	if lines[0].AuthorTime != "2024-05-01T12:00:00Z" {
		t.Fatalf("lines[0].AuthorTime = %q", lines[0].AuthorTime)
	}
	if lines[0].Summary != "Add first line" {
		t.Fatalf("lines[0].Summary = %q", lines[0].Summary)
	}
	if lines[0].Text != "package main" {
		t.Fatalf("lines[0].Text = %q", lines[0].Text)
	}

	if lines[1].Line != 11 {
		t.Fatalf("lines[1].Line = %d, want 11", lines[1].Line)
	}
	if lines[1].Author != "Bob" {
		t.Fatalf("lines[1].Author = %q", lines[1].Author)
	}
	if lines[1].Text != "func main() {}" {
		t.Fatalf("lines[1].Text = %q", lines[1].Text)
	}
}

func TestParseBlameEmpty(t *testing.T) {
	lines, err := ParseBlame("")
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 0 {
		t.Fatalf("len(lines) = %d, want 0", len(lines))
	}
}

func TestParseBlameIgnoresIncompleteFinalRecord(t *testing.T) {
	input := "" +
		"abcdef1234567890 1 10 1\n" +
		"author Alice\n" +
		"author-mail <alice@example.com>\n" +
		"author-time 1714564800\n" +
		"summary Add first line\n"

	lines, err := ParseBlame(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 0 {
		t.Fatalf("len(lines) = %d, want 0", len(lines))
	}
}

func TestParseBlameRejectsContentWithoutRecord(t *testing.T) {
	_, err := ParseBlame("\tcontent without record\n")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFormatGitUnixTimeInvalid(t *testing.T) {
	if got := formatGitUnixTime("not-a-number"); got != "" {
		t.Fatalf("formatGitUnixTime returned %q, want empty", got)
	}
}

func TestBlameReturnsLineRange(t *testing.T) {
	dir := t.TempDir()
	initTestRepo(t, dir)

	writeFile(t, dir, "file.txt", "one\ntwo\nthree\n")
	gitCommand(t, dir, "add", "file.txt")
	gitCommand(t, dir, "commit", "-m", "add file")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Blame(context.Background(), BlameArgs{
		RootID:    "repo",
		Path:      "file.txt",
		StartLine: 2,
		EndLine:   3,
		MaxBytes:  65536,
	})
	if err != nil {
		t.Fatalf("Blame returned error: %v", err)
	}

	if result.MaxBytes != 65536 {
		t.Fatalf("MaxBytes = %d, want 65536", result.MaxBytes)
	}
	if result.Path != "file.txt" {
		t.Fatalf("Path = %q, want file.txt", result.Path)
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
	if result.Lines[0].Line != 2 || result.Lines[0].Text != "two" {
		t.Fatalf("Lines[0] = %#v", result.Lines[0])
	}
	if result.Lines[1].Line != 3 || result.Lines[1].Text != "three" {
		t.Fatalf("Lines[1] = %#v", result.Lines[1])
	}
	if result.Truncated {
		t.Fatal("Truncated = true, want false")
	}
}

func TestParseBlameHeaderAllowsThreeFields(t *testing.T) {
	line, err := parseBlameHeader("abcdef1234567890 1 10")
	if err != nil {
		t.Fatal(err)
	}

	if line.Line != 10 {
		t.Fatalf("Line = %d, want 10", line.Line)
	}
	if line.Commit != "abcdef1234567890" {
		t.Fatalf("Commit = %q", line.Commit)
	}
}

func TestParseBlameHeaderTrimsBoundaryPrefix(t *testing.T) {
	line, err := parseBlameHeader("^abcdef1234567890 1 10 1")
	if err != nil {
		t.Fatal(err)
	}

	if line.Commit != "abcdef1234567890" {
		t.Fatalf("Commit = %q", line.Commit)
	}
	if line.ShortCommit != "abcdef123456" {
		t.Fatalf("ShortCommit = %q", line.ShortCommit)
	}
}
