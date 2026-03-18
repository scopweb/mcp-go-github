# GitHub MCP Server v4.0

Go-based MCP server that connects GitHub to Claude Desktop, enabling direct repository operations from Claude's interface.

**Tools:** 26 consolidated tools (85 operations) | **Architecture:** Hybrid (Local Git + GitHub API + Admin Controls)

## What's New in v4.0

- **Consolidated Tool Design**: 85 operations across just 26 tools — prevents AI confusion from tool-count limits
- **Operation Parameter Pattern**: Each tool accepts an `operation` parameter to select the specific action
- **`--toolsets` Flag**: Start the server exposing only selected tool groups (`git`, `github`, `admin`, `files`)
- **Real Auto-Backup**: Writes a JSON backup before HIGH/CRITICAL operations when `enable_auto_backup: true`
- **4-Tier Safety System**: Risk classification (LOW/MEDIUM/HIGH/CRITICAL) with confirmation tokens
- **22 Administrative Operations**: Repository settings, branch protection, webhooks, collaborators, teams
- **Audit Logging**: JSON-based operation tracking with automatic rotation

## Token Permissions Required

### Minimum Required

```
repo        - Full control of private repositories (essential)
```

### Optional (for full functionality)

```
security_events  - For dismissing security alerts
workflow         - For re-running GitHub Actions workflows
admin:repo_hook  - Enhanced webhook management
admin:org        - For team management in organizations
```

### Generate Token

