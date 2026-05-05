package git

type StatusArgs struct {
	RootID string `json:"root_id" jsonschema:"configured root id"`
}

type StatusResult struct {
	RootID    string      `json:"root_id"`
	Branch    string      `json:"branch"`
	Clean     bool        `json:"clean"`
	Changes   []Change    `json:"changes"`
	Truncated bool        `json:"truncated"`
}

type Change struct {
	Path     string `json:"path"`
	OldPath  string `json:"old_path,omitempty"`
	Index    string `json:"index"`
	Worktree string `json:"worktree"`
	Staged   bool   `json:"staged"`
	Unstaged bool   `json:"unstaged"`
	Status   string `json:"status"`
}