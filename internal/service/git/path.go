package git

import (
	"fmt"

	"github.com/tedla-brandsema/mcpfs/internal/core"
)

type pathScopeKind int

const (
	pathScopeFile pathScopeKind = iota
	pathScopeFileOrDir
)

func (s *Service) resolveOptionalPath(root *core.Root, requested string, kind pathScopeKind) (string, error) {
	if requested == "" {
		return "", nil
	}

	rel, err := s.resolve(root, requested)
	if err != nil {
		return "", err
	}

	switch kind {
	case pathScopeFile:
		if !root.Matcher.AllowFile(rel) {
			return "", fmt.Errorf("file is excluded")
		}
	case pathScopeFileOrDir:
		if !root.Matcher.AllowFile(rel) && !root.Matcher.AllowDir(rel) {
			return "", fmt.Errorf("path is excluded")
		}
	default:
		return "", fmt.Errorf("unknown path scope kind")
	}

	return rel, nil
}

func appendGitPathspec(args []string, rel string) []string {
	if rel == "" {
		return args
	}
	return append(args, "--", rel)
}
