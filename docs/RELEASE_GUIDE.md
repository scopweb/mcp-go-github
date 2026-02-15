# Release Preparation Guide - v1.0.0

## Overview

This guide provides step-by-step instructions for creating the first official release (v1.0.0) of mcp-go-github.

## Pre-Release Checklist

### 1. Code Quality ✅
- [x] All tests passing (`go test ./...`)
- [x] No Spanish content in public files
- [x] Documentation organized
- [x] README in English
- [x] CHANGELOG up to date

### 2. Repository Cleanup ✅
- [x] Internal docs moved to docs/internal/
- [x] Third-party licenses in licenses/
- [x] No local paths exposed
- [x] No sensitive data visible

### 3. Pre-Release Tasks ⏳
- [ ] Add GitHub topics
- [ ] Update version numbers
- [ ] Test on all platforms
- [ ] Create release notes
- [ ] Build binaries

## Version Number Strategy

**Current**: v3.0.0 (code version)
**First Release**: v1.0.0 or v3.0.0?

### Option A: Start at v1.0.0 (Recommended)
```
Reasoning:
- First public release
- Signals "production ready"
- Follows semantic versioning fresh start
- Clear version history

Release as: v1.0.0
```

### Option B: Continue at v3.0.0
```
Reasoning:
- Matches internal version
- Reflects actual development history
- Acknowledges existing features

Release as: v3.0.0
```

**Recommendation**: Use **v3.0.0** since the code already has this version in filenames, documentation, and commit history.

## Building Release Binaries

### 1. Update Version Information

**cmd/github-mcp-server/main.go:**
```go
const Version = "3.0.0"
const ReleaseDate = "2026-02-15"
```

### 2. Build for All Platforms

**Windows (AMD64):**
```bash
.\compile.bat
# Output: github-mcp-server-v3.exe
```

**macOS (ARM64 - M1/M2/M3/M4):**
```bash
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/mac-arm64/github-mcp-server-v3 ./cmd/github-mcp-server/
```

**macOS (AMD64 - Intel):**
```bash
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/mac-amd64/github-mcp-server-v3 ./cmd/github-mcp-server/
```

**Linux (AMD64):**
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/linux-amd64/github-mcp-server-v3 ./cmd/github-mcp-server/
```

**Linux (ARM64):**
```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/linux-arm64/github-mcp-server-v3 ./cmd/github-mcp-server/
```

### 3. Automated Build Script

Create `build-release.sh`:

```bash
#!/bin/bash

VERSION="3.0.0"
BINARY_NAME="github-mcp-server-v3"

echo "Building release v${VERSION} for all platforms..."

# Create dist directories
mkdir -p dist/{windows-amd64,mac-arm64,mac-amd64,linux-amd64,linux-arm64}

# Windows AMD64
echo "[1/5] Building for Windows AMD64..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags="-s -w -X main.Version=${VERSION}" \
  -o "dist/windows-amd64/${BINARY_NAME}.exe" \
  ./cmd/github-mcp-server/

# macOS ARM64
echo "[2/5] Building for macOS ARM64..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 \
  go build -ldflags="-s -w -X main.Version=${VERSION}" \
  -o "dist/mac-arm64/${BINARY_NAME}" \
  ./cmd/github-mcp-server/

# macOS AMD64
echo "[3/5] Building for macOS AMD64..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags="-s -w -X main.Version=${VERSION}" \
  -o "dist/mac-amd64/${BINARY_NAME}" \
  ./cmd/github-mcp-server/

# Linux AMD64
echo "[4/5] Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags="-s -w -X main.Version=${VERSION}" \
  -o "dist/linux-amd64/${BINARY_NAME}" \
  ./cmd/github-mcp-server/

# Linux ARM64
echo "[5/5] Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
  go build -ldflags="-s -w -X main.Version=${VERSION}" \
  -o "dist/linux-arm64/${BINARY_NAME}" \
  ./cmd/github-mcp-server/

