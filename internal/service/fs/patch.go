package fs

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/tedla-brandsema/mcpfs/internal/config"
	"github.com/tedla-brandsema/mcpfs/internal/core"
	"github.com/tedla-brandsema/mcpfs/internal/limits"
)

const (
	defaultPatchDiffBytes        = 65536
	defaultPatchDiffContextLines = 3
)

func (s *Service) Patch(ctx context.Context, args PatchArgs) (PatchResult, error) {
	_ = ctx

	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("mcpfs.patch", args.RootID, args.Path, err.Error())
		return PatchResult{}, err
	}

	if root.Mode != config.ModeReadWrite {
		err := fmt.Errorf("root %q is not writable", root.ID)
		s.logDenied("mcpfs.patch", root.ID, args.Path, err.Error())
		return PatchResult{}, err
	}

	if len(args.Edits) == 0 {
		err := fmt.Errorf("edits are required")
		s.logDenied("mcpfs.patch", root.ID, args.Path, err.Error())
		return PatchResult{}, err
	}

	abs, err := core.ResolveWritableInsideRoot(root.RealPath, args.Path)
	if err != nil {
		s.logDenied("mcpfs.patch", root.ID, args.Path, err.Error())
		return PatchResult{}, err
	}

	rel, err := root.Rel(abs)
	if err != nil {
		s.logDenied("mcpfs.patch", root.ID, args.Path, err.Error())
		return PatchResult{}, err
	}
	rel = cleanFSRel(rel)

	if !root.Matcher.AllowFile(rel) {
		err := fmt.Errorf("file is excluded")
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}

	info, err := os.Stat(abs)
	if err != nil {
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}
	if info.IsDir() {
		err := fmt.Errorf("path is a directory")
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}
	if info.Size() > root.MaxFileBytes {
		err := fmt.Errorf("file exceeds max_file_bytes: size=%d max=%d", info.Size(), root.MaxFileBytes)
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}

	if _, err := s.verifyExpectedSHA256(root, rel, args.ExpectedSHA256); err != nil {
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}

	before := string(data)
	after, editsApplied, err := applyPatchEdits(before, args.Edits)
	if err != nil {
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}

	if int64(len(after)) > root.MaxFileBytes {
		err := fmt.Errorf("patched content exceeds max_file_bytes: size=%d max=%d", len(after), root.MaxFileBytes)
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}

	maxDiffBytes := args.MaxDiffBytes
	if maxDiffBytes <= 0 {
		maxDiffBytes = defaultPatchDiffBytes
	}

	diffContextLines := args.DiffContextLines
	if diffContextLines <= 0 {
		diffContextLines = defaultPatchDiffContextLines
	}

	diff := buildUnifiedDiff(rel, before, after, diffContextLines)
	diffText, diffTruncated := limits.CapStringBytes(diff, maxDiffBytes)

	changed := before != after
	if changed && !args.DryRun {
		if err := os.WriteFile(abs, []byte(after), fs.FileMode(info.Mode().Perm())); err != nil {
			s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
			return PatchResult{}, err
		}
	}

	result := PatchResult{
		RootID:           root.ID,
		Path:             rel,
		Mode:             string(root.Mode),
		DryRun:           args.DryRun,
		Changed:          changed,
		EditsApplied:     editsApplied,
		BytesBefore:      len(before),
		BytesAfter:       len(after),
		MaxDiffBytes:     maxDiffBytes,
		DiffContextLines: diffContextLines,
		Diff:             diffText,
		DiffTruncated:    diffTruncated,
	}

	s.logAllowed("mcpfs.patch", root.ID, rel, "edits_applied", result.EditsApplied, "changed", result.Changed, "dry_run", result.DryRun)
	return result, nil
}

func applyPatchEdits(content string, edits []PatchEdit) (string, int, error) {
	patched := content

	for i, edit := range edits {
		if edit.Old == "" {
			return "", 0, fmt.Errorf("edits[%d].old must not be empty", i)
		}

		count := strings.Count(patched, edit.Old)
		if count == 0 {
			return "", 0, fmt.Errorf("edits[%d].old matched 0 times", i)
		}
		if count > 1 {
			return "", 0, fmt.Errorf("edits[%d].old matched %d times", i, count)
		}

		patched = strings.Replace(patched, edit.Old, edit.New, 1)
	}

	return patched, len(edits), nil
}

