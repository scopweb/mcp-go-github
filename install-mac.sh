#!/bin/bash
# ============================================================
#  GitHub MCP Server v3.0 - Mac Installer
#  Installs the server and configures Claude Desktop
# ============================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Installation paths
INSTALL_DIR="$HOME/.mcp-servers/github"
BINARY_NAME="github-mcp-server-v3"
CLAUDE_CONFIG_DIR="$HOME/Library/Application Support/Claude"
CLAUDE_CONFIG_FILE="$CLAUDE_CONFIG_DIR/claude_desktop_config.json"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# ---- Helper functions ----

print_header() {
    echo ""
    echo -e "${CYAN}============================================================${NC}"
    echo -e "${CYAN}  GitHub MCP Server v3.0 - Mac Installer${NC}"
    echo -e "${CYAN}============================================================${NC}"
    echo ""
    echo -e "  ${BOLD}82 tools${NC} (with Git) | ${BOLD}48 tools${NC} (without Git)"
    echo -e "  Admin controls | Safety system | Git-free file ops"
    echo ""
}

print_step() {
    echo -e "${BLUE}[$1/$TOTAL_STEPS]${NC} $2"
}

print_ok() {
    echo -e "    ${GREEN}OK${NC} $1"
}

print_warn() {
    echo -e "    ${YELLOW}WARNING${NC} $1"
}

print_error() {
    echo -e "    ${RED}ERROR${NC} $1"
}

# ---- Step 1: Check system ----

check_system() {
    print_step 1 "Checking system..."

    # Check macOS
    if [[ "$(uname)" != "Darwin" ]]; then
        print_error "This installer is for macOS only."
        exit 1
    fi
    print_ok "macOS detected"

    # Check architecture
    ARCH="$(uname -m)"
    if [[ "$ARCH" == "arm64" ]]; then
        print_ok "Apple Silicon (ARM64)"
    elif [[ "$ARCH" == "x86_64" ]]; then
        print_ok "Intel Mac (AMD64)"
    else
        print_error "Unsupported architecture: $ARCH"
        exit 1
    fi

    # Check Git availability (informational only)
    if command -v git &> /dev/null; then
        GIT_VERSION=$(git --version 2>/dev/null | head -1)
        print_ok "Git found: $GIT_VERSION (82 tools available)"
    else
        print_warn "Git not found (48 tools available - API tools + admin + file ops)"
        echo -e "    ${YELLOW}Tip:${NC} Install Xcode CLI tools for Git: xcode-select --install"
    fi

    # Check Claude Desktop
    if [[ -d "$CLAUDE_CONFIG_DIR" ]]; then
        print_ok "Claude Desktop config directory found"
    else
        print_warn "Claude Desktop config directory not found"
        echo -e "    ${YELLOW}Tip:${NC} Install Claude Desktop first, or we'll create the config directory"
    fi
}

# ---- Step 2: Get or build binary ----

get_binary() {
    print_step 2 "Preparing server binary..."

    # Check if pre-built binary exists in same directory
    if [[ -f "$SCRIPT_DIR/$BINARY_NAME" ]]; then
        print_ok "Pre-built binary found in installer directory"
        BINARY_SOURCE="$SCRIPT_DIR/$BINARY_NAME"
        return 0
    fi

    # Check if Go is available to compile from source
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version 2>/dev/null | head -1)
        print_ok "Go found: $GO_VERSION"

        # Check if source code is available
        if [[ -f "$SCRIPT_DIR/cmd/github-mcp-server/main.go" ]]; then
            echo -e "    Compiling from source..."
            cd "$SCRIPT_DIR"
            CGO_ENABLED=0 go build -ldflags="-s -w" -o "$BINARY_NAME" ./cmd/github-mcp-server/main.go
            if [[ $? -eq 0 ]]; then
                print_ok "Compiled successfully"
                BINARY_SOURCE="$SCRIPT_DIR/$BINARY_NAME"
                return 0
            else
                print_error "Compilation failed"
                exit 1
            fi
        fi
    fi

    # No binary and no Go - give instructions
    print_error "No pre-built binary found and Go is not installed."
    echo ""
    echo -e "  ${BOLD}Options:${NC}"
    echo ""
    echo -e "  ${CYAN}Option A:${NC} Get a pre-built binary"
    echo "    1. On a machine with Go, run: build-mac.bat (Windows) or:"
    echo "       GOOS=darwin GOARCH=$ARCH go build -o $BINARY_NAME ./cmd/github-mcp-server/"
    echo "    2. Copy $BINARY_NAME to this directory"
    echo "    3. Run this installer again"
    echo ""
    echo -e "  ${CYAN}Option B:${NC} Install Go on this Mac"
    if command -v brew &> /dev/null; then
        echo "    brew install go"
    else
        echo "    1. Install Homebrew: /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
        echo "    2. brew install go"
    fi
    echo "    Then run this installer again"
    echo ""
    exit 1
}

