# CLAUDE.md - MCP GitHub Server v3.0

**Status**: ✅ Production Ready | **Tools**: 77+ | **Architecture**: Hybrid (Local Git + GitHub API + Admin Controls)

## Core Architecture

**Go-based MCP server** providing GitHub integration for Claude Desktop via JSON-RPC 2.0.

### Key Files
- `cmd/github-mcp-server/main.go` - Entry point with multi-profile support
- `internal/server/server.go` - MCP handler with 77+ tool definitions
- `pkg/git/operations.go` - Local Git operations (0 tokens)
- `pkg/github/client.go` - GitHub API client wrapper with 11 methods
- `pkg/admin/admin.go` - Administrative operations client **[NEW v3.0]**
- `internal/hybrid/operations.go` - Smart Git-first operations
- `pkg/dashboard/dashboard.go` - GitHub dashboard operations
- `pkg/safety/` - Safety layer with risk classification and audit **[NEW v3.0]**
- `internal/server/admin_tools.go` - 22 administrative tool definitions **[NEW v3.0]**
- `internal/server/admin_handlers.go` - Admin operation handlers **[NEW v3.0]**
- `internal/server/safety_middleware.go` - Safety execution wrapper **[NEW v3.0]**

### Features
- **Multi-profile**: Single executable, multiple GitHub accounts via `--profile` flag
- **Protocol auto-detection**: Automatically detects and responds with client's MCP protocol version (universal compatibility)
- **Hybrid ops**: Local Git first (0 tokens) → GitHub API fallback
- **45+ Git tools**: Complete workflow from basic to advanced operations
- **Conflict management**: Safe merge, detection, resolution, backups
- **Response tools**: Comment on issues/PRs, create PR reviews
- **Repair tools**: Close issues, merge PRs, re-run workflows, dismiss security alerts
- **Administrative controls**: Full repository management, collaborators, webhooks, teams **[NEW v3.0]**
- **4-tier safety system**: Risk-based operation classification with confirmation tokens **[NEW v3.0]**
- **Audit logging**: JSON-based operation tracking with automatic rotation **[NEW v3.0]**
- **Dry-run mode**: Preview destructive operations before execution **[NEW v3.0]**
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

## Available Tools (77+)

- **Info** (8): `git_status`, `git_list_files`, `git_get_file_content`, etc.
- **Basic Git** (6): `git_add`, `git_commit`, `git_push`, `git_pull`, `git_checkout`
- **Advanced Git** (7): `git_merge`, `git_rebase`, `git_force_push`, `git_sync_with_remote`
- **Conflicts** (6): `git_safe_merge`, `git_resolve_conflicts`, `git_detect_conflicts`
- **Hybrid** (2): `create_file`, `update_file` (Git-first, API fallback)
- **GitHub API** (4): `github_list_repos`, `github_create_repo`, `github_list_prs`, `github_create_pr`
- **Dashboard** (7): `github_dashboard`, `github_notifications`, `github_assigned_issues`, `github_prs_to_review`, `github_security_alerts`, `github_failed_workflows`, `github_mark_notification_read`
- **Response** (3): `github_comment_issue`, `github_comment_pr`, `github_review_pr`
- **Repair** (6): `github_close_issue`, `github_merge_pr`, `github_rerun_workflow`, `github_dismiss_dependabot_alert`, `github_dismiss_code_alert`, `github_dismiss_secret_alert`
- **Repository Admin** (4): `github_get_repo_settings`, `github_update_repo_settings`, `github_archive_repository`, `github_delete_repository` **[NEW v3.0]**
- **Branch Protection** (3): `github_get_branch_protection`, `github_update_branch_protection`, `github_delete_branch_protection` **[NEW v3.0]**
- **Webhooks** (5): `github_list_webhooks`, `github_create_webhook`, `github_update_webhook`, `github_delete_webhook`, `github_test_webhook` **[NEW v3.0]**
- **Collaborators** (8): `github_list_collaborators`, `github_check_collaborator`, `github_add_collaborator`, `github_update_collaborator_permission`, `github_remove_collaborator`, `github_list_invitations`, `github_accept_invitation`, `github_cancel_invitation` **[NEW v3.0]**
- **Teams** (2): `github_list_repo_teams`, `github_add_repo_team` **[NEW v3.0]**

## New in v3.0: Administrative Controls with Safety Layer

**Version 3.0** introduces comprehensive repository administration with a sophisticated 4-tier safety system.

### Administrative Capabilities (22 Tools)

#### Repository Settings
- **`github_get_repo_settings`**: Retrieve complete repository configuration
- **`github_update_repo_settings`**: Modify repository properties (name, description, visibility, features)
- **`github_archive_repository`**: Archive repository (read-only mode)
- **`github_delete_repository`**: Permanently delete repository (⚠️ CRITICAL)

