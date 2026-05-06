package fs

import (
	"context"
)

func (s *Service) Roots(ctx context.Context, args RootsArgs) (RootsResult, error) {
	_ = ctx
	_ = args

	out := RootsResult{
		Roots: make([]RootInfo, 0, len(s.order)),
	}

	for _, id := range s.order {
		root := s.roots[id]
		out.Roots = append(out.Roots, RootInfo{
			ID:           root.ID,
			Mode:         string(root.Mode),
			MaxFileBytes: root.MaxFileBytes,
		})
	}

	s.logger.Info("mcpfs allowed", "service", s.Name(), "event", "mcpfs.roots", "roots", len(out.Roots))
	return out, nil
}
