package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	projectservice "github.com/tedla-brandsema/mcpfs/internal/service/project"
)

const (
	localProjectConfigDir  = ".mcpfs"
	localProjectConfigFile = "project.cfg.json"
)

type initOptions struct {
	pathOpt string
}

func runInit(args []string, logger *slog.Logger) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var opts initOptions
	fs.StringVar(&opts.pathOpt, "path", "", "project directory; defaults to the current directory")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	if fs.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %v\n", fs.Args())
		return 2
	}

	projectPath, err := resolveProjectPath(opts.pathOpt)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	configPath, created, err := writeLocalProjectConfig(projectPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if created {
		logger.Info("created project config", "path", configPath)
		fmt.Fprintf(os.Stdout, "created %s\n", configPath)
	} else {
		logger.Info("project config already exists", "path", configPath)
		fmt.Fprintf(os.Stdout, "exists %s\n", configPath)
	}

	return 0
}

func writeLocalProjectConfig(projectDir string) (string, bool, error) {
	configDir := filepath.Join(projectDir, localProjectConfigDir)
	configPath := filepath.Join(configDir, localProjectConfigFile)

	if _, err := os.Stat(configPath); err == nil {
		return configPath, false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", false, fmt.Errorf("stat project config: %w", err)
	}

	if err := projectservice.WriteDefaultRegistryConfig(configPath); err != nil {
		return "", false, err
	}

	return configPath, true, nil
}