type diffOpKind int

const (
	diffOpEqual diffOpKind = iota
	diffOpDelete
	diffOpInsert
)

type diffOp struct {
	kind    diffOpKind
	text    string
	oldLine int
	newLine int
}

func buildUnifiedDiff(path string, before string, after string, contextLines int) string {
	if before == after {
		return ""
	}
	if contextLines < 0 {
		contextLines = 0
	}

	beforeLines := splitPatchLines(before)
	afterLines := splitPatchLines(after)
	ops := lineDiffOps(beforeLines, afterLines)
	hunks := diffHunks(ops, contextLines)
	if len(hunks) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("--- a/")
	b.WriteString(path)
	b.WriteString("\n")
	b.WriteString("+++ b/")
	b.WriteString(path)
	b.WriteString("\n")

	for _, hunk := range hunks {
		oldStart, oldCount, newStart, newCount := hunkHeader(ops[hunk.start:hunk.end])
		b.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", oldStart, oldCount, newStart, newCount))

		for _, op := range ops[hunk.start:hunk.end] {
			switch op.kind {
			case diffOpEqual:
				b.WriteString(" ")
			case diffOpDelete:
				b.WriteString("-")
			case diffOpInsert:
				b.WriteString("+")
			}
			b.WriteString(op.text)
		}
	}

	return b.String()
}

func lineDiffOps(before []string, after []string) []diffOp {
	n := len(before)
	m := len(after)

	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}

	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if before[i] == after[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	ops := make([]diffOp, 0, n+m)
	oldLine := 1
	newLine := 1
	for i, j := 0, 0; i < n || j < m; {
		switch {
		case i < n && j < m && before[i] == after[j]:
			ops = append(ops, diffOp{kind: diffOpEqual, text: before[i], oldLine: oldLine, newLine: newLine})
			i++
			j++
			oldLine++
			newLine++
		case j < m && (i == n || dp[i][j+1] > dp[i+1][j]):
			ops = append(ops, diffOp{kind: diffOpInsert, text: after[j], newLine: newLine})
			j++
			newLine++
		case i < n:
			ops = append(ops, diffOp{kind: diffOpDelete, text: before[i], oldLine: oldLine})
			i++
			oldLine++
		}
	}

	return ops
}

type diffHunk struct {
	start int
	end   int
}

func diffHunks(ops []diffOp, contextLines int) []diffHunk {
	var hunks []diffHunk
	for i := 0; i < len(ops); i++ {
		if ops[i].kind == diffOpEqual {
			continue
		}

		changeStart := i
		for i+1 < len(ops) && ops[i+1].kind != diffOpEqual {
			i++
		}
		changeEnd := i + 1

		start := maxInt(0, changeStart-contextLines)
		end := minInt(len(ops), changeEnd+contextLines)

		if len(hunks) > 0 && start <= hunks[len(hunks)-1].end {
			if end > hunks[len(hunks)-1].end {
				hunks[len(hunks)-1].end = end
			}
			continue
		}

		hunks = append(hunks, diffHunk{start: start, end: end})
	}
	return hunks
}

func hunkHeader(ops []diffOp) (oldStart int, oldCount int, newStart int, newCount int) {
	oldStart = 1
	newStart = 1
	for _, op := range ops {
		if op.oldLine > 0 {
			oldStart = op.oldLine
			break
		}
	}
	for _, op := range ops {
		if op.newLine > 0 {
			newStart = op.newLine
			break
		}
	}

	for _, op := range ops {
		if op.kind == diffOpEqual || op.kind == diffOpDelete {
			oldCount++
		}
		if op.kind == diffOpEqual || op.kind == diffOpInsert {
			newCount++
		}
	}

	if oldCount == 0 {
		oldStart = maxInt(0, oldStart-1)
	}
	if newCount == 0 {
		newStart = maxInt(0, newStart-1)
	}

	return oldStart, oldCount, newStart, newCount
}

func splitPatchLines(content string) []string {
	if content == "" {
		return nil
	}

	lines := strings.SplitAfter(content, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
