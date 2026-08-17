# AGENTS.md

Trust executable sources over README/CLAUDE.md when they disagree. `go.mod` is Go **1.26.6** and `github.com/google/go-github/v88`. `VERSION` is **4.1.0**.

Broader product docs: `CLAUDE.md`, `CONTRIBUTING.md`, `CHANGELOG.md`. This file is only what an agent would otherwise get wrong.

## Commands

```bash
go test ./...                  # what CI runs
go test ./pkg/git/ -v          # one package
go test -race ./...            # script/test (stricter than CI)
go build ./...                 # what CI builds
golangci-lint run              # what CI lints (.golangci.yml)
.\compile.bat                  # Windows binary + VERSION ldflags → github-mcp-server-v4.exe
```

- Unflagged `go build` reports server version `"dev"`. Release builds inject `internal/server.Version` from `VERSION`.
- `script/lint` and `script/test` are bash. On Windows PowerShell use the commands above, not those scripts.
- CI (`.github/workflows/ci.yml`): `go build ./...` → `go test ./... -v -timeout 60s` → `golangci-lint` latest.
- PRs: Conventional Commits, update `CHANGELOG.md`, no hardcoded paths/tokens.

Do **not** build with `-mod=vendor`. `vendor/` is stale (`go-github/v81`); `go.mod` is v88.

## Layout

Custom MCP over stdio JSON-RPC (not `mark3labs/mcp-go`).

| Path | Role |
|------|------|
| `cmd/github-mcp-server/main.go` | Flags, clients, stdin loop |
| `internal/server/` | Protocol + tool dispatch. **No tests here.** |
| `internal/hybrid/` | Git-first file writes, API fallback |
| `pkg/git/` | Local git via `exec.Command("git", args...)` (no shell) |
| `pkg/github/`, `pkg/admin/`, `pkg/dashboard/` | GitHub API wrappers |
| `pkg/safety/` | Risk map, tokens, validators, audit |
| `pkg/config/` | Loads `safety.json` |
| `pkg/interfaces/` | Update this first when adding client methods |

`ejemplos_github_repositorios/` and `test/security/README.md` are leftovers from other projects. Ignore them.

## Adding or changing tools

Do **not** add a new top-level tool if an existing one can take another `operation`. That 26-tool cap is intentional.

1. Schema in `internal/server/tool_definitions_*.go` (or `admin_tools.go` / `file_tools.go`).
2. Dispatch in `CallTool` (`server.go`) or `HandleAdminTool` (`admin_handlers.go`).
3. Register in the right `ListTools()` group — **hybrid `gh_*` tools live in toolset `files`, not `github`**:
   - `git` — all `git_*` (14)
   - `files` — `gh_create_file`, `gh_update_file`, `gh_push_files`, `github_files` (4)
   - `github` — `github_repo`, `github_dashboard`, `github_respond`, `github_repair` (4)
   - `admin` — 4 admin tools
4. Hybrid file tools **must** keep the `gh_` prefix (collision with filesystem MCP servers).
5. Admin mutations: add `"tool:operation"` to `operationRiskMap` in `pkg/safety/risk_classifier.go`, validators if needed, and wrap with `Safety.WrapExecution(ctx, "tool:op", args, ...)`.
6. Parse JSON numbers with `getIntArg` / `getInt64Arg` / `getStringArg` in `args.go`. Never `arguments["n"].(float64)`.

`isGitTool` is prefix `git_` only. When git is missing, `git_*` are hidden; `gh_*` stay listed and fall back to the API.

## Safety config (code, not docs)

`pkg/config.LoadConfig` expects nested camelCase. Flat `mode` / `enable_audit_log` keys in README, CLAUDE.md, and `safety.json.example` are **ignored**.

```json
{
  "version": "3.0",
  "safetyMode": "moderate",
  "globalSettings": {
    "enableAuditLog": true,
    "auditLogPath": "./mcp-admin-audit.log",
    "requireConfirmationAbove": "high",
    "enableAutoBackup": true,
    "backupPath": "./.mcp-backups"
  }
}
```

Missing file → moderate defaults. `--profile=foo` prefers `./safety.foo.json`, else `./safety.json`. There is no `--safety-config` flag.

Only admin tools go through safety. Tokens are `CONF:` + 12 hex, 5 minutes, single-use, bound to parameters. Backup failure logs and continues.

## Git client

- `enterWorkingDir` uses process-global `os.Chdir`. Do not add concurrent git work.
- Git tests mock by **full command string** (`"git status --porcelain"`). New invocations need a matching mock key.
- `git_set_workspace` changes the repo used by later `git_*` calls.

## Protocol

- One JSON-RPC object per stdin line. Scanner max 10MB.
- `notifications/*` with no `id` must not write a response.
- `initialize` echoes the client's `protocolVersion` (default `2025-11-25`).

## Secrets / local junk

`.env` is gitignored. `.mcp.json` is **not** — never put `GITHUB_TOKEN` there. `safety.json`, `.mcp-backups/`, and `mcp-admin-audit.log*` are gitignored.
