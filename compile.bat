@echo off
echo ============================================
echo   GitHub MCP Server v3.0 - Build (Windows)
echo ============================================
echo.

cd /d "C:\MCPs\clone\mcp-go-github"

REM Verify Go
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed
    exit /b 1
)

echo [1/3] Cleaning dependencies...
go mod tidy

echo [2/3] Compiling for Windows...
go build -ldflags="-s -w" -o github-mcp-server-v3.exe ./cmd/github-mcp-server/main.go
if %errorlevel% neq 0 (
    echo [ERROR] Compilation failed
    exit /b 1
)

echo [3/3] Verifying build...
github-mcp-server-v3.exe --help >nul 2>&1
echo.
echo ============================================
echo   Build successful: github-mcp-server-v3.exe
echo ============================================
echo.
echo   v3.0 Features:
echo   - 82 tools (48 without Git)
echo   - 22 admin tools with safety system
echo   - 4 file operations (no Git required)
echo   - Git auto-detection
echo   - Multi-profile support
echo.
echo   To build for Mac: build-mac.bat
echo.
pause
