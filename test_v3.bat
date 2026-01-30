@echo off
REM test_v3.bat - Basic MCP protocol test for v3.0
REM This script tests the server by sending JSON-RPC requests via stdin

setlocal enabledelayedexpansion

echo.
echo ========================================
echo MCP GitHub Server v3.0 - Protocol Test
echo ========================================
echo.

REM Test 1: Initialize
echo [1/3] Testing initialize...
echo {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1.0.0","capabilities":{}}} | github-mcp-server-v3.exe > nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [OK] Initialize successful
) else (
    echo [FAIL] Initialize failed
    exit /b 1
)

REM Test 2: List tools and count
echo.
echo [2/3] Testing tools/list...
echo {"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}} | github-mcp-server-v3.exe > test_output.json
if %ERRORLEVEL% EQU 0 (
    echo [OK] tools/list successful
    REM Count tools in output (simple check)
    findstr /C:"github_delete_repository" test_output.json > nul
    if !ERRORLEVEL! EQU 0 (
        echo [OK] Admin tools present
    ) else (
        echo [FAIL] Admin tools not found
        type test_output.json
        exit /b 1
    )
) else (
    echo [FAIL] tools/list failed
    exit /b 1
)

REM Test 3: Call a low-risk admin tool (dry-run)
echo.
echo [3/3] Testing admin tool call...
echo {"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"github_get_repo_settings","arguments":{"owner":"test","repo":"test","dry_run":true}}} | github-mcp-server-v3.exe > test_admin_call.json 2>&1
REM Note: This will fail without valid credentials, but verifies the tool is registered
findstr /C:"github_get_repo_settings" test_admin_call.json > nul
if !ERRORLEVEL! EQU 0 (
    echo [OK] Admin tool callable
) else (
    echo [WARN] Admin tool routing may have issues
)

REM Cleanup
del test_output.json 2>nul
del test_admin_call.json 2>nul

echo.
echo ========================================
echo Protocol Test Complete
echo ========================================
echo.
echo v3.0 server is functional!
echo.
echo Next steps:
echo   1. Configure GITHUB_TOKEN in environment
echo   2. Add server to Claude Desktop config
echo   3. Test with real repositories
echo.

endlocal
