package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrCreateRegistryWritesSeedConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "mcpfs", projectConfigFileName)

	registry, err := LoadOrCreateRegistry(configPath)
	if err != nil {
		t.Fatalf("LoadOrCreateRegistry returned error: %v", err)
	}

	if !registry.IsImportantFile("README.md") {
		t.Fatal("README.md is not important")
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("project config was not written: %v", err)
	}
}

func TestLoadOrCreateRegistryLoadsExistingConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "mcpfs", projectConfigFileName)

	data := []byte(`{
		"project": {
			"important_files": ["local.project"],
			"source_extensions": [".localsrc"],
			"test_patterns": ["*.localtest"],
			"documentation_extensions": [],
			"documentation_files": [],
			"configuration_extensions": [],
			"configuration_files": ["local.project"]
		}
	}`)

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	registry, err := LoadOrCreateRegistry(configPath)
	if err != nil {
		t.Fatalf("LoadOrCreateRegistry returned error: %v", err)
	}

	if !registry.IsImportantFile("local.project") {
		t.Fatal("local.project is not important")
	}
	if registry.IsImportantFile("README.md") {
		t.Fatal("README.md should not be important for custom registry")
	}
	if !registry.IsSourceFile("main.localsrc") {
		t.Fatal("main.localsrc is not a source file")
	}
	if !registry.IsTestFile("main.localtest") {
		t.Fatal("main.localtest is not a test file")
	}
	if !registry.IsConfigurationFile("local.project") {
		t.Fatal("local.project is not a configuration file")
	}
}

func TestLoadOrCreateRegistryRejectsInvalidJSON(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "mcpfs", projectConfigFileName)

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte(`{invalid`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadOrCreateRegistry(configPath)
	if err == nil {
		t.Fatal("LoadOrCreateRegistry returned nil error")
	}
}

func TestRegistryMergeOverlaysNonEmptyProjectRules(t *testing.T) {
	base := Registry{
		Project: ProjectRules{
			ImportantFiles:          []string{"README.md"},
			SourceExtensions:        []string{".go"},
			DocumentationExtensions: []string{".md"},
		},
	}
	local := Registry{
		Project: ProjectRules{
			ImportantFiles:   []string{"local.project"},
			SourceExtensions: []string{".localsrc"},
		},
	}

	merged := base.Merge(local)

	if !merged.IsImportantFile("local.project") || merged.IsImportantFile("README.md") {
		t.Fatalf("merged important file rules did not overlay as expected")
	}
	if !merged.IsSourceFile("main.localsrc") || merged.IsSourceFile("main.go") {
		t.Fatalf("merged source extension rules did not overlay as expected")
	}
	if !merged.IsDocumentationFile("README.md") {
		t.Fatalf("empty local documentation rules should inherit base documentation rules")
	}
}
