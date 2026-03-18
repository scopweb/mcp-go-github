@echo off
echo ============================================
echo   GitHub MCP Server v4.0 - Build (Windows)
echo ============================================
echo.

REM Verify Go
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed
    exit /b 1
)

echo [1/3] Cleaning dependencies...
go mod tidy

echo [2/3] Compiling for Windows...
go build -ldflags="-s -w" -o github-mcp-server-v4.exe ./cmd/github-mcp-server/main.go
if %errorlevel% neq 0 (
    echo [ERROR] Compilation failed
    exit /b 1
)

echo [3/3] Verifying build...
github-mcp-server-v4.exe --help >nul 2>&1
echo.
echo ============================================
echo   Build successful: github-mcp-server-v4.exe
echo ============================================
echo.
echo   v4.0 Features:
echo   - 26 consolidated tools (85 operations)
echo   - Operation parameter pattern (reduces AI confusion)
echo   - 4 admin tools with safety system (22 operations)
echo   - Git auto-detection + hybrid mode
echo   - Multi-profile support
echo.
echo   To build for Mac: build-mac.bat
echo.
pause
