package fs

import (
	"context"
	"testing"
)

func TestRootsReturnsRootsSortedByID(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	svc := newTestService(t,
		testRootConfig("z-root", dirB),
		testRootConfig("a-root", dirA),
	)

	result, err := svc.Roots(context.Background(), RootsArgs{})
	if err != nil {
		t.Fatalf("Roots returned error: %v", err)
	}

	if len(result.Roots) != 2 {
		t.Fatalf("len(Roots) = %d, want 2", len(result.Roots))
	}
	if result.Roots[0].ID != "a-root" {
		t.Fatalf("Roots[0].ID = %q, want %q", result.Roots[0].ID, "a-root")
	}
	if result.Roots[1].ID != "z-root" {
		t.Fatalf("Roots[1].ID = %q, want %q", result.Roots[1].ID, "z-root")
	}
}
