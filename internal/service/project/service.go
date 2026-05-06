package project

import (
	"fmt"
	"log/slog"

	"github.com/tedla-brandsema/mcpfs/internal/core"
	fsservice "github.com/tedla-brandsema/mcpfs/internal/service/fs"
	gitservice "github.com/tedla-brandsema/mcpfs/internal/service/git"
)

type Service struct {
	fs                 *fsservice.Service
	git                *gitservice.Service
	registry           Registry
	registryPath       string
	rootRegistries     map[string]Registry
	localRegistryPaths map[string]string
	logger             *slog.Logger
}

func New(fsSvc *fsservice.Service, gitSvc *gitservice.Service, logger *slog.Logger) (*Service, error) {
	return NewWithRoots(fsSvc, gitSvc, nil, logger)
}

func NewWithRoots(fsSvc *fsservice.Service, gitSvc *gitservice.Service, roots []*core.Root, logger *slog.Logger) (*Service, error) {
	registry, registryPath, err := LoadOrCreateDefaultRegistry()
	if err != nil {
		return nil, fmt.Errorf("load project registry: %w", err)
	}

	return NewWithRegistryAndRoots(fsSvc, gitSvc, registry, registryPath, roots, logger)
}

func NewWithRegistry(fsSvc *fsservice.Service, gitSvc *gitservice.Service, registry Registry, registryPath string, logger *slog.Logger) *Service {
	svc, err := NewWithRegistryAndRoots(fsSvc, gitSvc, registry, registryPath, nil, logger)
	if err != nil {
		panic(err)
	}
	return svc
}

func NewWithRegistryAndRoots(
	fsSvc *fsservice.Service,
	gitSvc *gitservice.Service,
	registry Registry,
	registryPath string,
	roots []*core.Root,
	logger *slog.Logger,
) (*Service, error) {
	if logger == nil {
		logger = slog.Default()
	}

	rootRegistries := make(map[string]Registry)
	localRegistryPaths := make(map[string]string)

	for _, root := range roots {
		rootRegistry, configPath, loaded, err := LoadRootRegistry(root, registry)
		if err != nil {
			return nil, fmt.Errorf("load project-local config for root %q: %w", root.ID, err)
		}

		if loaded {
			rootRegistries[root.ID] = rootRegistry
			localRegistryPaths[root.ID] = configPath
			logger.Info("loaded project-local config", "root_id", root.ID, "path", configPath)
		}
	}

	return &Service{
		fs:                 fsSvc,
		git:                gitSvc,
		registry:           registry,
		registryPath:       registryPath,
		rootRegistries:     rootRegistries,
		localRegistryPaths: localRegistryPaths,
		logger:             logger,
	}, nil
}

func (s *Service) Name() string {
	return "project"
}

func (s *Service) logAllowed(event string, rootID string, path string, attrs ...any) {
	args := []any{
		"service", s.Name(),
		"event", event,
		"root_id", rootID,
		"path", path,
	}
	args = append(args, attrs...)

	s.logger.Info("mcpfs allowed", args...)
}

func (s *Service) logDenied(event string, rootID string, path string, reason string) {
	s.logger.Warn(
		"mcpfs denied",
		slog.String("service", s.Name()),
		slog.String("event", event),
		slog.String("root_id", rootID),
		slog.String("path", path),
		slog.String("reason", reason),
	)
}

func (s *Service) registryFor(rootID string) Registry {
	if registry, ok := s.rootRegistries[rootID]; ok {
		return registry
	}
	return s.registry
}
