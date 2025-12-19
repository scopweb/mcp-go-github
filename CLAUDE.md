# CLAUDE.md - MCP GitHub Server v2.1

**Status**: ✅ Production Ready | **Tools**: 55+ | **Architecture**: Hybrid (Local Git + GitHub API)

## Core Architecture

**Go-based MCP server** providing GitHub integration for Claude Desktop via JSON-RPC 2.0.

### Key Files
- `cmd/github-mcp-server/main.go` - Entry point with multi-profile support
- `internal/server/server.go` - MCP handler with 55+ tool definitions
- `pkg/git/operations.go` - Local Git operations (0 tokens)
- `pkg/github/client.go` - GitHub API client wrapper with 11 new methods
- `internal/hybrid/operations.go` - Smart Git-first operations
- `pkg/dashboard/dashboard.go` - GitHub dashboard operations

### Features
- **Multi-profile**: Single executable, multiple GitHub accounts via `--profile` flag
- **Hybrid ops**: Local Git first (0 tokens) → GitHub API fallback
- **45+ Git tools**: Complete workflow from basic to advanced operations
- **Conflict management**: Safe merge, detection, resolution, backups
- **Response tools**: Comment on issues/PRs, create PR reviews
- **Repair tools**: Close issues, merge PRs, re-run workflows, dismiss security alerts
- **Security**: Path traversal prevention, command injection protection

## Build & Test

```bash
.\compile.bat                    # Build
go test ./...                    # Run all tests
go test ./pkg/git/ -v           # Test specific package
```

## Dependencies
- Go 1.24.0+
- `github.com/google/go-github/v77 v77.0.0`
- `golang.org/x/oauth2 v0.33.0`

## Available Tools (55+)

- **Info** (8): `git_status`, `git_list_files`, `git_get_file_content`, etc.
- **Basic Git** (6): `git_add`, `git_commit`, `git_push`, `git_pull`, `git_checkout`
- **Advanced Git** (7): `git_merge`, `git_rebase`, `git_force_push`, `git_sync_with_remote`
- **Conflicts** (6): `git_safe_merge`, `git_resolve_conflicts`, `git_detect_conflicts`
- **Hybrid** (2): `create_file`, `update_file` (Git-first, API fallback)
- **GitHub API** (4): `github_list_repos`, `github_create_repo`, `github_list_prs`, `github_create_pr`
- **Dashboard** (7): `github_dashboard`, `github_notifications`, `github_assigned_issues`, `github_prs_to_review`, `github_security_alerts`, `github_failed_workflows`, `github_mark_notification_read`
- **Response** (3): `github_comment_issue`, `github_comment_pr`, `github_review_pr` **[NEW v2.1]**
- **Repair** (6): `github_close_issue`, `github_merge_pr`, `github_rerun_workflow`, `github_dismiss_dependabot_alert`, `github_dismiss_code_alert`, `github_dismiss_secret_alert` **[NEW v2.1]**

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

**Token Permissions Required**: `repo` (essential for all operations), `security_events` (optional but recommended for alert dismissal)

**Multi-profile setup** (recommended):
```json
{
  "mcpServers": {
    "github-personal": {
      "command": "path\\to\\github-mcp-modular.exe",
      "args": ["--profile", "personal"],
      "env": {"GITHUB_TOKEN": "ghp_token_personal"}
    }
  }
}
```