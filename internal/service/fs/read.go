package fs

import (
	"context"
	"fmt"
	"io"
	iofs "io/fs"

	"github.com/tedla-brandsema/mcpfs/internal/limits"
)

func (s *Service) Read(ctx context.Context, args ReadArgs) (ReadResult, error) {
	_ = ctx

	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("mcpfs.read", args.RootID, args.Path, err.Error())
		return ReadResult{}, err
	}

	rel, err := s.resolve(root, args.Path)
	if err != nil {
		s.logDenied("mcpfs.read", root.ID, args.Path, err.Error())
		return ReadResult{}, err
	}

	if !root.Matcher.AllowFile(rel) {
		err := fmt.Errorf("file is excluded")
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}

	info, err := iofs.Stat(root.ReadFS, rel)
	if err != nil {
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}
	if info.IsDir() {
		err := fmt.Errorf("path is a directory")
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}
	if info.Size() > root.MaxFileBytes {
		err := fmt.Errorf("file exceeds max_file_bytes: size=%d max=%d", info.Size(), root.MaxFileBytes)
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}
	if args.Offset < 0 {
		err := fmt.Errorf("offset must be >= 0")
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}

	limit := limits.ClampInt64(args.Limit, root.MaxFileBytes, root.MaxFileBytes)

	f, err := root.ReadFS.Open(rel)
	if err != nil {
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}
	defer f.Close()

	if args.Offset > 0 {
		n, err := io.CopyN(io.Discard, f, args.Offset)
		if err != nil && err != io.EOF {
			s.logDenied("mcpfs.read", root.ID, rel, err.Error())
			return ReadResult{}, err
		}
		if n < args.Offset {
			result := ReadResult{
				RootID:    root.ID,
				Path:      rel,
				Bytes:     0,
				Size:      info.Size(),
				Offset:    args.Offset,
				Limit:     limit,
				Truncated: false,
				Content:   "",
			}
			s.logAllowed("mcpfs.read", root.ID, rel, "bytes", result.Bytes, "truncated", result.Truncated)
			return result, nil
		}
	}

	data, err := io.ReadAll(io.LimitReader(f, limit))
	if err != nil {
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}

	result := ReadResult{
		RootID:    root.ID,
		Path:      rel,
		Bytes:     len(data),
		Size:      info.Size(),
		Offset:    args.Offset,
		Limit:     limit,
		Truncated: args.Offset+int64(len(data)) < info.Size(),
		Content:   string(data),
	}

	s.logAllowed("mcpfs.read", root.ID, rel, "bytes", result.Bytes, "truncated", result.Truncated)
	return result, nil
}
