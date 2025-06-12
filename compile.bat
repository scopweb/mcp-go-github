@echo off
echo 🚀 Compilando GitHub MCP Server...
echo.

REM Cambiar al directorio del proyecto
cd /d "C:\MCPs\clone\github-go-server-mcp"

REM Verificar Go
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Go no está instalado
    exit /b 1
)

REM Limpiar módulo
echo 📦 Limpiando dependencias...
go mod tidy

REM Compilar
echo 🔧 Compilando...
go build -o github-mcp-modular.exe main.go
if %errorlevel% neq 0 (
    echo ❌ Error de compilación
    exit /b 1
)

echo ✅ Compilación exitosa: github-mcp-modular.exe
echo.
echo 💡 Características v2.0:
echo - ✅ Soporte multi-perfil
echo - ✅ Sistema híbrido Git local + GitHub API  
echo - ✅ 15+ herramientas disponibles
echo - ✅ Logs informativos con emojis
echo.
echo 🎯 Para usar:
echo   1. Configurar token(s) en Claude Desktop
echo   2. Usar --profile nombre para diferenciar instancias
echo   3. Reiniciar Claude Desktop
echo.
pause
