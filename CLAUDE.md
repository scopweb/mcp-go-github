# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 🚀 Quick Project Status (v2.0)

**Status**: ✅ Production Ready | **Tools**: 45+ | **Git Operations**: 0 Tokens | **Multi-Profile**: ✅ Implemented

### Latest Updates
- ✅ **45+ Git Tools**: Comprehensive local Git operations (zero API tokens)
- ✅ **Advanced Conflict Management**: Safe merge, detection, and resolution strategies
- ✅ **Multi-Profile Support**: Single executable for multiple GitHub accounts
- ✅ **Dependencies Updated**: go-github v76.0.0, oauth2 v0.32.0
- ✅ **Hybrid System**: Local Git prioritized over GitHub API calls
- ✅ **Security Hardened**: Path traversal prevention, argument injection protection

## Project Architecture

This is a **Go-based MCP (Model Context Protocol) server** that provides GitHub integration for Claude Desktop. The architecture follows a modular design with hybrid operation support (local Git + GitHub API).

### Core Components

- **cmd/github-mcp-server/main.go**: Entry point with profile-based configuration system
- **internal/server/server.go**: JSON-RPC 2.0 protocol handler with 45+ tool definitions
- **pkg/git/operations.go**: Local Git operations (status, add, commit, push, pull, checkout, merge, rebase, etc.)
- **pkg/github/client.go**: GitHub API client wrapper
- **internal/hybrid/operations.go**: Smart operations that prefer local Git over API calls
- **pkg/types/types.go**: Shared type definitions and protocol structures
- **pkg/interfaces/interfaces.go**: Interface definitions for Git and GitHub operations

### Key Architecture Features

- **Multi-profile support**: Single executable handles multiple GitHub accounts via `--profile` flag
- **Hybrid operations**: Prioritizes local Git commands to minimize API token usage (0 tokens for local ops)
- **MCP protocol compliance**: Implements JSON-RPC 2.0 for Claude Desktop integration
- **Smart Git detection**: Automatically detects Git environment and repository state
- **45+ Git operations**: Complete Git workflow from basic to advanced operations
- **Conflict management suite**: Safe merge, detection, resolution, and backup creation
- **Security hardened**: Path traversal prevention, command injection protection, input validation

## Development Commands

### Build and Compilation
```bash
# Clean dependencies and build
go mod tidy
go mod vendor
go build -o mcp-go-github-modular.exe ./cmd/github-mcp-server/main.go

# Use the provided build script (Windows)
.\compile.bat
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run specific package tests
go test ./pkg/git/ -v
go test ./internal/hybrid/ -v
```

### Dependencies
- Go 1.24.0+ (toolchain 1.24.6)
- `github.com/google/go-github/v76 v76.0.0` - GitHub API client
- `golang.org/x/oauth2 v0.32.0` - OAuth2 authentication
- `github.com/stretchr/testify v1.11.1` - Testing framework (dev only)

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

## Available MCP Tools (45+ Tools)

The server provides comprehensive Git and GitHub operations organized into 8 categories:

### 🔍 Information Tools (8 tools) - 0 tokens
- `git_status`, `git_list_files`, `git_get_file_content`, `git_get_file_sha`
- `git_get_last_commit`, `git_get_changed_files`, `git_validate_repo`, `git_context`

### ⚙️ Basic Git Operations (6 tools) - 0 tokens
- `git_set_workspace`, `git_add`, `git_commit`, `git_push`, `git_pull`, `git_checkout`

### 📊 Analysis & Management (7 tools) - 0 tokens
- `git_log_analysis`, `git_diff_files`, `git_branch_list`, `git_stash`
- `git_remote`, `git_tag`, `git_clean`

### 🚀 Advanced Git Operations (7 tools) - 0 tokens
- `git_checkout_remote`, `git_merge`, `git_rebase`, `git_pull_with_strategy`
- `git_force_push`, `git_push_upstream`, `git_sync_with_remote`

### 🛡️ Conflict Management (6 tools) - 0 tokens
- `git_safe_merge`, `git_conflict_status`, `git_resolve_conflicts`
- `git_validate_clean_state`, `git_detect_conflicts`, `git_create_backup`

### 🔀 Hybrid Operations (2 tools) - 0 tokens with local Git
- `create_file`, `update_file` (prioritize local Git, fallback to GitHub API)

### 🌐 GitHub API Operations (4 tools) - requires API tokens
- `github_list_repos`, `github_create_repo`, `github_list_prs`, `github_create_pr`

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
- Path validation to prevent directory traversal attacks
- Command injection protection through argument sanitization

## File Structure Summary

```
mcp-go-github/
├── cmd/
│   └── github-mcp-server/
│       └── main.go              # Entry point with profile system
├── internal/
│   ├── hybrid/
│   │   └── operations.go        # Hybrid Git/API operations
│   └── server/
│       └── server.go            # MCP JSON-RPC handler (45+ tools)
├── pkg/
│   ├── git/
│   │   ├── operations.go        # Local Git operations
│   │   └── operations_test.go   # Git operation tests
│   ├── github/
│   │   ├── client.go            # GitHub API client
│   │   └── client_test.go       # API client tests
│   ├── interfaces/
│   │   └── interfaces.go        # Operation interfaces
│   └── types/
│       └── types.go             # Protocol type definitions
├── vendor/                      # Vendored dependencies
├── compile.bat                  # Windows build script
├── go.mod                       # Go module definition
└── README.md                    # User documentation
```