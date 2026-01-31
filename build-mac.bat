@echo off
echo ============================================
echo   GitHub MCP Server v3.0 - Build for Mac
echo ============================================
echo.

cd /d "C:\MCPs\clone\mcp-go-github"

REM Verify Go
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed
    exit /b 1
)

echo [1/4] Cleaning dependencies...
go mod tidy

REM Build for Apple Silicon (M1/M2/M3/M4)
echo [2/4] Building for macOS ARM64 (Apple Silicon)...
set GOOS=darwin
set GOARCH=arm64
set CGO_ENABLED=0
go build -ldflags="-s -w" -o dist/mac-arm64/github-mcp-server-v3 ./cmd/github-mcp-server/main.go
if %errorlevel% neq 0 (
    echo [ERROR] Failed to build for ARM64
    exit /b 1
)
echo       OK: dist/mac-arm64/github-mcp-server-v3

REM Build for Intel Mac
echo [3/4] Building for macOS AMD64 (Intel)...
set GOOS=darwin
set GOARCH=amd64
set CGO_ENABLED=0
go build -ldflags="-s -w" -o dist/mac-amd64/github-mcp-server-v3 ./cmd/github-mcp-server/main.go
if %errorlevel% neq 0 (
    echo [ERROR] Failed to build for AMD64
    exit /b 1
)
echo       OK: dist/mac-amd64/github-mcp-server-v3

REM Copy installer and config to dist folders
echo [4/4] Packaging installer files...
copy install-mac.sh dist\mac-arm64\ >nul 2>&1
copy install-mac.sh dist\mac-amd64\ >nul 2>&1
copy safety.json.example dist\mac-arm64\ >nul 2>&1
copy safety.json.example dist\mac-amd64\ >nul 2>&1

REM Reset GOOS for Windows
set GOOS=windows
set GOARCH=amd64

echo.
echo ============================================
echo   Build complete!
echo ============================================
echo.
echo   Apple Silicon (M1/M2/M3/M4):
echo     dist\mac-arm64\github-mcp-server-v3
echo.
echo   Intel Mac:
echo     dist\mac-amd64\github-mcp-server-v3
echo.
echo   To install on Mac:
echo     1. Copy the correct folder to the Mac
echo     2. Run: chmod +x install-mac.sh
echo     3. Run: ./install-mac.sh
echo.
pause
