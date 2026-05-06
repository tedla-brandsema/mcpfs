package fs

import (
	"context"
	"fmt"
	iofs "io/fs"
)

func (s *Service) Stat(ctx context.Context, args StatArgs) (StatResult, error) {
	_ = ctx

	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("mcpfs.stat", args.RootID, args.Path, err.Error())
		return StatResult{}, err
	}

	rel, err := s.resolve(root, args.Path)
	if err != nil {
		s.logDenied("mcpfs.stat", root.ID, args.Path, err.Error())
		return StatResult{}, err
	}

	info, err := iofs.Stat(root.ReadFS, rel)
	if err != nil {
		s.logDenied("mcpfs.stat", root.ID, rel, err.Error())
		return StatResult{}, err
	}

	if info.IsDir() {
		if !root.Matcher.AllowDir(rel) {
			err := fmt.Errorf("directory is excluded")
			s.logDenied("mcpfs.stat", root.ID, rel, err.Error())
			return StatResult{}, err
		}
	} else if !root.Matcher.AllowFile(rel) {
		err := fmt.Errorf("file is excluded")
		s.logDenied("mcpfs.stat", root.ID, rel, err.Error())
		return StatResult{}, err
	}

	typ := "file"
	if info.IsDir() {
		typ = "dir"
	}

	result := StatResult{
		RootID: root.ID,
		Path:   rel,
		Type:   typ,
		Size:   info.Size(),
		MTime:  info.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
		Mode:   info.Mode().String(),
	}

	s.logAllowed("mcpfs.stat", root.ID, rel, "type", typ)
	return result, nil
}
