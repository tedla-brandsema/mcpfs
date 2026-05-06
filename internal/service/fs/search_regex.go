package fs

import (
	"bufio"
	"context"
	"fmt"
	iofs "io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/tedla-brandsema/mcpfs/internal/core"
	"github.com/tedla-brandsema/mcpfs/internal/limits"
)

func (s *Service) SearchRegex(ctx context.Context, args SearchRegexArgs) (SearchRegexResult, error) {
	_ = ctx

	if args.Query == "" {
		return SearchRegexResult{}, fmt.Errorf("query is required")
	}

	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("mcpfs.search_regex", args.RootID, "", err.Error())
		return SearchRegexResult{}, err
	}

	caseSensitive := true
	if args.CaseSensitive != nil {
		caseSensitive = *args.CaseSensitive
	}

	pattern := args.Query
	if !caseSensitive {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		s.logDenied("mcpfs.search_regex", root.ID, args.Query, err.Error())
		return SearchRegexResult{}, fmt.Errorf("invalid regex: %w", err)
	}

	maxResults := limits.ClampInt(args.MaxResults, 50, 500)
	glob := filepath.ToSlash(strings.TrimSpace(args.Glob))

	result := SearchRegexResult{
		RootID:        root.ID,
		Query:         args.Query,
		Glob:          glob,
		CaseSensitive: caseSensitive,
		Matches:       make([]SearchMatch, 0),
	}

	err = iofs.WalkDir(root.ReadFS, ".", func(pathRel string, d iofs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}

		if pathRel == "." {
			return nil
		}

		safeRel, err := s.resolve(root, pathRel)
		if err != nil {
			s.logDenied("mcpfs.search_regex", root.ID, pathRel, err.Error())
			if d.IsDir() {
				return iofs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if !root.Matcher.AllowDir(safeRel) {
				return iofs.SkipDir
			}
			return nil
		}

		if !root.Matcher.AllowFile(safeRel) {
			return nil
		}

		if glob != "" {
			ok, _ := doublestar.PathMatch(glob, safeRel)
			if !ok {
				return nil
			}
		}

		info, err := iofs.Stat(root.ReadFS, safeRel)
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if info.Size() > root.MaxFileBytes {
			return nil
		}

		matches, err := searchRegexFile(root, safeRel, re, maxResults-len(result.Matches))
		if err != nil {
			return nil
		}

		result.Matches = append(result.Matches, matches...)

		if len(result.Matches) >= maxResults {
			result.Truncated = true
			return iofs.SkipAll
		}

		return nil
	})
	if err != nil {
		s.logDenied("mcpfs.search_regex", root.ID, args.Query, err.Error())
		return SearchRegexResult{}, err
	}

	s.logAllowed(
		"mcpfs.search_regex",
		root.ID,
		args.Query,
		"matches", len(result.Matches),
		"truncated", result.Truncated,
	)

	return result, nil
}

func searchRegexFile(root *core.Root, rel string, re *regexp.Regexp, remaining int) ([]SearchMatch, error) {
	if remaining <= 0 {
		return nil, nil
	}

	f, err := root.ReadFS.Open(rel)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	var matches []SearchMatch
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()

		if re.MatchString(line) {
			matches = append(matches, SearchMatch{
				Path:    rel,
				Line:    lineNo,
				Preview: strings.TrimSpace(line),
			})

			if len(matches) >= remaining {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return matches, err
	}

	return matches, nil
}