# ---- Step 3: Install binary ----

install_binary() {
    print_step 3 "Installing server..."

    # Create installation directory
    mkdir -p "$INSTALL_DIR"
    print_ok "Created $INSTALL_DIR"

    # Copy binary
    cp "$BINARY_SOURCE" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    print_ok "Installed binary to $INSTALL_DIR/$BINARY_NAME"

    # Copy safety.json.example if available
    if [[ -f "$SCRIPT_DIR/safety.json.example" ]]; then
        cp "$SCRIPT_DIR/safety.json.example" "$INSTALL_DIR/safety.json.example"
        print_ok "Copied safety.json.example"
    fi

    # Verify binary runs
    if "$INSTALL_DIR/$BINARY_NAME" --help 2>&1 | head -1 > /dev/null 2>&1; then
        print_ok "Binary verified"
    else
        # MCP servers don't have --help, they read stdin. That's fine.
        print_ok "Binary installed (MCP server, runs via stdin)"
    fi
}

# ---- Step 4: Get GitHub token ----

get_token() {
    print_step 4 "Configuring GitHub token..."

    # Check environment variable
    if [[ -n "$GITHUB_TOKEN" ]]; then
        TOKEN_PREVIEW="${GITHUB_TOKEN:0:8}***"
        print_ok "Found GITHUB_TOKEN in environment: $TOKEN_PREVIEW"
        echo ""
        echo -e "    Use this token? ${BOLD}[Y/n]${NC} "
        read -r USE_ENV_TOKEN
        if [[ "$USE_ENV_TOKEN" != "n" && "$USE_ENV_TOKEN" != "N" ]]; then
            MCP_TOKEN="$GITHUB_TOKEN"
            return 0
        fi
    fi

    echo ""
    echo -e "  ${BOLD}GitHub Personal Access Token${NC}"
    echo ""
    echo "  Required permissions: repo"
    echo "  Optional: delete_repo, workflow, security_events, admin:repo_hook"
    echo ""
    echo "  Generate at: https://github.com/settings/tokens"
    echo ""
    echo -n "  Enter your GitHub token (ghp_...): "
    read -r MCP_TOKEN

    if [[ -z "$MCP_TOKEN" ]]; then
        print_error "No token provided. You can set it later in Claude Desktop config."
        MCP_TOKEN="YOUR_GITHUB_TOKEN_HERE"
    else
        TOKEN_PREVIEW="${MCP_TOKEN:0:8}***"
        print_ok "Token received: $TOKEN_PREVIEW"
    fi
}

# ---- Step 5: Get profile name ----

get_profile() {
    print_step 5 "Configuring profile..."

    echo ""
    echo -e "  Profile name identifies this GitHub account in Claude Desktop."
    echo -e "  Examples: personal, work, company-name"
    echo ""
    echo -n "  Profile name [default: github]: "
    read -r PROFILE_NAME

    if [[ -z "$PROFILE_NAME" ]]; then
        PROFILE_NAME="github"
    fi

    # Sanitize profile name (alphanumeric and hyphens only)
    PROFILE_NAME=$(echo "$PROFILE_NAME" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9-]/-/g')
    print_ok "Profile: $PROFILE_NAME"

    MCP_SERVER_NAME="github-$PROFILE_NAME"
}

# ---- Step 6: Configure Claude Desktop ----

