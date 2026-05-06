package git

import "testing"

func TestAppendGitPathspecEmptyPath(t *testing.T) {
	args := []string{"log", "-n", "5"}

	got := appendGitPathspec(args, "")

	if len(got) != len(args) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(args))
	}
	for i := range args {
		if got[i] != args[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], args[i])
		}
	}
}

func TestAppendGitPathspec(t *testing.T) {
	got := appendGitPathspec([]string{"show", "HEAD"}, "internal/service/git/show.go")

	want := []string{"show", "HEAD", "--", "internal/service/git/show.go"}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d; got = %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q; got = %#v", i, got[i], want[i], got)
		}
	}
}
