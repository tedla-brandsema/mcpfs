package fs

import (
	"bufio"
	"context"
	"fmt"
	iofs "io/fs"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/tedla-brandsema/mcpfs/internal/core"
	"github.com/tedla-brandsema/mcpfs/internal/limits"
)

func (s *Service) Search(ctx context.Context, args SearchArgs) (SearchResult, error) {
	_ = ctx

	if args.Query == "" {
		return SearchResult{}, fmt.Errorf("query is required")
	}

	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("mcpfs.search", args.RootID, "", err.Error())
		return SearchResult{}, err
	}

	maxResults := limits.ClampInt(args.MaxResults, 50, 500)

	glob := filepath.ToSlash(strings.TrimSpace(args.Glob))

	result := SearchResult{
		RootID:     root.ID,
		Query:      args.Query,
		MaxResults: maxResults,
		Matches:    make([]SearchMatch, 0),
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
			s.logDenied("mcpfs.search", root.ID, pathRel, err.Error())
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

		matches, err := searchFile(root, safeRel, args.Query, maxResults-len(result.Matches))
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
		s.logDenied("mcpfs.search", root.ID, args.Query, err.Error())
		return SearchResult{}, err
	}

	s.logAllowed("mcpfs.search", root.ID, args.Query, "matches", len(result.Matches), "truncated", result.Truncated)
	return result, nil
}

func searchFile(root *core.Root, rel string, query string, remaining int) ([]SearchMatch, error) {
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

		if strings.Contains(line, query) {
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
