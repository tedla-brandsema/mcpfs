# mcpfs

`mcpfs` is a small, read-only Model Context Protocol filesystem server.

It lets an MCP client inspect explicitly configured project folders without uploading files or copying snippets into chat. It is designed for live developer context: list files, read files, search source, inspect Git status, view diffs, and review recent commit history.

The core idea is simple:

```text
configured roots
  → safe path resolver
  → .gitignore-aware matcher
  → read-only filesystem tools
  → read-only git inspection tools
  → MCP
```

## Features

* Explicit JSON-configured filesystem roots.
* Read-only filesystem access.
* `.gitignore` support.
* Additional include/exclude glob rules.
* Root escape protection.
* Symlink escape protection.
* File size limits.
* Structured JSON logging.
* STDIO transport for local MCP hosts and MCP Inspector.
* Streamable HTTP transport for remote MCP clients.
* Optional embedded ngrok tunnel for quick remote development.
* HTTP auth modes: `none`, `bearer`, and `oidc`.
* OIDC/JWT validation using JWKS, issuer, audience, expiry, not-before, and identity allowlists.
* MCP tool annotations mark exposed tools as read-only.
* Read-only Git inspection tools:

  * `git_status`
  * `git_diff`
  * `git_log`

## MCP tools

All exposed tools are read-only and are annotated with MCP `readOnlyHint`.

Filesystem tools:

| Tool        | Description                                         |
| ----------- | --------------------------------------------------- |
| `fs_roots`  | List configured filesystem roots.                   |
| `fs_list`   | List files and directories under a configured root. |
| `fs_read`   | Read a file under a configured root.                |
| `fs_search` | Search text files under a configured root.          |
| `fs_stat`   | Return metadata for a file or directory.            |

Git tools:

| Tool         | Description                                                                                                         |
| ------------ | ------------------------------------------------------------------------------------------------------------------- |
| `git_status` | Return `git status --porcelain=v1 -b` as structured JSON.                                                           |
| `git_diff`   | Return a diff for the whole root or a specific path. Supports staged diffs and synthetic diffs for untracked files. |
| `git_log`    | Return recent commit history, optionally scoped to a path.                                                          |

## Security model

`mcpfs` is designed to expose only the authority you explicitly configure.

Tool paths are always interpreted relative to a configured root. `mcpfs` rejects:

* absolute paths
* `..` root escapes
* symlink escapes
* explicitly excluded paths
* `.gitignore` ignored paths
* files larger than the configured root limit

`.gitignore` support is an additional policy layer. It never weakens the root boundary.

The HTTP transports support three auth modes:

* `none` — no HTTP authentication. Useful only for local-only setups or short-lived development tunnels.
* `bearer` — static bearer token loaded from an environment variable.
* `oidc` — JWT/OIDC validation using a configured issuer, audience, JWKS URL, and identity allowlist.

All MCP tools are read-only. No write tools are exposed.

Do not expose `mcpfs` to the public internet without understanding what roots you configured and which auth mode is active.

## Installation

```bash
git clone https://github.com/tedla-brandsema/mcpfs.git
cd mcpfs

go test ./...
go build -o ./bin/mcpfs ./cmd/mcpfs
```

## STDIO usage

STDIO is useful for local MCP clients and MCP Inspector.

```bash
./bin/mcpfs -config config.example.json
```

Or run directly:

```bash
go run ./cmd/mcpfs -config config.example.json
```

### MCP Inspector

```bash
bunx @modelcontextprotocol/inspector
```

Use:

```text
Transport Type:
STDIO

Command:
go

Arguments:
run
/path/to/mcpfs/cmd/mcpfs
-config
/path/to/mcpfs/config.example.json
```

Or with a built binary:

```text
Transport Type:
STDIO

Command:
/path/to/mcpfs/bin/mcpfs

Arguments:
-config
/path/to/mcpfs/config.example.json
```

## HTTP usage

HTTP transport is useful when the MCP client needs to connect to a network endpoint.

Example config using bearer auth:

