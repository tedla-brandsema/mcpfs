# Changelog

## v0.4.0

### Added

* Opt-in filesystem write support through `fs_write` for roots configured with `mode: "read_write"`.
* Writable path resolution that checks existing parent symlinks before creating new files.
* Command execution framework with explicit modes:
  * `disabled`
  * `predefined`
  * `unguarded`
* `cmd_list` for listing configured command IDs.
* `cmd_run` for running predefined commands by ID.
* `cmd_exec` for arbitrary argv command execution in `unguarded` mode.
* Command execution timeouts, output limits, stdout/stderr capture, exit code metadata, timeout metadata, truncation metadata, and structured logs.
* Project-local config support through `.mcpfs/project.cfg.json`.
* Global MCPFS config bootstrap from an embedded default config.
* `mcpfs init` for writing project-local config.
* `mcpfs project add`, `mcpfs project rm`, and `mcpfs project ls` for managing configured roots.

### Changed

* MCPFS is now positioned as a power-user local MCP workbench rather than only a read-only filesystem bridge.
* README now documents the power-tool warning, write access, command execution modes, `cmd_exec`, and release-era usage examples.
* Default/global config version is now `0.4.0`.

### Security

* Added explicit warning language for write and command execution capabilities.
* Command execution is disabled by default.
* `cmd_exec` is registered only when `commands.mode` is `unguarded`.
* `cmd_run` and `cmd_exec` execute argv arrays directly and do not perform shell interpolation unless the configured/client-provided argv explicitly invokes a shell.

## v0.3.0

### Added

* `fs_tree` for bounded tree output.
* `fs_read_lines` for line-range reads.
* `fs_search_regex` for regex-based search.
* `git_show` for commit inspection.
* `git_blame` for blame inspection.
* `project_overview` for compact project summaries.
* Tool result metadata consistency across services.
* Global config bootstrap and project overview registry support.