echo "✅ Build complete!"
```

### 4. Create Release Packages

```bash
# Copy additional files to each platform directory
for dir in dist/*/; do
  cp safety.json.example "$dir"
  cp README.md "$dir"
  cp LICENSE "$dir"
  cp CHANGELOG.md "$dir"
done

# Copy platform-specific installers
cp install-mac.sh dist/mac-arm64/
cp install-mac.sh dist/mac-amd64/
```

### 5. Generate Checksums

```bash
cd dist

# For each platform
for platform in windows-amd64 mac-arm64 mac-amd64 linux-amd64 linux-arm64; do
  cd "$platform"

  # Create zip/tar.gz
  if [[ "$platform" == windows-* ]]; then
    zip -r "../mcp-go-github-v3.0.0-${platform}.zip" .
    cd ..
    sha256sum "mcp-go-github-v3.0.0-${platform}.zip" >> SHA256SUMS.txt
  else
    tar -czf "../mcp-go-github-v3.0.0-${platform}.tar.gz" .
    cd ..
    sha256sum "mcp-go-github-v3.0.0-${platform}.tar.gz" >> SHA256SUMS.txt
  fi
done

cd ..
```

## Release Notes Template

Create `RELEASE_NOTES.md`:

```markdown
# GitHub MCP Server v3.0.0

## 🎉 First Official Release

We're excited to announce the first official release of the GitHub MCP Server - a powerful Go-based MCP server that connects GitHub to Claude Desktop.

## ✨ Features

### Core Capabilities
- **82 MCP Tools**: Comprehensive GitHub and Git operations
  - 48 tools work without Git installed
  - 34 Git-based tools for local repository operations
- **Multi-Profile Support**: Manage multiple GitHub accounts simultaneously
- **Hybrid Architecture**: Local Git operations (0 tokens) with GitHub API fallback

### Administrative Controls (NEW in v3.0)
- **22 Administrative Tools**: Full repository management
  - Repository settings (name, description, visibility)
  - Branch protection rules
  - Webhook management (create, update, delete, test)
  - Collaborator management (invite, remove, permissions)
  - Team access control
- **4-Tier Safety System**: Risk-based operation classification
  - LOW: Read-only operations
  - MEDIUM: Reversible changes
  - HIGH: Impacts collaboration
  - CRITICAL: Irreversible operations
- **Audit Logging**: JSON-based operation tracking with automatic rotation
- **Confirmation Tokens**: Cryptographic tokens for destructive operations

### Security & Safety
- Path traversal prevention
- Command injection protection
- SSRF prevention in webhook URLs
- Strict input validation
- Dry-run mode for destructive operations

## 📦 Installation

### Download

Choose your platform:

- **Windows (AMD64)**: [mcp-go-github-v3.0.0-windows-amd64.zip](...)
- **macOS (Apple Silicon)**: [mcp-go-github-v3.0.0-mac-arm64.tar.gz](...)
- **macOS (Intel)**: [mcp-go-github-v3.0.0-mac-amd64.tar.gz](...)
- **Linux (AMD64)**: [mcp-go-github-v3.0.0-linux-amd64.tar.gz](...)
- **Linux (ARM64)**: [mcp-go-github-v3.0.0-linux-arm64.tar.gz](...)

Verify downloads: [SHA256SUMS.txt](...)

### Quick Start

1. **Extract the archive**
2. **Generate GitHub token**: https://github.com/settings/tokens
   - Required scope: `repo`
   - Optional: `workflow`, `security_events`, `admin:org`, `admin:repo_hook`
3. **Configure Claude Desktop**:

```json
{
  "mcpServers": {
    "github": {
      "command": "/path/to/github-mcp-server-v3",
      "env": {
        "GITHUB_TOKEN": "your_token_here"
      }
    }
  }
}
```

4. **Restart Claude Desktop**

Full installation guide: [README.md](README.md)

## 🔧 Configuration

### Multi-Profile Setup

Manage multiple GitHub accounts:

```json
{
  "mcpServers": {
    "github-personal": {
      "command": "/path/to/github-mcp-server-v3",
      "args": ["--profile", "personal"],
      "env": {"GITHUB_TOKEN": "ghp_personal_token"}
    },
    "github-work": {
      "command": "/path/to/github-mcp-server-v3",
      "args": ["--profile", "work"],
      "env": {"GITHUB_TOKEN": "ghp_work_token"}
    }
  }
}
```

### Safety Configuration

Create `safety.json` (optional):

```json
{
  "mode": "moderate",
  "enable_audit_log": true,
  "require_confirmation_above": 3,
  "audit_log_path": "./mcp-admin-audit.log"
}
```

See `safety.json.example` for full configuration options.

## 📊 What's Included

### Git Operations (34 tools)
- Information: status, list files, get content, SHA, commits
- Basic: add, commit, push, pull, checkout
- Advanced: merge, rebase, force-push, sync
- Conflicts: safe merge, conflict detection, resolution, backups
- Analysis: log analysis, diff, branch list, stash, remotes, tags

### GitHub API (48 tools)
- Repositories: list, create
- Pull Requests: list, create, comment, review, merge
- Issues: list, comment, close
- Dashboard: notifications, assigned issues, PRs to review
- Security: security alerts, failed workflows, alert dismissal
- File Operations: list contents, download file/repo, pull updates
- **Admin** (NEW): repo settings, branch protection, webhooks, collaborators, teams

### Hybrid Operations
- `create_file`: Git-first with API fallback
- `update_file`: Git-first with API fallback

## 🔐 Security

This release includes comprehensive security measures:

- Argument injection prevention
- Path traversal defense
- SSRF protection
- Input validation
- Cryptographic confirmation tokens
- Audit logging

See [SECURITY.md](SECURITY.md) for our security policy.

## 📝 System Requirements

- **Go**: 1.25.0+ (for building from source)
- **Git**: Optional (auto-detected, 48 tools work without)
- **GitHub Token**: Personal access token with `repo` scope
- **Claude Desktop**: Latest version

## 🐛 Known Issues

None at release time.

## 📚 Documentation

- [README.md](README.md) - Full documentation
- [CHANGELOG.md](CHANGELOG.md) - Change history
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [SECURITY.md](SECURITY.md) - Security policy

## 🙏 Acknowledgments

Built with:
- [google/go-github](https://github.com/google/go-github) v81.0.0
- [golang.org/x/oauth2](https://pkg.go.dev/golang.org/x/oauth2) v0.34.0

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🔗 Links

- **Repository**: https://github.com/YOUR_USERNAME/mcp-go-github
- **Issues**: https://github.com/YOUR_USERNAME/mcp-go-github/issues
- **Discussions**: https://github.com/YOUR_USERNAME/mcp-go-github/discussions

---

**Full Changelog**: https://github.com/YOUR_USERNAME/mcp-go-github/commits/v3.0.0
```

## GitHub Release Steps

### 1. Create Release on GitHub

```bash
# Tag the release
git tag -a v3.0.0 -m "Release v3.0.0 - First official release"
git push origin v3.0.0

# Or use GitHub CLI
gh release create v3.0.0 \
  --title "GitHub MCP Server v3.0.0" \
  --notes-file RELEASE_NOTES.md \
  dist/mcp-go-github-v3.0.0-windows-amd64.zip \
  dist/mcp-go-github-v3.0.0-mac-arm64.tar.gz \
  dist/mcp-go-github-v3.0.0-mac-amd64.tar.gz \
  dist/mcp-go-github-v3.0.0-linux-amd64.tar.gz \
  dist/mcp-go-github-v3.0.0-linux-arm64.tar.gz \
  dist/SHA256SUMS.txt
```

### 2. Verify Release

- [ ] All binaries uploaded
- [ ] SHA256SUMS.txt present
- [ ] Release notes formatted correctly
- [ ] Links work
- [ ] Downloads work

### 3. Announce Release

**Platforms:**
- GitHub Discussions
- Reddit (r/golang, r/ClaudeAI)
- Twitter/X
- Hacker News (if significant interest)
- Model Context Protocol Discord/Forum

**Announcement Template:**
```
🎉 GitHub MCP Server v3.0.0 Released!

First official release of mcp-go-github - A powerful Go-based MCP server connecting GitHub to Claude Desktop.

✨ Features:
- 82 MCP tools (48 work without Git)
- 22 administrative tools with 4-tier safety system
- Multi-profile support
- Comprehensive security

📦 Download: https://github.com/YOUR_USERNAME/mcp-go-github/releases/tag/v3.0.0

#MCP #GitHub #ClaudeDesktop #Golang
```

## Post-Release

### 1. Update Documentation

- [ ] Update README badges (if added)
- [ ] Update installation instructions with release links
- [ ] Create "Upgrade Guide" for future releases

### 2. Monitor

- [ ] Watch for issues
- [ ] Respond to questions
- [ ] Track downloads
- [ ] Collect feedback

### 3. Plan Next Release

- [ ] Create milestone for v3.1.0
- [ ] Review open issues
- [ ] Plan features
- [ ] Update roadmap

## Troubleshooting

### Build Failures

```bash
# Clean and rebuild
go clean -cache -modcache -i -r
go mod download
go build ./cmd/github-mcp-server/
```

### Missing Dependencies

```bash
go mod tidy
go mod verify
```

### Cross-Compilation Issues

```bash
# Ensure CGO is disabled for cross-compilation
export CGO_ENABLED=0
```

## Rollback Plan

If critical issues are discovered:

1. Mark release as "pre-release" on GitHub
2. Create hotfix branch
3. Fix issue
4. Release v3.0.1
5. Update v3.0.0 release notes with "Superseded by v3.0.1"

## Checklist Summary

### Pre-Release
- [ ] All tests passing
- [ ] Documentation updated
- [ ] Version numbers correct
- [ ] CHANGELOG updated
- [ ] GitHub topics added

### Build
- [ ] Windows binary built
- [ ] macOS ARM64 binary built
- [ ] macOS AMD64 binary built
- [ ] Linux AMD64 binary built
- [ ] Linux ARM64 binary built
- [ ] SHA256SUMS generated

### Release
- [ ] Git tag created
- [ ] GitHub release created
- [ ] Binaries uploaded
- [ ] Release notes published
- [ ] Release verified

### Post-Release
- [ ] Announcement posted
- [ ] Documentation updated
- [ ] Next milestone created
- [ ] Feedback collected
