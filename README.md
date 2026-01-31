# GitHub MCP Server v3.0

Go-based MCP server that connects GitHub to Claude Desktop, enabling direct repository operations from Claude's interface.

**Tools:** 82 (with Git) | 48 (without Git) | **Architecture:** Hybrid (Local Git + GitHub API + Admin Controls)

## What's New in v3.0

- **22 Administrative Tools**: Repository settings, branch protection, webhooks, collaborators, teams
- **4-Tier Safety System**: Risk classification (LOW/MEDIUM/HIGH/CRITICAL) with confirmation tokens
- **Git-Free File Operations**: Clone, pull, download repos via GitHub API (no Git required)
- **Smart Git Detection**: Auto-detects Git availability, filters tools accordingly
- **Audit Logging**: JSON-based operation tracking with automatic rotation

## Permisos Necesarios del Token

### Minimos Requeridos:
```
repo        - Full control of private repositories (essential)
```

### Opcionales (para funcionalidad completa):
```
delete_repo      - Para github_delete_repository
workflow         - Para re-run de GitHub Actions workflows
security_events  - Para dismissar alertas de seguridad
admin:repo_hook  - Enhanced webhook management (v3.0)
admin:org        - Para team management en organizaciones (v3.0)
```

### Generar Token:
1. Ve a: [GitHub Settings > Personal Access Tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Selecciona los scopes necesarios
4. Copia el token generado

## Instalacion

```bash
# Instalar dependencias
go mod tidy

# Compilar (usando el script incluido)
.\compile.bat

# O compilar manualmente
go build -o github-mcp-server-v3.exe ./cmd/github-mcp-server/
```

## Testing

```bash
# Ejecutar todos los tests
go test ./...

# Ejecutar tests con verbose
go test ./... -v

# Ejecutar tests de un paquete especifico
go test ./pkg/git/ -v
go test ./pkg/safety/ -v
```

## Configuracion Claude Desktop

### Multi-profile (Recomendado)

```json
{
  "mcpServers": {
    "github-personal": {
      "command": "path\\to\\github-mcp-server-v3.exe",
      "args": ["--profile", "personal"],
      "env": {
        "GITHUB_TOKEN": "ghp_token_personal"
      }
    },
    "github-work": {
      "command": "path\\to\\github-mcp-server-v3.exe",
      "args": ["--profile", "work"],
      "env": {
        "GITHUB_TOKEN": "ghp_token_work"
      }
    }
  }
}
```

### Configuracion Basica (Un solo token)

```json
{
  "mcpServers": {
    "github-mcp": {
      "command": "path\\to\\github-mcp-server-v3.exe",
      "args": [],
      "env": {
        "GITHUB_TOKEN": "tu_token_aqui"
      }
    }
  }
}
```

## Herramientas Disponibles (82 Tools)

### Informacion Git (8)

| Herramienta | Descripcion | Tokens |
|-------------|-------------|--------|
| `git_status` | Estado del repositorio Git local | 0 |
| `git_list_files` | Lista todos los archivos en el repositorio | 0 |
| `git_get_file_content` | Obtiene contenido de un archivo desde Git | 0 |
| `git_get_file_sha` | Obtiene el SHA de un archivo especifico | 0 |
| `git_get_last_commit` | Obtiene el SHA del ultimo commit | 0 |
| `git_get_changed_files` | Lista archivos modificados | 0 |
| `git_validate_repo` | Valida si un directorio es un repo Git valido | 0 |
| `git_context` | Auto-detecta contexto Git | 0 |

### Operaciones Git Basicas (6)

| Herramienta | Descripcion | Tokens |
|-------------|-------------|--------|
| `git_set_workspace` | Configura directorio de trabajo | 0 |
| `git_add` | Agrega archivos al staging area | 0 |
| `git_commit` | Hace commit de los cambios | 0 |
| `git_push` | Sube cambios al remoto | 0 |
| `git_pull` | Baja cambios del remoto | 0 |
| `git_checkout` | Cambia de rama o crea nueva | 0 |

### Analisis y Gestion Git (7)

| Herramienta | Descripcion | Tokens |
|-------------|-------------|--------|
| `git_log_analysis` | Analisis del historial de commits | 0 |
| `git_diff_files` | Muestra archivos modificados con estadisticas | 0 |
| `git_branch_list` | Lista ramas con informacion detallada | 0 |
| `git_stash` | Operaciones de stash | 0 |
| `git_remote` | Gestion de repositorios remotos | 0 |
| `git_tag` | Gestion de tags/etiquetas | 0 |
| `git_clean` | Limpieza de archivos sin seguimiento | 0 |

### Operaciones Git Avanzadas (7)

| Herramienta | Descripcion | Tokens |
|-------------|-------------|--------|
| `git_checkout_remote` | Checkout de rama remota con tracking | 0 |
| `git_merge` | Merge de ramas con validaciones | 0 |
| `git_rebase` | Rebase con rama especificada | 0 |
| `git_pull_with_strategy` | Pull con estrategias | 0 |
| `git_force_push` | Push con --force-with-lease | 0 |
| `git_push_upstream` | Push configurando upstream | 0 |
| `git_sync_with_remote` | Sincronizacion con rama remota | 0 |

### Gestion de Conflictos (6)

| Herramienta | Descripcion | Tokens |
|-------------|-------------|--------|
| `git_safe_merge` | Merge seguro con backup | 0 |
| `git_conflict_status` | Estado de conflictos | 0 |
| `git_resolve_conflicts` | Resolucion automatica | 0 |
| `git_validate_clean_state` | Valida working directory limpio | 0 |
| `git_detect_conflicts` | Detecta conflictos potenciales | 0 |
| `git_create_backup` | Crea backup del estado actual | 0 |

### Operaciones Hibridas (2)

| Herramienta | Descripcion | Tokens |
|-------------|-------------|--------|
| `create_file` | Crea archivo (Git local, fallback API) | 0* |
| `update_file` | Actualiza archivo (Git local, fallback API) | 0* |

### GitHub API (4)

| Herramienta | Descripcion |
|-------------|-------------|
| `github_list_repos` | Lista repositorios del usuario |
| `github_create_repo` | Crea nuevo repositorio |
| `github_list_prs` | Lista pull requests |
| `github_create_pr` | Crea nuevo pull request |

### File Operations - Sin Git (4) [NEW v3.0]

| Herramienta | Descripcion |
|-------------|-------------|
| `github_list_repo_contents` | Listar archivos y directorios via API |
| `github_download_file` | Descargar archivo individual |
| `github_download_repo` | Clonar repositorio completo via API |
| `github_pull_repo` | Actualizar directorio local via API |

### Dashboard (7)

| Herramienta | Descripcion |
|-------------|-------------|
| `github_dashboard` | Panel general de actividad |
| `github_notifications` | Notificaciones pendientes |
| `github_assigned_issues` | Issues asignados |
| `github_prs_to_review` | PRs pendientes de review |
| `github_security_alerts` | Alertas de seguridad |
| `github_failed_workflows` | Workflows fallidos |
| `github_mark_notification_read` | Marcar notificacion como leida |

### Response (3)

| Herramienta | Descripcion |
|-------------|-------------|
| `github_comment_issue` | Comentar en issue |
| `github_comment_pr` | Comentar en pull request |
| `github_review_pr` | Crear review en PR (APPROVE/REQUEST_CHANGES/COMMENT) |

### Repair (6)

| Herramienta | Descripcion |
|-------------|-------------|
| `github_close_issue` | Cerrar issue |
| `github_merge_pr` | Mergear pull request |
| `github_rerun_workflow` | Re-ejecutar workflow |
| `github_dismiss_dependabot_alert` | Dismissar alerta Dependabot |
| `github_dismiss_code_alert` | Dismissar alerta Code Scanning |
| `github_dismiss_secret_alert` | Dismissar alerta Secret Scanning |

### Repository Admin (4) [NEW v3.0]

| Herramienta | Riesgo | Descripcion |
|-------------|--------|-------------|
| `github_get_repo_settings` | LOW | Ver configuracion del repositorio |
| `github_update_repo_settings` | MEDIUM | Modificar nombre, descripcion, visibilidad |
| `github_archive_repository` | CRITICAL | Archivar repositorio (read-only) |
| `github_delete_repository` | CRITICAL | Eliminar repositorio PERMANENTEMENTE |

### Branch Protection (3) [NEW v3.0]

| Herramienta | Riesgo | Descripcion |
|-------------|--------|-------------|
| `github_get_branch_protection` | LOW | Ver reglas de proteccion |
| `github_update_branch_protection` | HIGH | Configurar reglas de proteccion |
| `github_delete_branch_protection` | CRITICAL | Eliminar proteccion de branch |

### Webhooks (5) [NEW v3.0]

| Herramienta | Riesgo | Descripcion |
|-------------|--------|-------------|
| `github_list_webhooks` | LOW | Listar webhooks del repo |
| `github_create_webhook` | MEDIUM | Crear webhook |
| `github_update_webhook` | MEDIUM | Modificar webhook |
| `github_delete_webhook` | HIGH | Eliminar webhook |
| `github_test_webhook` | LOW | Enviar test delivery |

### Collaborators (8) [NEW v3.0]

| Herramienta | Riesgo | Descripcion |
|-------------|--------|-------------|
| `github_list_collaborators` | LOW | Listar colaboradores |
| `github_check_collaborator` | LOW | Verificar acceso |
| `github_add_collaborator` | MEDIUM | Invitar con permisos |
| `github_update_collaborator_permission` | MEDIUM | Cambiar nivel de acceso |
| `github_remove_collaborator` | HIGH | Revocar acceso |
| `github_list_invitations` | LOW | Ver invitaciones pendientes |
| `github_accept_invitation` | MEDIUM | Aceptar invitacion |
| `github_cancel_invitation` | MEDIUM | Cancelar invitacion |

### Teams (2) [NEW v3.0]

| Herramienta | Riesgo | Descripcion |
|-------------|--------|-------------|
| `github_list_repo_teams` | LOW | Listar teams con acceso |
| `github_add_repo_team` | MEDIUM | Otorgar acceso de team |

## Safety System (v3.0)

### 4 Niveles de Riesgo

| Nivel | Descripcion | Comportamiento (modo moderate) |
|-------|-------------|-------------------------------|
| **LOW** | Solo lectura | Ejecucion directa |
| **MEDIUM** | Cambios reversibles | Dry-run opcional |
| **HIGH** | Impacta colaboracion | Requiere token de confirmacion |
| **CRITICAL** | Irreversible | Token + recomendacion de backup |

### Modos de Seguridad

| Modo | Confirma desde | Uso recomendado |
|------|---------------|-----------------|
| `strict` | MEDIUM+ | Entornos de produccion criticos |
| `moderate` | HIGH+ | Uso general (default) |
| `permissive` | CRITICAL | Desarrollo local |
| `disabled` | Nunca | No recomendado |

### Configuracion de Seguridad

Crear `safety.json` junto al ejecutable (opcional):

```json
{
  "mode": "moderate",
  "enable_audit_log": true,
  "require_confirmation_above": 3,
  "audit_log_path": "./mcp-admin-audit.log",
  "audit_log_max_size_mb": 10,
  "audit_log_max_backups": 5
}
```

Si no existe `safety.json`, usa modo **moderate** con audit logging habilitado.

Ver `safety.json.example` para configuracion completa de referencia.

## Git-Free Mode (v3.0)

En sistemas sin Git instalado (ej: Mac sin Xcode Command Line Tools), el servidor:

1. **Detecta automaticamente** la ausencia de Git
2. **Filtra** las herramientas git_ del listado
3. **Mantiene operativas** todas las herramientas API, admin, dashboard, file operations
4. **Retorna error amigable** si se intenta usar una herramienta Git

Las 4 herramientas de File Operations (`github_list_repo_contents`, `github_download_file`, `github_download_repo`, `github_pull_repo`) permiten clonar y actualizar repositorios usando solo la API de GitHub, sin necesidad de Git.

## Seguridad

- Prevencion de inyeccion de argumentos en comandos Git
- Defensa contra Path Traversal
- Validacion estricta de entradas de usuario
- Prevencion de SSRF en URLs de webhooks (v3.0)
- Tokens de confirmacion criptograficos para operaciones destructivas (v3.0)
- Audit logging de operaciones administrativas (v3.0)

## Requisitos del Sistema

- **Go**: 1.25.0 o superior
- **Git**: Opcional (auto-detectado, 48 tools funcionan sin Git)
- **github.com/google/go-github**: v81.0.0
- **golang.org/x/oauth2**: v0.34.0
- **GitHub Token**: Con permiso `repo` minimo

## Estado del Proyecto

- 82 herramientas MCP operativas (48 sin Git)
- Sistema hibrido Git local + GitHub API
- 22 herramientas administrativas con safety layer
- 4 herramientas de archivo sin Git
- Soporte multi-perfil
- Testing completo con repositorio real
- Listo para produccion (v3.0)

**Changelog**: Ver [CHANGELOG.md](CHANGELOG.md) para historial completo de cambios