```json
{
  "server": {
    "name": "mcpfs",
    "version": "0.2.0",
    "transport": "http",
    "addr": "127.0.0.1:8080",
    "path": "/mcp",
    "auth": {
      "mode": "bearer",
      "token_env": "MCPFS_TOKEN"
    }
  },
  "roots": [
    {
      "id": "project",
      "path": "/path/to/project",
      "mode": "read",
      "include": [
        "**/*.go",
        "**/*.md",
        "**/*.mod",
        "**/*.sum",
        "**/*.json",
        "**/*.yaml",
        "**/*.yml"
      ],
      "exclude": [
        "**/.git/**",
        "**/.env",
        "**/.env.*",
        "**/*secret*",
        "**/*credential*",
        "**/*.pem",
        "**/*.key"
      ],
      "use_gitignore": true,
      "max_file_bytes": 262144
    }
  ]
}
```

Run:

```bash
export MCPFS_TOKEN="$(openssl rand -hex 32)"
./bin/mcpfs -config config.http.example.json
```

Health check:

```bash
curl http://127.0.0.1:8080/healthz
```

Authenticated MCP endpoint:

```bash
curl -i \
  -H "Authorization: Bearer $MCPFS_TOKEN" \
  http://127.0.0.1:8080/mcp
```

A plain `GET /mcp` may return a protocol-level MCP response such as `405 Method Not Allowed`. That is expected; it means the request reached the MCP handler.

## OIDC/JWT auth

`mcpfs` can run as an OIDC/JWT-protected MCP resource server.

It does not implement an OAuth authorization server. Instead, it validates bearer JWTs issued by an external identity provider such as Google Identity Platform, Firebase, Auth0, WorkOS, Zitadel, Keycloak, Dex, or Azure Entra ID.

Example config:

```json
{
  "server": {
    "name": "mcpfs",
    "version": "0.2.0",
    "transport": "http",
    "addr": "127.0.0.1:8080",
    "path": "/mcp",
    "auth": {
      "mode": "oidc",
      "issuer": "https://accounts.google.com",
      "audience": "mcpfs",
      "jwks_url": "https://www.googleapis.com/oauth2/v3/certs",
      "allowed_emails": ["you@example.com"]
    }
  },
  "roots": [
    {
      "id": "project",
      "path": "/path/to/project",
      "mode": "read",
      "include": ["**/*"],
      "exclude": ["**/.git/**", "**/.env", "**/.env.*"],
      "use_gitignore": true,
      "max_file_bytes": 262144
    }
  ]
}
```

OIDC validation checks:

* JWT signature using the configured JWKS URL.
* `iss`
* `aud`
* `exp`
* `nbf`, when present
* `allowed_emails` and/or `allowed_subjects`

At least one of `allowed_emails` or `allowed_subjects` must be configured.

Example using subject allowlisting instead of email allowlisting:

```json
"auth": {
  "mode": "oidc",
  "issuer": "https://issuer.example.com",
  "audience": "mcpfs",
  "jwks_url": "https://issuer.example.com/.well-known/jwks.json",
  "allowed_subjects": ["user-or-client-subject-id"]
}
```

## Embedded ngrok usage

`mcpfs` can start an ngrok tunnel automatically for quick remote development.

First, install/configure an ngrok account and create an authtoken.

Run:

```bash
export NGROK_AUTHTOKEN="<your-ngrok-authtoken>"
./scripts/run-ngrok.sh
```

The app logs a public MCP URL:

```text
https://example.ngrok-free.dev/mcp
```

Use that URL in your remote MCP client.

For short-lived development with ChatGPT Developer Mode, `config.ngrok.example.json` uses:

```json
"auth": {
  "mode": "none"
}
```

Only do this while actively testing, and only with carefully scoped roots.

You can also run ngrok with bearer or OIDC auth by changing the `auth` block in the ngrok config.

## Configuration

Root config fields:

