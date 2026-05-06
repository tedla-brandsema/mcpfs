package git

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tedla-brandsema/mcpfs/internal/limits"
)

const (
	defaultBlameLineWindow = 200
	maxBlameLineWindow     = 1000
)

func (s *Service) Blame(ctx context.Context, args BlameArgs) (BlameResult, error) {
	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("git.blame", args.RootID, args.Path, err.Error())
		return BlameResult{}, err
	}

	if args.Path == "" {
		err := fmt.Errorf("path is required")
		s.logDenied("git.blame", root.ID, args.Path, err.Error())
		return BlameResult{}, err
	}

	pathForResult, err := s.resolveOptionalPath(root, args.Path, pathScopeFile)
	if err != nil {
		s.logDenied("git.blame", root.ID, args.Path, err.Error())
		return BlameResult{}, err
	}

	startLine := args.StartLine
	if startLine <= 0 {
		startLine = 1
	}

	endLine := args.EndLine
	if endLine <= 0 {
		endLine = startLine + defaultBlameLineWindow - 1
	}
	if endLine < startLine {
		err := fmt.Errorf("end_line must be >= start_line")
		s.logDenied("git.blame", root.ID, pathForResult, err.Error())
		return BlameResult{}, err
	}

	window := endLine - startLine + 1
	window = limits.ClampInt(window, defaultBlameLineWindow, maxBlameLineWindow)
	endLine = startLine + window - 1

	maxBytes := limits.ClampInt(args.MaxBytes, defaultGitOutputLimit, defaultGitOutputLimit)

	stdout, stderr, truncated, err := runGit(ctx, root.RealPath, maxBytes,
		"blame",
		"--line-porcelain",
		"-L",
		fmt.Sprintf("%d,%d", startLine, endLine),
		"--",
		pathForResult,
	)
	if err != nil {
		err := fmt.Errorf("git blame: %w: %s", err, stderr)
		s.logDenied("git.blame", root.ID, pathForResult, err.Error())
		return BlameResult{}, err
	}

	output, outputTruncated := limits.CapStringBytes(stdout, maxBytes)
	truncated = truncated || outputTruncated

	lines, err := ParseBlame(output)
	if err != nil {
		s.logDenied("git.blame", root.ID, pathForResult, err.Error())
		return BlameResult{}, err
	}

	result := BlameResult{
		RootID:    root.ID,
		Path:      pathForResult,
		StartLine: startLine,
		EndLine:   endLine,
		Bytes:     len(output),
		MaxBytes:  maxBytes,
		Lines:     lines,
		Truncated: truncated,
	}

	s.logAllowed(
		"git.blame",
		root.ID,
		pathForResult,
		"start_line", result.StartLine,
		"end_line", result.EndLine,
		"bytes", result.Bytes,
		"lines", len(result.Lines),
		"truncated", result.Truncated,
	)

	return result, nil
}

func ParseBlame(output string) ([]BlameLine, error) {
	if strings.TrimSpace(output) == "" {
		return []BlameLine{}, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	lines := make([]BlameLine, 0)
	var current *BlameLine
	inRecord := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "\t") {
			if current == nil {
				return nil, fmt.Errorf("invalid blame output: content line without record")
			}

			current.Text = strings.TrimPrefix(line, "\t")
			lines = append(lines, *current)
			current = nil
			inRecord = false
			continue
		}

		if maybeBlameHeader(line) {
			blameLine, err := parseBlameHeader(line)
			if err != nil {
				return nil, err
			}

			current = &blameLine
			inRecord = true
			continue
		}

		if !inRecord || current == nil {
			continue
		}

		parseBlameMetadata(current, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// If output was truncated mid-record, ignore the incomplete final record.
	return lines, nil
}

func maybeBlameHeader(line string) bool {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return false
	}

	if len(parts[0]) < 7 {
		return false
	}

	if _, err := strconv.Atoi(parts[1]); err != nil {
		return false
	}
	if _, err := strconv.Atoi(parts[2]); err != nil {
		return false
	}
	if len(parts) >= 4 {
		if _, err := strconv.Atoi(parts[3]); err != nil {
			return false
		}
	}

	return true
}

func parseBlameHeader(line string) (BlameLine, error) {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return BlameLine{}, fmt.Errorf("invalid blame header: %q", line)
	}

	lineNo, err := strconv.Atoi(parts[2])
	if err != nil {
		return BlameLine{}, fmt.Errorf("invalid blame line number: %w", err)
	}

	commit := strings.TrimPrefix(parts[0], "^")
	shortCommit := commit
	if len(shortCommit) > 12 {
		shortCommit = shortCommit[:12]
	}

	return BlameLine{
		Line:        lineNo,
		Commit:      commit,
		ShortCommit: shortCommit,
	}, nil
}

func parseBlameMetadata(line *BlameLine, metadata string) {
	key, value, ok := strings.Cut(metadata, " ")
	if !ok {
		return
	}

	switch key {
	case "author":
		line.Author = value
	case "author-mail":
		line.AuthorEmail = strings.Trim(value, "<>")
	case "author-time":
		line.AuthorTime = formatGitUnixTime(value)
	case "summary":
		line.Summary = value
	}
}

func formatGitUnixTime(value string) string {
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return ""
	}

	return time.Unix(seconds, 0).UTC().Format(time.RFC3339)
}