#### Branch Protection
- **`github_get_branch_protection`**: View branch protection rules
- **`github_update_branch_protection`**: Configure branch protection (required reviews, status checks, restrictions)
- **`github_delete_branch_protection`**: Remove all branch protection

#### Webhook Management
- **`github_list_webhooks`**: List all repository webhooks
- **`github_create_webhook`**: Create new webhook with custom configuration
- **`github_update_webhook`**: Modify existing webhook settings
- **`github_delete_webhook`**: Remove webhook
- **`github_test_webhook`**: Send test delivery to webhook endpoint

#### Collaborator Management
- **`github_list_collaborators`**: View all repository collaborators
- **`github_check_collaborator`**: Verify collaborator access level
- **`github_add_collaborator`**: Invite user with specified permission (pull/triage/push/maintain/admin)
- **`github_update_collaborator_permission`**: Change collaborator permission level
- **`github_remove_collaborator`**: Revoke user access
- **`github_list_invitations`**: View pending invitations
- **`github_accept_invitation`**: Accept repository invitation
- **`github_cancel_invitation`**: Cancel pending invitation

#### Team Access (Organization Repos)
- **`github_list_repo_teams`**: View teams with repository access
- **`github_add_repo_team`**: Grant team access with permission level

### Safety System Architecture

#### 4-Tier Risk Classification
- **LOW (1)**: Read-only operations → Execute immediately
- **MEDIUM (2)**: Reversible changes → Requires dry-run confirmation in strict mode
- **HIGH (3)**: Impacts collaboration → Requires confirmation token
- **CRITICAL (4)**: Irreversible operations → Requires confirmation + backup recommendation

#### Safety Modes
- **Strict**: Maximum protection - confirms MEDIUM+ operations
- **Moderate**: Balanced - confirms HIGH+ operations (recommended for production)
- **Permissive**: Minimal - only confirms CRITICAL operations
- **Disabled**: No safety checks (⚠️ not recommended)

#### Security Features
- **Confirmation Tokens**: Single-use SHA256 tokens with 5-minute expiration
- **Parameter Validation**: Prevents path traversal, command injection, SSRF
- **Audit Logging**: JSON logs with timestamps, operation details, rollback commands
- **Dry-Run Mode**: Preview operations before execution (default for destructive actions)
- **Automatic Rotation**: Audit logs rotate at 10MB with 5 backup files

### Configuration

Create `safety.json` in the server directory (see `safety.json.example`):

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

If no configuration exists, defaults to **moderate mode** with audit logging enabled.

## New in v2.1: Response & Repair Capabilities

### Response Tools
- **`github_comment_issue`**: Add comments to issues
- **`github_comment_pr`**: Add comments to pull requests
- **`github_review_pr`**: Create PR reviews (APPROVE, REQUEST_CHANGES, or COMMENT)

### Repair Tools
- **`github_close_issue`**: Close issues with optional closing comment
- **`github_merge_pr`**: Merge pull requests (supports merge, squash, rebase methods)
- **`github_rerun_workflow`**: Re-run failed GitHub Actions workflows
- **`github_dismiss_dependabot_alert`**: Dismiss Dependabot security alerts
- **`github_dismiss_code_alert`**: Dismiss Code Scanning alerts
- **`github_dismiss_secret_alert`**: Dismiss Secret Scanning alerts

## Configuration

### Token Permissions

**Required for basic operations:**
- `repo` - Essential for all repository operations, webhooks, and collaborators

**Optional but recommended:**
- `security_events` - For security alert dismissal (v2.1 repair tools)
- `admin:org` - For organization team management (v3.0 team tools)
- `admin:repo_hook` - Enhanced webhook management (v3.0 webhook tools)

**⚠️ Administrative Features**: v3.0 admin tools require `repo` scope with admin access to target repositories. Ensure your token has appropriate permissions for destructive operations.

### Multi-profile Setup (Recommended)

```json
{
  "mcpServers": {
    "github-personal": {
      "command": "path\\to\\github-mcp-server-v3.exe",
      "args": ["--profile", "personal"],
      "env": {"GITHUB_TOKEN": "ghp_token_personal"}
    },
    "github-work": {
      "command": "path\\to\\github-mcp-server-v3.exe",
      "args": ["--profile", "work"],
      "env": {"GITHUB_TOKEN": "ghp_token_work"}
    }
  }
}
```

### Safety Configuration

Place `safety.json` in the same directory as the executable, or specify with `--safety-config`:

```bash
github-mcp-server-v3.exe --safety-config=/path/to/custom-safety.json
```