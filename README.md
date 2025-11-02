# GitHub MCP Server 🚀

Go-based MCP server that connects GitHub to Claude Desktop, enabling direct repository operations from Claude's interface.

**🎯 Latest Update:** Updated to `go-github v76.0.0` for enhanced stability and compatibility.

## ✨ Nuevas Características

### 🎯 **Soporte de Perfiles Múltiples**
- **Un solo ejecutable** para múltiples cuentas GitHub
- **Configuración diferenciada** por perfil
- **Logs informativos** con identificación de perfil
- **Gestión simplificada** de tokens

## 📋 Permisos Necesarios del Token

Para que todas las funciones trabajen correctamente, tu **GitHub Personal Access Token** debe tener estos permisos:

### 🔑 Mínimos Requeridos:
```
✅ repo (Full control of private repositories)
  - Necesario para crear repos, issues, PRs
  - Permite lectura/escritura en repositorios
```

### 🔧 Opcionales (para funcionalidad completa):
```
✅ delete_repo (Delete repositories) - Solo si necesitas borrar repos
✅ workflow (Update GitHub Action workflows) - Para trabajar con Actions
✅ admin:repo_hook (Repository hooks) - Para webhooks
```

### 📝 Generar Token:
1. Ve a: [GitHub Settings → Personal Access Tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Selecciona los scopes necesarios arriba
4. Copia el token generado

## 🛠️ Instalación

```bash
# Instalar dependencias
go mod tidy

# Compilar (usando el script incluido)
.\compile.bat

# O compilar manualmente
go build -o github-mcp-modular.exe .
```

## 🧪 Testing

El proyecto incluye tests unitarios completos:

```bash
# Ejecutar todos los tests
go test ./...

# Ejecutar tests con verbose
go test ./... -v

# Ejecutar tests de un módulo específico
go test ./internal/hybrid/ -v
```

## ⚙️ Configuración Claude Desktop

### 🔥 **Configuración con Perfiles Múltiples** (Recomendado)

```json
{
  "mcpServers": {
    "github-personal": {
      "command": "C:\\MCPs\\clone\\github-go-server-mcp\\github-mcp-modular.exe",
      "args": ["--profile", "personal"],
      "env": {
        "GITHUB_TOKEN": "ghp_token_personal"
      }
    },
    "github-empresa": {
      "command": "C:\\MCPs\\clone\\github-go-server-mcp\\github-mcp-modular.exe",
      "args": ["--profile", "empresa"],
      "env": {
        "GITHUB_TOKEN": "ghp_token_empresa"
      }
    }
  }
}
```

### 📦 **Configuración Básica** (Un solo token)

```json
{
  "mcpServers": {
    "github-mcp": {
      "command": "C:\\MCPs\\clone\\github-go-server-mcp\\github-mcp-modular.exe",
      "args": [],
      "env": {
        "GITHUB_TOKEN": "tu_token_aqui_con_permisos_repo"
      }
    }
  }
}
```

## 🧪 Herramientas Disponibles (45+ Herramientas)

### 🔍 Herramientas de Información Git

| Herramienta | Descripción | Tokens |
|-------------|-------------|--------|
| **git_status** | Estado del repositorio Git local y configuración | 0 |
| **git_list_files** | Lista todos los archivos en el repositorio | 0 |
| **git_get_file_content** | Obtiene contenido de un archivo desde Git | 0 |
| **git_get_file_sha** | Obtiene el SHA de un archivo específico | 0 |
| **git_get_last_commit** | Obtiene el SHA del último commit | 0 |
| **git_get_changed_files** | Lista archivos modificados (working/staging) | 0 |
| **git_validate_repo** | Valida si un directorio es un repositorio Git válido | 0 |
| **git_context** | Auto-detecta contexto Git para optimizar tokens | 0 |

### ⚙️ Operaciones Git Básicas

| Herramienta | Descripción | Tokens |
|-------------|-------------|--------|
| **git_set_workspace** | Configura el directorio de trabajo para Git | 0 |
| **git_add** | Agrega archivos al staging area | 0 |
| **git_commit** | Hace commit de los cambios en staging | 0 |
| **git_push** | Sube cambios al repositorio remoto | 0 |
| **git_pull** | Baja cambios del repositorio remoto | 0 |
| **git_checkout** | Cambia de rama o crea nueva rama | 0 |

### 📊 Análisis y Gestión Git

| Herramienta | Descripción | Tokens |
|-------------|-------------|--------|
| **git_log_analysis** | Análisis completo del historial de commits | 0 |
| **git_diff_files** | Muestra archivos modificados con estadísticas | 0 |
| **git_branch_list** | Lista todas las ramas con información detallada | 0 |
| **git_stash** | Operaciones de stash (list, push, pop, apply, drop, clear) | 0 |
| **git_remote** | Gestión de repositorios remotos (list, add, remove, show, fetch) | 0 |
| **git_tag** | Gestión de tags/etiquetas (list, create, delete, push, show) | 0 |
| **git_clean** | Limpieza de archivos sin seguimiento | 0 |

### 🚀 Operaciones Git Avanzadas

| Herramienta | Descripción | Tokens |
|-------------|-------------|--------|
| **git_checkout_remote** | Checkout de rama remota con tracking local | 0 |
| **git_merge** | Merge de ramas con validaciones de seguridad | 0 |
| **git_rebase** | Rebase con rama especificada | 0 |
| **git_pull_with_strategy** | Pull con estrategias (merge, rebase, ff-only) | 0 |
| **git_force_push** | Push con --force-with-lease (con backup automático) | 0 |
| **git_push_upstream** | Push configurando upstream tracking | 0 |
| **git_sync_with_remote** | Sincronización automática con rama remota | 0 |

### 🛡️ Gestión de Conflictos

| Herramienta | Descripción | Tokens |
|-------------|-------------|--------|
| **git_safe_merge** | Merge seguro con backup y detección de conflictos | 0 |
| **git_conflict_status** | Estado detallado de conflictos en merge/rebase | 0 |
| **git_resolve_conflicts** | Resolución automática con estrategias (theirs, ours, abort) | 0 |
| **git_validate_clean_state** | Valida que el working directory esté limpio | 0 |
| **git_detect_conflicts** | Detecta conflictos potenciales entre ramas | 0 |
| **git_create_backup** | Crea backup/tag del estado actual | 0 |

### 🔀 Operaciones Híbridas (Git Local → GitHub API)

| Herramienta | Descripción | Tokens |
|-------------|-------------|--------|
| **create_file** | Crea archivo PRIORIZANDO Git local sobre GitHub API | 0* |
| **update_file** | Actualiza archivo PRIORIZANDO Git local sobre GitHub API | 0* |

*Usa 0 tokens si Git local está disponible, fallback a GitHub API si es necesario

### 🌐 GitHub API (Operaciones en la Nube)

| Herramienta | Descripción | Tokens |
|-------------|-------------|--------|
| **github_list_repos** | Lista repositorios del usuario | ✓ |
| **github_create_repo** | Crea nuevo repositorio | ✓ |
| **github_list_prs** | Lista pull requests | ✓ |
| **github_create_pr** | Crea nuevo pull request | ✓ |

## 🚀 Uso

1. **Compilar el servidor**: `.\compile.bat`
2. **Generar token(s) GitHub** con permisos `repo`
3. **Configurar Claude Desktop** con perfiles
4. **Reiniciar Claude Desktop**
5. **Verificar logs** para confirmar inicio correcto

## 💡 Ventajas del Sistema de Perfiles

- ✅ **Un solo ejecutable** para mantener
- ✅ **Múltiples cuentas GitHub** simultáneas
- ✅ **Logs diferenciados** por perfil
- ✅ **Actualizaciones automáticas** para todas las instancias
- ✅ **Configuración más limpia**

## ⚠️ Solución de Problemas

### Error 403 "Resource not accessible by personal access token"
- ❌ Tu token no tiene permisos suficientes
- ✅ Genera nuevo token con scope `repo`
- ✅ Reinicia Claude Desktop después del cambio

### Error "null" en respuestas
- ⚠️ Normal para repos vacíos o sin PRs/issues
- ✅ El MCP funciona correctamente

### Logs del servidor
Verifica los logs de Claude Desktop para ver mensajes como:
```
🚀 Starting GitHub MCP Server with profile: personal
📋 Profile: personal | Token: ghp_111***
🔧 Git environment detected for profile: personal
```

## 🔒 Mejoras de Seguridad (Implementadas por GitHub Copilot)

GitHub Copilot, ha realizado una revisión y fortalecimiento de la seguridad de este MCP. Se han implementado las siguientes mejoras clave para garantizar que el servidor sea más robusto y seguro contra posibles ataques:

-   **Prevención de Inyección de Argumentos**: Se ha neutralizado el riesgo de que un atacante pueda inyectar comandos no deseados (como `--force`) a través de los argumentos de las herramientas `git`.
-   **Defensa contra "Path Traversal"**: Se ha añadido una capa de validación que impide el acceso a archivos o directorios fuera del repositorio de trabajo, protegiendo la integridad del sistema.
-   **Validación Estricta de Entradas**: El servidor ahora verifica rigurosamente los datos de entrada, rechazando cualquier solicitud con argumentos mal formados o ausentes antes de que pueda causar un comportamiento inesperado.

Con estos cambios, el MCP es ahora mucho más seguro. ¡Un saludo, amigo!

## 📊 Estado del Proyecto

- ✅ **Funciones de lectura**: Completamente operativas
- ✅ **Funciones de escritura**: Completamente operativas  
- ✅ **Sistema híbrido Git**: Git local + GitHub API
- ✅ **Soporte multi-perfil**: Implementado y testeado
- ✅ **Gestión de permisos**: Documentada y verificada
- ✅ **Testing completo**: Todas las funciones probadas con tests unitarios
- ✅ **Dependencias actualizadas**: go-github v76.0.0, oauth2 v0.32.0
- ✅ **45+ Herramientas Git**: Operaciones locales (0 tokens) y avanzadas
- ✅ **Gestión de conflictos**: Merge seguro, detección y resolución automática
- ✅ **Listo para producción**: Stable release v2.0

📋 **Changelog**: Ver [CHANGELOG.md](CHANGELOG.md) para historial completo de cambios

## 🔧 Requisitos del Sistema

- **Go**: 1.24.0 o superior (actualizado)
- **Git**: Para operaciones locales (opcional pero recomendado)
- **Windows**: PowerShell para scripts de compilación
- **GitHub Token**: Con permisos `repo` mínimos


 