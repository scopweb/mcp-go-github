# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Architecture

This is a **Go-based MCP (Model Context Protocol) server** that provides GitHub integration for Claude Desktop. The architecture follows a modular design with hybrid operation support (local Git + GitHub API).

### Core Components

- **main.go**: Entry point with profile-based configuration system
- **internal/server/**: JSON-RPC 2.0 protocol handler for MCP communication
- **internal/git/**: Local Git operations and environment detection
- **internal/github/**: GitHub API client wrapper
- **internal/hybrid/**: Smart operations that prefer local Git over API calls
- **internal/types/**: Shared type definitions and protocol structures

### Key Architecture Features

- **Multi-profile support**: Single executable handles multiple GitHub accounts via `--profile` flag
- **Hybrid operations**: Prioritizes local Git commands to minimize API token usage
- **MCP protocol compliance**: Implements JSON-RPC 2.0 for Claude Desktop integration
- **Smart Git detection**: Automatically detects Git environment and repository state
- **Advanced Git operations**: 13 new operations including remote checkout, merge strategies, conflict resolution
- **Safety features**: Automatic backups, clean state validation, conflict detection

## Development Commands

### Build and Compilation
```bash
# Clean dependencies and build
go mod tidy
go build -o github-mcp-modular.exe main.go

# Use the provided build script (Windows)
.\compile.bat
```

### Testing
- No automated test suite is currently implemented
- Manual testing through Claude Desktop integration required

### Dependencies
- Go 1.23.0+
- `github.com/google/go-github/v66` - GitHub API client
- `golang.org/x/oauth2` - OAuth2 authentication

## Configuration Requirements

### GitHub Token Permissions
Minimum required scopes for GitHub Personal Access Token:
- `repo` (Full control of private repositories) - Essential for all operations
- `delete_repo` (Delete repositories) - Optional, only if deletion needed
- `workflow` (Update GitHub Action workflows) - Optional, for Actions integration

### Claude Desktop Setup
The server supports both single-token and multi-profile configurations:

**Multi-profile (recommended):**
```json
{
  "mcpServers": {
    "github-local-personal": {
      "command": "path\\to\\github-mcp-modular.exe",
      "args": ["--profile", "personal"],
      "env": {
        "GITHUB_TOKEN": "ghp_token_personal"
      }
    },
    "github-local-trabajo": {
      "command": "path\\to\\github-mcp-modular.exe", 
      "args": ["--profile", "trabajo"],
      "env": {
        "GITHUB_TOKEN": "ghp_token_trabajo"
      }
    }
  }
}
```

## Available MCP Tools

The server provides 25+ tools organized into categories:
- **Repository operations**: list_repos, create_repo, get_repo
- **Branch management**: list_branches, checkout_remote, merge, rebase
- **Pull requests**: list_prs, create_pr
- **Issues**: list_issues, create_issue
- **Local Git**: git_status, git_list_files, pull_with_strategy, force_push, push_upstream
- **File operations**: create_file, update_file (hybrid mode)
- **Advanced Git**: sync_with_remote, safe_merge, conflict_status, resolve_conflicts
- **Safety tools**: validate_clean_state, detect_conflicts, create_backup

## Development Notes

### Hybrid Operation Logic
The system implements a "smart" approach:
1. **Local Git first**: Attempts local Git operations to save API calls
2. **GitHub API fallback**: Uses API only when local Git fails or is unavailable
3. **Token optimization**: Reduces API token consumption through local-first strategy

### Profile System
- Each profile instance runs independently with its own GitHub token
- Logging includes profile identification for multi-instance debugging
- Single executable supports unlimited profiles

### Error Handling
- Comprehensive JSON-RPC 2.0 error responses
- Token permission validation with helpful error messages
- Git environment detection with graceful fallbacks