1. Go to: [GitHub Settings > Personal Access Tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Select the required scopes
4. Copy the generated token

## Installation

```bash
# Install dependencies
go mod tidy

# Compile (using included script)
.\compile.bat          # Windows
./build-mac.bat        # macOS cross-compile

# Or compile manually
go build -o github-mcp-server-v4.exe ./cmd/github-mcp-server/
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests for a specific package
go test ./pkg/git/ -v
go test ./pkg/safety/ -v
```

## Claude Desktop Configuration

### Multi-profile (Recommended)

```json
{
  "mcpServers": {
    "github-personal": {
      "command": "C:\\path\\to\\github-mcp-server-v4.exe",
      "args": ["--profile", "personal"],
      "env": {
        "GITHUB_TOKEN": "ghp_your_personal_token"
      }
    },
    "github-work": {
      "command": "C:\\path\\to\\github-mcp-server-v4.exe",
      "args": ["--profile", "work"],
      "env": {
        "GITHUB_TOKEN": "ghp_your_work_token"
      }
    }
  }
}
```

### Basic Configuration (Single token)

```json
{
  "mcpServers": {
    "github-mcp": {
      "command": "C:\\path\\to\\github-mcp-server-v4.exe",
      "args": [],
      "env": {
        "GITHUB_TOKEN": "your_token_here"
      }
    }
  }
}
```

### Toolset Filtering (Optional)

Expose only specific tool groups to reduce attack surface:

```json
{
  "args": ["--toolsets", "git,github"]
}
```

Available groups: `git` (14 tools), `github` (4 tools), `admin` (4 tools), `files` (4 tools). Default is `all`.

## Available Tools (26)

Tools use an `operation` parameter to expose multiple operations under one name. This reduces the tool count from 85 to 26, preventing AI model confusion.

### Git Info (2 tools)

| Tool | Operations |
|------|-----------|
| `git_info` | `status`, `file_sha`, `last_commit`, `file_content`, `changed_files`, `validate_repo`, `list_files`, `context`, `validate_clean` |
| `git_set_workspace` | Set working directory for all Git operations |

### Git Basic (3 tools)

| Tool | Description |
|------|-------------|
| `git_init` | Initialize a new Git repository |
| `git_add` | Stage files for commit |
| `git_commit` | Commit staged changes |

### Git Advanced (9 tools)

| Tool | Operations |
|------|-----------|
| `git_history` | `log`, `diff` |
| `git_branch` | `checkout`, `checkout_remote`, `list`, `merge`, `rebase`, `backup` |
| `git_sync` | `push`, `pull`, `force_push`, `push_upstream`, `sync`, `pull_strategy` |
| `git_conflict` | `status`, `resolve`, `detect`, `safe_merge` |
| `git_stash` | `list`, `push`, `pop`, `apply`, `drop`, `clear` |
| `git_remote` | `list`, `add`, `remove`, `show`, `fetch` |
| `git_tag` | `list`, `create`, `delete`, `push`, `show` |
| `git_clean` | `untracked`, `untracked_dirs`, `ignored`, `all` |
| `git_reset` | Undo commits (soft/mixed/hard) |

### Hybrid (3 tools)

| Tool | Description |
|------|-------------|
| `create_file` | Create file — Git-first, GitHub API fallback |
| `update_file` | Update file — Git-first, GitHub API fallback |
| `push_files` | Write multiple files + git add/commit/push in one call |

### GitHub API (1 tool)

| Tool | Operations |
|------|-----------|
| `github_repo` | `list_repos`, `create_repo`, `list_prs`, `create_pr` |

### Dashboard (1 tool)

| Tool | Operations |
|------|-----------|
| `github_dashboard` | `full`, `notifications`, `issues`, `prs_review`, `security`, `workflows`, `mark_read` |

### Response (1 tool)

| Tool | Operations |
|------|-----------|
| `github_respond` | `comment_issue`, `comment_pr`, `review_pr` |

### Repair (1 tool)

| Tool | Operations |
|------|-----------|
| `github_repair` | `close_issue`, `merge_pr`, `rerun_workflow`, `dismiss_alert` |

### Admin (4 tools)

| Tool | Operations | Risk |
|------|-----------|------|
| `github_admin_repo` | `get_settings`, `update_settings`, `archive`, `delete` | LOW → CRITICAL |
| `github_branch_protection` | `get`, `update`, `delete` | LOW → CRITICAL |
| `github_webhooks` | `list`, `create`, `update`, `delete`, `test` | LOW → HIGH |
| `github_collaborators` | `list`, `check`, `add`, `update_permission`, `remove`, `list_invitations`, `accept_invitation`, `cancel_invitation`, `list_teams`, `add_team` | LOW → HIGH |

### File Operations (1 tool)

| Tool | Operations |
|------|-----------|
| `github_files` | `list`, `download`, `download_repo`, `pull_repo` |

## Safety System

### 4-Tier Risk Classification

| Level | Description | Behavior (moderate mode) |
|-------|-------------|-------------------------|
| **LOW** | Read-only | Direct execution |
| **MEDIUM** | Reversible changes | Optional dry-run |
| **HIGH** | Impacts collaboration | Requires confirmation token |
| **CRITICAL** | Irreversible | Token + backup recommendation |

Safety uses composite keys `tool:operation` (e.g., `github_webhooks:delete`) for risk classification.

### Safety Modes

| Mode | Confirms from | Recommended use |
|------|--------------|-----------------|
| `strict` | MEDIUM+ | Critical production environments |
| `moderate` | HIGH+ | General use (default) |
| `permissive` | CRITICAL | Local development |
| `disabled` | Never | Not recommended |

### Safety Configuration

Create `safety.json` next to the executable (optional):

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

If `safety.json` doesn't exist, defaults to **moderate mode** with audit logging enabled.

## Git-Free Mode

On systems without Git installed, the server:

1. Automatically detects Git absence
2. Filters `git_*` tools from the listing
3. Keeps all API, admin, dashboard, and file operation tools operational
4. Returns a friendly error if a Git tool is attempted

The `github_files` tool (`list`, `download`, `download_repo`, `pull_repo`) allows cloning and updating repositories using only the GitHub API, without Git.

## Security

- Command injection prevention in all Git operations
- Path traversal defense
- SSRF prevention in webhook URLs
- Cryptographic confirmation tokens for destructive operations
- Audit logging of all administrative operations

For vulnerability reports, see our [Security Policy](SECURITY.md).

## System Requirements

- **Go**: 1.25.0 or higher
- **Git**: Optional (auto-detected; file/API tools work without it)
- `github.com/google/go-github` v81.0.0
- `golang.org/x/oauth2` v0.36.0
- **GitHub Token**: Minimum `repo` permission

## Project Status

- 26 consolidated tools exposing 85 operations
- Hybrid local Git + GitHub API system
- 4-tier safety system with confirmation tokens and auto-backup
- Multi-profile support
- Complete test suite
- Production ready (v4.0)

**Changelog**: See [CHANGELOG.md](CHANGELOG.md) for complete change history.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License — see [LICENSE](LICENSE) for details.
