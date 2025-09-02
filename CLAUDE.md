# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Repository: **github.com/scopweb/mcp-go-github**

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
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests for specific module
go test ./internal/hybrid/ -v
```

Comprehensive test coverage includes:
- Unit tests for all core components
- Hybrid operations testing
- Git environment detection
- GitHub API integration tests

### Dependencies
- Go 1.23.0+
- `github.com/google/go-github/v74` v74.0.0 - GitHub API client (latest stable)
- `golang.org/x/oauth2` - OAuth2 authentication
- `github.com/stretchr/testify` - Testing framework

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

The server provides 15+ fully tested tools organized into categories:
- **Repository operations**: list_repos, create_repo, get_repo ✅
- **Branch management**: list_branches ✅
- **Pull requests**: list_prs, create_pr ✅ 
- **Issues**: list_issues, create_issue ✅
- **Local Git**: git_status, git_list_files ✅
- **File operations**: create_file, update_file (hybrid mode) ✅

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

### Security Enhancements
- **Argument injection prevention**: Neutralizes potential command injection attacks
- **Path traversal protection**: Validates file paths to prevent unauthorized access
- **Input validation**: Strict validation of all user inputs before processing

### Project Status
- ✅ **Production ready**: Stable release v2.1
- ✅ **Full test coverage**: All functions tested with unit tests
- ✅ **Latest dependencies**: go-github v74.0.0
- ✅ **Security hardened**: Enhanced with injection prevention