# CLAUDE.md - MCP GitHub Server v4.0

**Status**: ✅ Production Ready | **Tools**: 26 (consolidated from 85) | **Architecture**: Hybrid (Local Git + GitHub API + Admin Controls)

## Core Architecture

**Go-based MCP server** providing GitHub integration for Claude Desktop via JSON-RPC 2.0.

### Key Files
- `cmd/github-mcp-server/main.go` - Entry point with multi-profile support
- `internal/server/server.go` - MCP handler with 26 consolidated tools
- `pkg/git/operations.go` - Local Git operations (0 tokens)
- `pkg/github/client.go` - GitHub API client wrapper with 11 methods
- `pkg/admin/admin.go` - Administrative operations client
- `internal/hybrid/operations.go` - Smart Git-first operations
- `pkg/dashboard/dashboard.go` - GitHub dashboard operations
- `pkg/safety/` - Safety layer with risk classification and audit
- `internal/server/admin_tools.go` - 4 consolidated admin tools (22 operations)
- `internal/server/admin_handlers.go` - Admin operation handlers
- `internal/server/safety_middleware.go` - Safety execution wrapper

### Design: Consolidated Operation Pattern
Tools use an `operation` parameter to expose multiple operations under one tool name.
This reduces the tool count from 85 to 26, preventing AI model confusion from tool reading limits.

Example: `git_branch` with `operation: checkout|checkout_remote|list|merge|rebase|backup`

Safety uses composite keys `tool:operation` (e.g., `github_webhooks:delete`) for risk classification.

### Features
- **Multi-profile**: Single executable, multiple GitHub accounts via `--profile` flag
- **Protocol auto-detection**: Automatically detects and responds with client's MCP protocol version
- **Hybrid ops**: Local Git first (0 tokens) → GitHub API fallback
- **26 consolidated tools**: All 85 operations accessible via operation parameter pattern
- **4-tier safety system**: Risk-based classification with confirmation tokens
- **Audit logging**: JSON-based operation tracking with automatic rotation
- **Dry-run mode**: Preview destructive operations before execution
- **Security**: Path traversal prevention, command injection protection, SSRF prevention

## Build & Test

```bash
.\compile.bat                    # Build
go test ./...                    # Run all tests
go test ./pkg/git/ -v           # Test specific package
```

## Dependencies
- Go 1.25.0+
- `github.com/google/go-github/v81 v81.0.0`
- `golang.org/x/oauth2 v0.34.0`

## Available Tools (26)

### Git Info (2 tools)
- **`git_info`** — operations: `status`, `file_sha`, `last_commit`, `file_content`, `changed_files`, `validate_repo`, `list_files`, `context`, `validate_clean`
- **`git_set_workspace`** — set working directory for all Git operations

### Git Basic (3 tools)
- **`git_init`** — initialize a new Git repository
- **`git_add`** — stage files for commit
- **`git_commit`** — commit staged changes

### Git Advanced (9 tools)
- **`git_history`** — operations: `log`, `diff`
- **`git_branch`** — operations: `checkout`, `checkout_remote`, `list`, `merge`, `rebase`, `backup`
- **`git_sync`** — operations: `push`, `pull`, `force_push`, `push_upstream`, `sync`, `pull_strategy`
- **`git_conflict`** — operations: `status`, `resolve`, `detect`, `safe_merge`
- **`git_stash`** — operations: `list`, `push`, `pop`, `apply`, `drop`, `clear`
- **`git_remote`** — operations: `list`, `add`, `remove`, `show`, `fetch`
- **`git_tag`** — operations: `list`, `create`, `delete`, `push`, `show`
- **`git_clean`** — operations: `untracked`, `untracked_dirs`, `ignored`, `all`
- **`git_reset`** — undo commits (soft/mixed/hard)

### Hybrid (3 tools)
- **`create_file`** — Git-first, API fallback
- **`update_file`** — Git-first, API fallback
- **`push_files`** — write multiple files + git add/commit/push

### GitHub API (1 tool)
- **`github_repo`** — operations: `list_repos`, `create_repo`, `list_prs`, `create_pr`

### Dashboard (1 tool)
- **`github_dashboard`** — operations: `full`, `notifications`, `issues`, `prs_review`, `security`, `workflows`, `mark_read`

### Response (1 tool)
- **`github_respond`** — operations: `comment_issue`, `comment_pr`, `review_pr`

### Repair (1 tool)
- **`github_repair`** — operations: `close_issue`, `merge_pr`, `rerun_workflow`, `dismiss_alert`

### Admin (4 tools)
- **`github_admin_repo`** — operations: `get_settings`, `update_settings`, `archive`, `delete`
- **`github_branch_protection`** — operations: `get`, `update`, `delete`
- **`github_webhooks`** — operations: `list`, `create`, `update`, `delete`, `test`
- **`github_collaborators`** — operations: `list`, `check`, `add`, `update_permission`, `remove`, `list_invitations`, `accept_invitation`, `cancel_invitation`, `list_teams`, `add_team`

### File Operations (1 tool)
- **`github_files`** — operations: `list`, `download`, `download_repo`, `pull_repo`

## Safety System

### 4-Tier Risk Classification (composite keys)
- **LOW (1)**: Read-only → Execute immediately (e.g., `github_admin_repo:get_settings`)
- **MEDIUM (2)**: Reversible changes → Dry-run in strict mode (e.g., `github_collaborators:add`)
- **HIGH (3)**: Impacts collaboration → Requires confirmation token (e.g., `github_webhooks:delete`)
- **CRITICAL (4)**: Irreversible → Confirmation + backup (e.g., `github_admin_repo:delete`)

### Safety Modes
- **Strict**: Confirms MEDIUM+ operations
- **Moderate**: Confirms HIGH+ operations (default)
- **Permissive**: Only confirms CRITICAL operations
- **Disabled**: No safety checks

### Configuration

Create `safety.json` in the server directory:

```json
{
  "mode": "moderate",
  "enable_audit_log": true,
  "require_confirmation_above": 3,
  "audit_log_path": "./mcp-admin-audit.log",
  "audit_log_max_size_mb": 10,
  "audit_log_max_backups": 5
}
```

## Token Permissions

**Required**: `repo` — all repository operations, webhooks, collaborators

**Optional**:
- `security_events` — security alert dismissal
- `admin:org` — organization team management
- `admin:repo_hook` — enhanced webhook management

### Multi-profile Setup

```json
{
  "mcpServers": {
    "github-personal": {
      "command": "path\\to\\github-mcp-server-v4.exe",
      "args": ["--profile", "personal"],
      "env": {"GITHUB_TOKEN": "ghp_token_personal"}
    },
    "github-work": {
      "command": "path\\to\\github-mcp-server-v4.exe",
      "args": ["--profile", "work"],
      "env": {"GITHUB_TOKEN": "ghp_token_work"}
    }
  }
}
```