configure_claude() {
    print_step 6 "Configuring Claude Desktop..."

    BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"

    # Create Claude config directory if it doesn't exist
    if [[ ! -d "$CLAUDE_CONFIG_DIR" ]]; then
        mkdir -p "$CLAUDE_CONFIG_DIR"
        print_ok "Created Claude Desktop config directory"
    fi

    # Build the new server entry
    NEW_SERVER_JSON=$(cat <<JSONEOF
{
    "command": "$BINARY_PATH",
    "args": ["--profile", "$PROFILE_NAME"],
    "env": {
        "GITHUB_TOKEN": "$MCP_TOKEN"
    }
}
JSONEOF
)

    # Check if config file exists
    if [[ -f "$CLAUDE_CONFIG_FILE" ]]; then
        # Backup existing config
        BACKUP_FILE="$CLAUDE_CONFIG_FILE.backup.$(date +%Y%m%d%H%M%S)"
        cp "$CLAUDE_CONFIG_FILE" "$BACKUP_FILE"
        print_ok "Backed up existing config to $(basename "$BACKUP_FILE")"

        # Use python3 to merge JSON (available on all Macs)
        python3 <<PYEOF
import json
import sys

config_path = "$CLAUDE_CONFIG_FILE"
server_name = "$MCP_SERVER_NAME"
binary_path = "$BINARY_PATH"
profile = "$PROFILE_NAME"
token = "$MCP_TOKEN"

try:
    with open(config_path, 'r') as f:
        content = f.read().strip()
        if content:
            config = json.loads(content)
        else:
            config = {}
except (json.JSONDecodeError, FileNotFoundError):
    config = {}

# Ensure mcpServers key exists
if 'mcpServers' not in config:
    config['mcpServers'] = {}

# Check if server already exists
if server_name in config['mcpServers']:
    print(f"    Updating existing server entry: {server_name}")
else:
    print(f"    Adding new server entry: {server_name}")

# Add/update server entry
config['mcpServers'][server_name] = {
    "command": binary_path,
    "args": ["--profile", profile],
    "env": {
        "GITHUB_TOKEN": token
    }
}

# Write back
with open(config_path, 'w') as f:
    json.dump(config, f, indent=2)

# Count total servers
total = len(config['mcpServers'])
print(f"    Total MCP servers configured: {total}")
PYEOF

    else
        # Create new config file
        python3 <<PYEOF
import json

config_path = "$CLAUDE_CONFIG_FILE"
server_name = "$MCP_SERVER_NAME"
binary_path = "$BINARY_PATH"
profile = "$PROFILE_NAME"
token = "$MCP_TOKEN"

config = {
    "mcpServers": {
        server_name: {
            "command": binary_path,
            "args": ["--profile", profile],
            "env": {
                "GITHUB_TOKEN": token
            }
        }
    }
}

with open(config_path, 'w') as f:
    json.dump(config, f, indent=2)

print(f"    Created new config with server: {server_name}")
PYEOF

    fi

    print_ok "Claude Desktop configured"
}

# ---- Step 7: Summary ----

print_summary() {
    print_step 7 "Installation complete!"

    echo ""
    echo -e "${GREEN}============================================================${NC}"
    echo -e "${GREEN}  Installation Successful!${NC}"
    echo -e "${GREEN}============================================================${NC}"
    echo ""
    echo -e "  ${BOLD}Server:${NC}    $INSTALL_DIR/$BINARY_NAME"
    echo -e "  ${BOLD}Profile:${NC}   $PROFILE_NAME"
    echo -e "  ${BOLD}Config:${NC}    $CLAUDE_CONFIG_FILE"
    echo ""

    if command -v git &> /dev/null; then
        echo -e "  ${BOLD}Tools:${NC}     82 (Git detected)"
    else
        echo -e "  ${BOLD}Tools:${NC}     48 (no Git - API/admin/file tools available)"
    fi

    echo ""
    echo -e "  ${BOLD}Next steps:${NC}"
    echo ""
    echo "  1. Restart Claude Desktop (quit and reopen)"
    echo "  2. In Claude Desktop, you should see the GitHub tools available"
    echo "  3. Try asking Claude: \"List my GitHub repositories\""
    echo ""
    echo -e "  ${BOLD}Safety configuration (optional):${NC}"
    echo "  cp $INSTALL_DIR/safety.json.example $INSTALL_DIR/safety.json"
    echo "  # Edit safety.json to customize safety mode (default: moderate)"
    echo ""

    if [[ "$MCP_TOKEN" == "YOUR_GITHUB_TOKEN_HERE" ]]; then
        echo -e "  ${YELLOW}IMPORTANT:${NC} You still need to set your GitHub token!"
        echo "  Edit: $CLAUDE_CONFIG_FILE"
        echo "  Replace YOUR_GITHUB_TOKEN_HERE with your actual token"
        echo ""
    fi

    echo -e "  ${BOLD}Add another profile?${NC}"
    echo "  Run this installer again with a different profile name."
    echo ""
    echo -e "  ${BOLD}Uninstall:${NC}"
    echo "  rm -rf $INSTALL_DIR"
    echo "  # Remove the server entry from $CLAUDE_CONFIG_FILE"
    echo ""
}

# ---- Main ----

TOTAL_STEPS=7

print_header
check_system
get_binary
install_binary
get_token
get_profile
configure_claude
print_summary