| Field            | Description                                                        |
| ---------------- | ------------------------------------------------------------------ |
| `id`             | Stable root identifier used by MCP tools.                          |
| `path`           | Local filesystem path to expose.                                   |
| `mode`           | Currently accepts `read` or `read_write`, but tools are read-only. |
| `include`        | Glob allowlist. Empty means all files not excluded are allowed.    |
| `exclude`        | Glob denylist.                                                     |
| `use_gitignore`  | Apply `.gitignore` rules as an additional deny/allow layer.        |
| `max_file_bytes` | Maximum readable file size. Defaults to `262144` when set to `0`.  |

Server config fields:

| Field                   | Description                                                           |
| ----------------------- | --------------------------------------------------------------------- |
| `name`                  | MCP server name.                                                      |
| `version`               | MCP server version.                                                   |
| `transport`             | `stdio`, `http`, or `http_ngrok`.                                     |
| `addr`                  | HTTP bind address. Defaults to `127.0.0.1:8080` for HTTP transports.  |
| `path`                  | MCP HTTP path. Defaults to `/mcp`.                                    |
| `auth.mode`             | HTTP auth mode: `none`, `bearer`, or `oidc`.                          |
| `auth.token_env`        | Environment variable containing the bearer token when using `bearer`. |
| `auth.issuer`           | Expected JWT issuer when using `oidc`.                                |
| `auth.audience`         | Expected JWT audience when using `oidc`.                              |
| `auth.jwks_url`         | JWKS URL used to verify JWT signatures when using `oidc`.             |
| `auth.allowed_emails`   | Optional email allowlist for OIDC-authenticated users.                |
| `auth.allowed_subjects` | Optional subject allowlist for OIDC-authenticated users.              |
| `ngrok_url`             | Optional reserved ngrok URL/domain.                                   |

Legacy `require_auth` and `auth_token_env` fields are still accepted for compatibility, but new configs should use `auth.mode`.

### Auth examples

No auth:

```json
"auth": {
  "mode": "none"
}
```

Bearer auth:

```json
"auth": {
  "mode": "bearer",
  "token_env": "MCPFS_TOKEN"
}
```

OIDC auth:

```json
"auth": {
  "mode": "oidc",
  "issuer": "https://accounts.google.com",
  "audience": "mcpfs",
  "jwks_url": "https://www.googleapis.com/oauth2/v3/certs",
  "allowed_emails": ["you@example.com"]
}
```

## Example workflows

Ask for the configured roots:

```json
{}
```

Call `fs_roots`.

List project files:

```json
{
  "root_id": "project",
  "path": ".",
  "recursive": false
}
```

Search code:

```json
{
  "root_id": "project",
  "query": "ResolveInsideRoot",
  "glob": "**/*.go"
}
```

Read a file:

```json
{
  "root_id": "project",
  "path": "internal/core/resolve.go"
}
```

Check working tree state:

```json
{
  "root_id": "project"
}
```

Call `git_status`.

Inspect a changed file:

```json
{
  "root_id": "project",
  "path": "internal/service/git/diff.go"
}
```

Call `git_diff`.

Review history:

```json
{
  "root_id": "project",
  "limit": 5
}
```

Call `git_log`.

## Development

Run tests:

```bash
go test ./...
```

Build:

```bash
go build -o ./bin/mcpfs ./cmd/mcpfs
```

Run with STDIO:

```bash
./bin/mcpfs -config config.example.json
```

Run with HTTP bearer auth:

```bash
export MCPFS_TOKEN="$(openssl rand -hex 32)"
./bin/mcpfs -config config.http.example.json
```

Run with HTTP OIDC auth:

```bash
./bin/mcpfs -config config.oidc.example.json
```

Run with embedded ngrok:

```bash
export NGROK_AUTHTOKEN="<your-ngrok-authtoken>"
./scripts/run-ngrok.sh
```

## Publishing checklist

Before pushing to GitHub:

```bash
go test ./...
git status
git grep -nE 'NGROK_AUTHTOKEN|MCPFS_TOKEN|Bearer |ngrok-free|password|secret|credential|BEGIN .*PRIVATE KEY|api[_-]?key'
git status --ignored --short
```

Also check that example configs contain only placeholder values.

## Status

`mcpfs` is currently read-only. The `mode` field accepts `read_write` for future compatibility, but no write tools are exposed.

The exposed MCP tools are annotated as read-only.
