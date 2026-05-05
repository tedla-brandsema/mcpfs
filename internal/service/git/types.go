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

type DiffArgs struct {
	RootID   string `json:"root_id" jsonschema:"configured root id"`
	Path     string `json:"path,omitempty" jsonschema:"optional relative path inside the root"`
	Staged   bool   `json:"staged,omitempty" jsonschema:"show staged diff instead of unstaged diff"`
	MaxBytes int    `json:"max_bytes,omitempty" jsonschema:"maximum diff bytes to return"`
}

type DiffResult struct {
	RootID    string `json:"root_id"`
	Path      string `json:"path,omitempty"`
	Staged    bool   `json:"staged"`
	Bytes     int    `json:"bytes"`
	Truncated bool   `json:"truncated"`
	Diff      string `json:"diff"`
}

type LogArgs struct {
	RootID   string `json:"root_id" jsonschema:"configured root id"`
	Limit    int    `json:"limit,omitempty" jsonschema:"maximum number of commits to return"`
	Path     string `json:"path,omitempty" jsonschema:"optional relative path inside the root"`
	MaxBytes int    `json:"max_bytes,omitempty" jsonschema:"maximum git output bytes to return"`
}

type LogResult struct {
	RootID    string      `json:"root_id"`
	Path      string      `json:"path,omitempty"`
	Limit     int         `json:"limit"`
	Commits   []LogCommit `json:"commits"`
	Truncated bool        `json:"truncated"`
}

type LogCommit struct {
	Hash       string `json:"hash"`
	ShortHash  string `json:"short_hash"`
	AuthorName string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
	AuthorDate string `json:"author_date"`
	Subject    string `json:"subject"`
	Body       string `json:"body,omitempty"`
}