# GitHub MCP Server v3.0

Go-based MCP server that connects GitHub to Claude Desktop, enabling direct repository operations from Claude's interface.

**Tools:** 82 (with Git) | 48 (without Git) | **Architecture:** Hybrid (Local Git + GitHub API + Admin Controls)

## What's New in v3.0

- **22 Administrative Tools**: Repository settings, branch protection, webhooks, collaborators, teams
- **4-Tier Safety System**: Risk classification (LOW/MEDIUM/HIGH/CRITICAL) with confirmation tokens
- **Git-Free File Operations**: Clone, pull, download repos via GitHub API (no Git required)
- **Smart Git Detection**: Auto-detects Git availability, filters tools accordingly
- **Audit Logging**: JSON-based operation tracking with automatic rotation

## Token Permissions Required

### Minimum Required:
```
repo        - Full control of private repositories (essential)
```

### Optional (for full functionality):
```
delete_repo      - For github_delete_repository
workflow         - For re-running GitHub Actions workflows
security_events  - For dismissing security alerts
admin:repo_hook  - Enhanced webhook management (v3.0)
admin:org        - For team management in organizations (v3.0)
```

### Generate Token:
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
./build-mac.bat        # macOS/Linux

# Or compile manually
go build -o mcp-go-github.exe ./cmd/github-mcp-server/
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
      "command": "C:\\path\\to\\mcp-go-github.exe",
      "args": ["--profile", "personal"],
      "env": {
        "GITHUB_TOKEN": "ghp_your_personal_token"
      }
    },
    "github-work": {
      "command": "C:\\path\\to\\mcp-go-github.exe",
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
      "command": "C:\\path\\to\\mcp-go-github.exe",
      "args": [],
      "env": {
        "GITHUB_TOKEN": "your_token_here"
      }
    }
  }
}
```

## Available Tools (82 Tools)

### Git Information (8)

| Tool | Description | Tokens |
|------|-------------|--------|
| `git_status` | Local Git repository status | 0 |
| `git_list_files` | List all files in repository | 0 |
| `git_get_file_content` | Get file content from Git | 0 |
| `git_get_file_sha` | Get SHA of specific file | 0 |
| `git_get_last_commit` | Get last commit SHA | 0 |
| `git_get_changed_files` | List modified files | 0 |
| `git_validate_repo` | Validate if directory is a valid Git repo | 0 |
| `git_context` | Auto-detect Git context | 0 |

### Basic Git Operations (6)

| Tool | Description | Tokens |
|------|-------------|--------|
| `git_set_workspace` | Set working directory | 0 |
| `git_add` | Add files to staging area | 0 |
| `git_commit` | Commit changes | 0 |
| `git_push` | Push changes to remote | 0 |
| `git_pull` | Pull changes from remote | 0 |
| `git_checkout` | Switch branch or create new | 0 |

### Git Analysis & Management (7)

| Tool | Description | Tokens |
|------|-------------|--------|
| `git_log_analysis` | Commit history analysis | 0 |
| `git_diff_files` | Show modified files with statistics | 0 |
| `git_branch_list` | List branches with detailed info | 0 |
| `git_stash` | Stash operations | 0 |
| `git_remote` | Remote repository management | 0 |
| `git_tag` | Tag management | 0 |
| `git_clean` | Clean untracked files | 0 |

### Advanced Git Operations (7)

| Tool | Description | Tokens |
|------|-------------|--------|
| `git_checkout_remote` | Checkout remote branch with tracking | 0 |
| `git_merge` | Merge branches with validation | 0 |
| `git_rebase` | Rebase with specified branch | 0 |
| `git_pull_with_strategy` | Pull with strategies | 0 |
| `git_force_push` | Push with --force-with-lease | 0 |
| `git_push_upstream` | Push setting upstream | 0 |
| `git_sync_with_remote` | Synchronize with remote branch | 0 |

### Conflict Management (6)

| Tool | Description | Tokens |
|------|-------------|--------|
| `git_safe_merge` | Safe merge with backup | 0 |
| `git_conflict_status` | Conflict status | 0 |
| `git_resolve_conflicts` | Automatic resolution | 0 |
| `git_validate_clean_state` | Validate clean working directory | 0 |
| `git_detect_conflicts` | Detect potential conflicts | 0 |
| `git_create_backup` | Create backup of current state | 0 |

### Hybrid Operations (2)

| Tool | Description | Tokens |
|------|-------------|--------|
| `create_file` | Create file (local Git, API fallback) | 0* |
| `update_file` | Update file (local Git, API fallback) | 0* |

### GitHub API (4)

| Tool | Description |
|------|-------------|
| `github_list_repos` | List user repositories |
| `github_create_repo` | Create new repository |
| `github_list_prs` | List pull requests |
| `github_create_pr` | Create new pull request |

### File Operations - Git-Free (4) [NEW v3.0]

| Tool | Description |
|------|-------------|
| `github_list_repo_contents` | List files and directories via API |
| `github_download_file` | Download individual file |
| `github_download_repo` | Clone complete repository via API |
| `github_pull_repo` | Update local directory via API |

### Dashboard (7)

| Tool | Description |
|------|-------------|
| `github_dashboard` | General activity panel |
| `github_notifications` | Pending notifications |
| `github_assigned_issues` | Assigned issues |
| `github_prs_to_review` | PRs pending review |
| `github_security_alerts` | Security alerts |
| `github_failed_workflows` | Failed workflows |
| `github_mark_notification_read` | Mark notification as read |

### Response (3)

| Tool | Description |
|------|-------------|
| `github_comment_issue` | Comment on issue |
| `github_comment_pr` | Comment on pull request |
| `github_review_pr` | Create PR review (APPROVE/REQUEST_CHANGES/COMMENT) |

### Repair (6)

| Tool | Description |
|------|-------------|
| `github_close_issue` | Close issue |
| `github_merge_pr` | Merge pull request |
| `github_rerun_workflow` | Re-run workflow |
| `github_dismiss_dependabot_alert` | Dismiss Dependabot alert |
| `github_dismiss_code_alert` | Dismiss Code Scanning alert |
| `github_dismiss_secret_alert` | Dismiss Secret Scanning alert |

### Repository Admin (4) [NEW v3.0]

| Tool | Risk | Description |
|------|------|-------------|
| `github_get_repo_settings` | LOW | View repository configuration |
| `github_update_repo_settings` | MEDIUM | Modify name, description, visibility |
| `github_archive_repository` | CRITICAL | Archive repository (read-only) |
| `github_delete_repository` | CRITICAL | Delete repository PERMANENTLY |

### Branch Protection (3) [NEW v3.0]

| Tool | Risk | Description |
|------|------|-------------|
| `github_get_branch_protection` | LOW | View protection rules |
| `github_update_branch_protection` | HIGH | Configure protection rules |
| `github_delete_branch_protection` | CRITICAL | Remove branch protection |

### Webhooks (5) [NEW v3.0]

| Tool | Risk | Description |
|------|------|-------------|
| `github_list_webhooks` | LOW | List repository webhooks |
| `github_create_webhook` | MEDIUM | Create webhook |
| `github_update_webhook` | MEDIUM | Modify webhook |
| `github_delete_webhook` | HIGH | Delete webhook |
| `github_test_webhook` | LOW | Send test delivery |

### Collaborators (8) [NEW v3.0]

| Tool | Risk | Description |
|------|------|-------------|
| `github_list_collaborators` | LOW | List collaborators |
| `github_check_collaborator` | LOW | Verify access |
| `github_add_collaborator` | MEDIUM | Invite with permissions |
| `github_update_collaborator_permission` | MEDIUM | Change access level |
| `github_remove_collaborator` | HIGH | Revoke access |
| `github_list_invitations` | LOW | View pending invitations |
| `github_accept_invitation` | MEDIUM | Accept invitation |
| `github_cancel_invitation` | MEDIUM | Cancel invitation |

### Teams (2) [NEW v3.0]

| Tool | Risk | Description |
|------|------|-------------|
| `github_list_repo_teams` | LOW | List teams with access |
| `github_add_repo_team` | MEDIUM | Grant team access |

## Safety System (v3.0)

### 4 Risk Levels

| Level | Description | Behavior (moderate mode) |
|-------|-------------|-------------------------|
| **LOW** | Read-only | Direct execution |
| **MEDIUM** | Reversible changes | Optional dry-run |
| **HIGH** | Impacts collaboration | Requires confirmation token |
| **CRITICAL** | Irreversible | Token + backup recommendation |

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

See `safety.json.example` for complete configuration reference.

## Git-Free Mode (v3.0)

On systems without Git installed (e.g., Mac without Xcode Command Line Tools), the server:

1. **Automatically detects** Git absence
2. **Filters** git_ tools from listing
3. **Keeps operational** all API, admin, dashboard, file operation tools
4. **Returns friendly error** if a Git tool is attempted

The 4 File Operations tools (`github_list_repo_contents`, `github_download_file`, `github_download_repo`, `github_pull_repo`) allow cloning and updating repositories using only the GitHub API, without Git.

## Security

- Prevention of argument injection in Git commands
- Path Traversal defense
- Strict user input validation
- SSRF prevention in webhook URLs (v3.0)
- Cryptographic confirmation tokens for destructive operations (v3.0)
- Audit logging of administrative operations (v3.0)

## System Requirements

- **Go**: 1.25.0 or higher
- **Git**: Optional (auto-detected, 48 tools work without Git)
- **github.com/google/go-github**: v81.0.0
- **golang.org/x/oauth2**: v0.34.0
- **GitHub Token**: Minimum `repo` permission

## Project Status

- 82 operational MCP tools (48 without Git)
- Hybrid local Git + GitHub API system
- 22 administrative tools with safety layer
- 4 Git-free file tools
- Multi-profile support
- Complete testing with real repository
- Production ready (v3.0)

**Changelog**: See [CHANGELOG.md](CHANGELOG.md) for complete change history

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Security

For security issues, please see our [Security Policy](SECURITY.md).
