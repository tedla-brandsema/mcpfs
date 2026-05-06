package fs

import (
	"context"
	"fmt"
	iofs "io/fs"
	"path/filepath"

	"github.com/tedla-brandsema/mcpfs/internal/core"
	"github.com/tedla-brandsema/mcpfs/internal/limits"
)

func (s *Service) List(ctx context.Context, args ListArgs) (ListResult, error) {
	_ = ctx

	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("mcpfs.list", args.RootID, args.Path, err.Error())
		return ListResult{}, err
	}

	requested := args.Path
	if requested == "" {
		requested = "."
	}

	rel, err := s.resolve(root, requested)
	if err != nil {
		s.logDenied("mcpfs.list", root.ID, requested, err.Error())
		return ListResult{}, err
	}

	info, err := iofs.Stat(root.ReadFS, rel)
	if err != nil {
		s.logDenied("mcpfs.list", root.ID, rel, err.Error())
		return ListResult{}, err
	}
	if !info.IsDir() {
		err := fmt.Errorf("path is not a directory")
		s.logDenied("mcpfs.list", root.ID, rel, err.Error())
		return ListResult{}, err
	}
	if !root.Matcher.AllowDir(rel) {
		err := fmt.Errorf("directory is excluded")
		s.logDenied("mcpfs.list", root.ID, rel, err.Error())
		return ListResult{}, err
	}

	maxEntries := limits.ClampInt(args.MaxEntries, 200, 1000)

	result := ListResult{
		RootID:     root.ID,
		Path:       rel,
		MaxEntries: maxEntries,
		Entries:    make([]Entry, 0),
	}

	if args.Recursive {
		err = iofs.WalkDir(root.ReadFS, rel, func(pathRel string, d iofs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return nil
			}

			if pathRel == rel {
				return nil
			}

			safeRel, err := s.resolve(root, pathRel)
			if err != nil {
				s.logDenied("mcpfs.list", root.ID, pathRel, err.Error())
				if d.IsDir() {
					return iofs.SkipDir
				}
				return nil
			}

			if d.IsDir() {
				if !root.Matcher.AllowDir(safeRel) {
					return iofs.SkipDir
				}
			} else if !root.Matcher.AllowFile(safeRel) {
				return nil
			}

			if len(result.Entries) >= maxEntries {
				result.Truncated = true
				if d.IsDir() {
					return iofs.SkipDir
				}
				return nil
			}

			entry, err := makeEntry(root, safeRel)
			if err == nil {
				result.Entries = append(result.Entries, entry)
			}

			return nil
		})
	} else {
		var entries []iofs.DirEntry
		entries, err = iofs.ReadDir(root.ReadFS, rel)
		if err == nil {
			for _, d := range entries {
				if len(result.Entries) >= maxEntries {
					result.Truncated = true
					break
				}

				entryRel := joinRel(rel, d.Name())

				safeRel, err := s.resolve(root, entryRel)
				if err != nil {
					s.logDenied("mcpfs.list", root.ID, entryRel, err.Error())
					continue
				}

				if d.IsDir() {
					if !root.Matcher.AllowDir(safeRel) {
						continue
					}
				} else if !root.Matcher.AllowFile(safeRel) {
					continue
				}

				entry, err := makeEntry(root, safeRel)
				if err == nil {
					result.Entries = append(result.Entries, entry)
				}
			}
		}
	}

	if err != nil {
		s.logDenied("mcpfs.list", root.ID, rel, err.Error())
		return ListResult{}, err
	}

	s.logAllowed("mcpfs.list", root.ID, rel, "entries", len(result.Entries), "truncated", result.Truncated)
	return result, nil
}

func makeEntry(root *core.Root, rel string) (Entry, error) {
	info, err := iofs.Stat(root.ReadFS, rel)
	if err != nil {
		return Entry{}, err
	}

	typ := "file"
	if info.IsDir() {
		typ = "dir"
	}

	return Entry{
		Path:  filepath.ToSlash(rel),
		Type:  typ,
		Size:  info.Size(),
		MTime: info.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
	}, nil
}
