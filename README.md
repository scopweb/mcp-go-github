# GitHub MCP Server (Go)

Servidor MCP (Model Context Protocol) para integración con GitHub en Claude Desktop. Implementa operaciones Git locales y GitHub API con modo híbrido.

## Características

- **JSON-RPC 2.0**: Protocolo estándar de comunicación
- **Operaciones Git locales**: `git_status`, `git_commit`, `git_push`, `git_pull`, etc.
- **GitHub API**: Gestión de repositorios, PRs e issues
- **Modo híbrido**: Prioriza Git local para ahorrar tokens
- **Multi-perfil**: Soporte para múltiples cuentas GitHub

## Quick Start

### Requisitos
- Go 1.23.0+
- Git instalado
- Token de GitHub con permisos `repo`

### Compilar

```bash
cd github-go-server-mcp
go mod tidy
go build -o github-mcp-modular.exe main.go
```

### Configurar en Claude Desktop

Edita `~/.claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "github-mcp": {
      "command": "C:\\path\\to\\github-mcp-modular.exe",
      "args": ["--profile", "personal"],
      "env": {
        "GITHUB_TOKEN": "ghp_your_token_here"
      }
    }
  }
}
```

## Herramientas Disponibles

### Git Básicas
- `git_status` - Estado del repositorio
- `git_commit` - Crear commit
- `git_push` / `git_pull` - Sincronizar cambios
- `git_checkout` - Cambiar de rama

### Git Avanzadas
- `git_merge` - Merge de ramas
- `git_rebase` - Rebase
- `git_safe_merge` - Merge seguro con backup
- `git_detect_conflicts` - Detectar conflictos

### GitHub API
- `github_list_repos` - Listar repositorios
- `github_create_repo` - Crear repositorio
- `github_list_prs` - Listar pull requests
- `github_create_pr` - Crear pull request

## Ventajas del Sistema

- ✅ **Un solo ejecutable** para mantener
- ✅ **Múltiples cuentas GitHub** simultáneas
- ✅ **Operaciones Git locales** sin consumir tokens
- ✅ **Backups automáticos** en operaciones críticas
- ✅ **Detección de conflictos** preventiva

## Solución de Problemas

### Error 403 "Resource not accessible by personal access token"
- Tu token no tiene permisos suficientes
- Genera nuevo token con scope `repo`
- Reinicia Claude Desktop después del cambio

### Logs del servidor
Verifica los logs para confirmar que arranca correctamente:
```
Starting MCP server with profile: personal
```

## Estado del Proyecto

- ✅ **Protocolo JSON-RPC 2.0**: Implementado y funcionando
- ✅ **Operaciones Git locales**: Todas las funciones implementadas
- ✅ **GitHub API**: Repositorios, PRs e issues
- ✅ **Modo híbrido**: Git local + GitHub API
- ✅ **Listo para producción**

## Documentación

Ver [CLAUDE.md](CLAUDE.md) para detalles de arquitectura y desarrollo.
