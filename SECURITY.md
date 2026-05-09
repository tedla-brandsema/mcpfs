# Security

MCPFS is a power tool.

When configured with writable roots or command execution, connected MCP clients can modify files and run programs on your machine. Treat access to an MCPFS server like access to your terminal.

## High-risk configuration

The following settings intentionally grant powerful capabilities:

* root `mode: "read_write"`
* `commands.mode: "predefined"`
* `commands.mode: "unguarded"`
* HTTP or ngrok transports exposed outside your local machine
* `auth.mode: "none"` on network transports

`commands.mode: "unguarded"` exposes `cmd_exec`, which allows connected MCP clients to run arbitrary argv commands from root-scoped working directories. This is terminal-level authority.

## Recommendations

* Run MCPFS locally or only on networks you control.
* Connect only MCP clients you trust.
* Keep roots as narrow as practical.
* Prefer `mode: "read"` unless write access is required.
* Prefer `commands.mode: "disabled"` or `commands.mode: "predefined"` unless arbitrary command execution is required.
* Use bearer or OIDC auth for HTTP transports.
* Do not expose MCPFS with `auth.mode: "none"` to untrusted networks.
* Review config files before starting the server.
* Treat project scripts and build tools as executable code.

## Reporting vulnerabilities

Please report security issues privately through GitHub Security Advisories if available, or contact the maintainer directly.

Do not open public issues for vulnerabilities that could put users at risk.
