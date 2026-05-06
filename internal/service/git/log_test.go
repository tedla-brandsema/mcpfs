package git

import "testing"

func TestParseLogEmpty(t *testing.T) {
	commits, err := ParseLog("")
	if err != nil {
		t.Fatal(err)
	}

	if len(commits) != 0 {
		t.Fatalf("commits len = %d, want 0", len(commits))
	}
}

func TestParseLogSingleCommit(t *testing.T) {
	input := "abc123def\x00abc123d\x00Tedla\x00tedla@example.com\x002026-05-05T08:00:00+02:00\x00Add git log\x00Body text\x00\x1e"

	commits, err := ParseLog(input)
	if err != nil {
		t.Fatal(err)
	}

	if len(commits) != 1 {
		t.Fatalf("commits len = %d, want 1", len(commits))
	}

	commit := commits[0]
	if commit.Hash != "abc123def" {
		t.Fatalf("hash = %q", commit.Hash)
	}
	if commit.ShortHash != "abc123d" {
		t.Fatalf("short_hash = %q", commit.ShortHash)
	}
	if commit.AuthorName != "Tedla" {
		t.Fatalf("author_name = %q", commit.AuthorName)
	}
	if commit.AuthorEmail != "tedla@example.com" {
		t.Fatalf("author_email = %q", commit.AuthorEmail)
	}
	if commit.AuthorDate != "2026-05-05T08:00:00+02:00" {
		t.Fatalf("author_date = %q", commit.AuthorDate)
	}
	if commit.Subject != "Add git log" {
		t.Fatalf("subject = %q", commit.Subject)
	}
	if commit.Body != "Body text" {
		t.Fatalf("body = %q", commit.Body)
	}
}

func TestParseLogMultipleCommits(t *testing.T) {
	input := "" +
		"hash1\x00h1\x00Alice\x00alice@example.com\x002026-05-05T08:00:00+02:00\x00First\x00\x00\x1e" +
		"hash2\x00h2\x00Bob\x00bob@example.com\x002026-05-04T08:00:00+02:00\x00Second\x00Body\x00\x1e"

	commits, err := ParseLog(input)
	if err != nil {
		t.Fatal(err)
	}

	if len(commits) != 2 {
		t.Fatalf("commits len = %d, want 2", len(commits))
	}

	if commits[0].Subject != "First" {
		t.Fatalf("first subject = %q", commits[0].Subject)
	}
	if commits[1].Subject != "Second" {
		t.Fatalf("second subject = %q", commits[1].Subject)
	}
	if commits[1].Body != "Body" {
		t.Fatalf("second body = %q", commits[1].Body)
	}
}

func TestParseLogInvalidRecord(t *testing.T) {
	_, err := ParseLog("too few fields\x00only two")
	if err == nil {
		t.Fatal("expected error")
	}
}
