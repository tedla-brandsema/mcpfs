package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

const defaultGitOutputLimit = 1024 * 1024

func (s *Service) Status(ctx context.Context, args StatusArgs) (StatusResult, error) {
	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("git.status", args.RootID, "", err.Error())
		return StatusResult{}, err
	}

	stdout, stderr, truncated, err := runGit(ctx, root.RealPath, defaultGitOutputLimit,
		"status",
		"--porcelain=v1",
		"-b",
		"--untracked-files=all",
	)
	if err != nil {
		err := fmt.Errorf("git status: %w: %s", err, stderr)
		s.logDenied("git.status", root.ID, "", err.Error())
		return StatusResult{}, err
	}

	branch, changes, err := ParseStatus(stdout)
	if err != nil {
		s.logDenied("git.status", root.ID, "", err.Error())
		return StatusResult{}, err
	}

	result := StatusResult{
		RootID:    root.ID,
		Branch:    branch,
		Clean:     len(changes) == 0,
		Changes:   changes,
		Truncated: truncated,
	}

	s.logAllowed("git.status", root.ID, "", "branch", result.Branch, "changes", len(result.Changes), "truncated", result.Truncated)
	return result, nil
}

func runGit(ctx context.Context, dir string, outputLimit int, args ...string) (string, string, bool, error) {
	if outputLimit <= 0 {
		outputLimit = defaultGitOutputLimit
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	var stdout cappedBuffer
	var stderr cappedBuffer

	stdout.limit = outputLimit
	stderr.limit = outputLimit

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	truncated := stdout.truncated || stderr.truncated
	return stdout.String(), stderr.String(), truncated, err
}

type cappedBuffer struct {
	bytes.Buffer
	limit     int
	truncated bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	if b.limit <= 0 {
		b.truncated = true
		return len(p), nil
	}

	remaining := b.limit - b.Buffer.Len()
	if remaining <= 0 {
		b.truncated = true
		return len(p), nil
	}

	if len(p) > remaining {
		_, _ = b.Buffer.Write(p[:remaining])
		b.truncated = true
		return len(p), nil
	}

	_, _ = b.Buffer.Write(p)
	return len(p), nil
}