package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/tedla-brandsema/mcpfs/internal/config"
)

type projectCommandOptions struct {
	idOpt   string
	cfgOpt  string
	pathOpt string
}

type projectCommandContext struct {
	id   string
	cfg  string
	path string
}

func runProject(args []string, logger *slog.Logger) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: mcpfs project <add|rm|ls> [flags]")
		return 2
	}

	switch args[0] {
	case "add":
		return runProjectAdd(args[1:], logger)
	case "rm":
		return runProjectRemove(args[1:], logger)
	case "ls":
		return runProjectList(args[1:], logger)
	default:
		fmt.Fprintf(os.Stderr, "unknown project command %q\n", args[0])
		fmt.Fprintln(os.Stderr, "usage: mcpfs project <add|rm|ls> [flags]")
		return 2
	}
}

func runProjectAdd(args []string, logger *slog.Logger) int {
	ctx, err := parseProjectContext("project add", args, true, true)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	if err := addRootToConfig(ctx.cfg, ctx.id, ctx.path); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	logger.Info("added project root", "config", ctx.cfg, "id", ctx.id, "path", ctx.path)
	fmt.Fprintf(os.Stdout, "added %s\t%s\n", ctx.id, ctx.path)

	return 0
}

func runProjectRemove(args []string, logger *slog.Logger) int {
	ctx, err := parseProjectContext("project rm", args, true, true)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	removed, err := removeRootFromConfig(ctx.cfg, ctx.id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if !removed {
		fmt.Fprintf(os.Stderr, "config %q does not contain root id %q\n", ctx.cfg, ctx.id)
		return 1
	}

	logger.Info("removed project root", "config", ctx.cfg, "id", ctx.id)
	fmt.Fprintf(os.Stdout, "removed %s\n", ctx.id)

	return 0
}

func runProjectList(args []string, logger *slog.Logger) int {
	ctx, err := parseProjectContext("project ls", args, false, false)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	cfg, err := config.LoadOrCreateGlobal(ctx.cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config %q: %v\n", ctx.cfg, err)
		return 1
	}

	if len(cfg.Roots) == 0 {
		fmt.Fprintln(os.Stdout, "no projects configured")
		return 0
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tMODE\tPATH")
	for _, root := range cfg.Roots {
		fmt.Fprintf(w, "%s\t%s\t%s\n", root.ID, root.Mode, root.Path)
	}
	if err := w.Flush(); err != nil {
		logger.Error("flush project list", "error", err)
		return 1
	}

	return 0
}

func parseProjectContext(command string, args []string, includeID bool, includePath bool) (projectCommandContext, error) {
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var opts projectCommandOptions

	if includeID {
		fs.StringVar(&opts.idOpt, "id", "", "project root id; defaults to the project directory name")
	}
	if includePath {
		fs.StringVar(&opts.pathOpt, "path", "", "project directory; defaults to the current directory")
	}
	fs.StringVar(&opts.cfgOpt, "cfg", "", "MCPFS config path; defaults to the global user config")

	if err := fs.Parse(args); err != nil {
		return projectCommandContext{}, err
	}
	if fs.NArg() != 0 {
		return projectCommandContext{}, fmt.Errorf("unexpected arguments: %v", fs.Args())
	}

	cfgPath, err := resolveMCPFSConfigPath(opts.cfgOpt)
	if err != nil {
		return projectCommandContext{}, err
	}

	ctx := projectCommandContext{
		cfg: cfgPath,
	}

	if includePath {
		projectPath, err := resolveProjectPath(opts.pathOpt)
		if err != nil {
			return projectCommandContext{}, err
		}

		ctx.path = projectPath
	}

	if includeID {
		ctx.id = opts.idOpt
		if ctx.id == "" {
			ctx.id = defaultProjectID(ctx.path)
		}
	}

	return ctx, nil
}

func resolveProjectPath(value string) (string, error) {
	if value == "" {
		value = "."
	}

	abs, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("resolve project path %q: %w", value, err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("stat project path %q: %w", abs, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("project path %q is not a directory", abs)
	}

	return abs, nil
}

func resolveMCPFSConfigPath(value string) (string, error) {
	if value == "" {
		return config.DefaultGlobalPath()
	}

	abs, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("resolve config path %q: %w", value, err)
	}

	return abs, nil
}

func defaultProjectID(projectPath string) string {
	base := filepath.Base(filepath.Clean(projectPath))
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "project"
	}

	return base
}

func addRootToConfig(configPath string, rootID string, rootPath string) error {
	cfg, err := config.LoadOrCreateGlobal(configPath)
	if err != nil {
		return fmt.Errorf("load config %q: %w", configPath, err)
	}

	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return fmt.Errorf("resolve root path: %w", err)
	}

	for _, root := range cfg.Roots {
		if root.ID == rootID {
			return fmt.Errorf("config %q already contains root id %q", configPath, rootID)
		}

		if samePath(root.Path, absRootPath) {
			return fmt.Errorf("config %q already contains root path %q as root id %q", configPath, absRootPath, root.ID)
		}
	}

	cfg.Roots = append(cfg.Roots, config.RootConfig{
		ID:           rootID,
		Path:         absRootPath,
		Mode:         config.ModeRead,
		Include:      []string{"**/*"},
		Exclude:      defaultRootExcludes(),
		UseGitignore: true,
		MaxFileBytes: 262144,
	})

	return writeMCPFSConfig(configPath, cfg)
}

func removeRootFromConfig(configPath string, rootID string) (bool, error) {
	cfg, err := config.LoadOrCreateGlobal(configPath)
	if err != nil {
		return false, fmt.Errorf("load config %q: %w", configPath, err)
	}

	next := cfg.Roots[:0]
	removed := false

	for _, root := range cfg.Roots {
		if root.ID == rootID {
			removed = true
			continue
		}
		next = append(next, root)
	}

	if !removed {
		return false, nil
	}

	cfg.Roots = next
	if err := writeMCPFSConfig(configPath, cfg); err != nil {
		return false, err
	}

	return true, nil
}

func writeMCPFSConfig(configPath string, cfg config.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	data = append(data, '\n')

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("write config %q: %w", configPath, err)
	}

	return nil
}

func defaultRootExcludes() []string {
	return []string{
		"**/.git/**",
		"**/.env",
		"**/.env.*",
		"**/*secret*",
		"**/*credential*",
		"**/*.pem",
		"**/*.key",
	}
}

func samePath(a string, b string) bool {
	absA, err := filepath.Abs(a)
	if err != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}

	absB, err := filepath.Abs(b)
	if err != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}

	return filepath.Clean(absA) == filepath.Clean(absB)
}
