# CLAUDE.md - MCP GitHub Server v2.0

**Status**: ✅ Production Ready | **Tools**: 45+ | **Architecture**: Hybrid (Local Git + GitHub API)

## Core Architecture

**Go-based MCP server** providing GitHub integration for Claude Desktop via JSON-RPC 2.0.

### Key Files
- `cmd/github-mcp-server/main.go` - Entry point with multi-profile support
- `internal/server/server.go` - MCP handler with 45+ tool definitions
- `pkg/git/operations.go` - Local Git operations (0 tokens)
- `pkg/github/client.go` - GitHub API client wrapper
- `internal/hybrid/operations.go` - Smart Git-first operations

### Features
- **Multi-profile**: Single executable, multiple GitHub accounts via `--profile` flag
- **Hybrid ops**: Local Git first (0 tokens) → GitHub API fallback
- **45+ Git tools**: Complete workflow from basic to advanced operations
- **Conflict management**: Safe merge, detection, resolution, backups
- **Security**: Path traversal prevention, command injection protection

## Build & Test

```bash
.\compile.bat                    # Build
go test ./...                    # Run all tests
go test ./pkg/git/ -v           # Test specific package
```

## Dependencies
- Go 1.24.0+
- `github.com/google/go-github/v76 v76.0.0`
- `golang.org/x/oauth2 v0.32.0`

## Available Tools (45+)

- **Info** (8): `git_status`, `git_list_files`, `git_get_file_content`, etc.
- **Basic Git** (6): `git_add`, `git_commit`, `git_push`, `git_pull`, `git_checkout`
- **Advanced Git** (7): `git_merge`, `git_rebase`, `git_force_push`, `git_sync_with_remote`
- **Conflicts** (6): `git_safe_merge`, `git_resolve_conflicts`, `git_detect_conflicts`
- **Hybrid** (2): `create_file`, `update_file` (Git-first, API fallback)
- **GitHub API** (4): `github_list_repos`, `github_create_repo`, `github_list_prs`, `github_create_pr`

## Configuration

**Token Permissions Required**: `repo` (essential), `delete_repo` (optional), `workflow` (optional)

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