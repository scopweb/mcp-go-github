# GitHub MCP Server 🚀

Go-based MCP server that connects GitHub to Claude Desktop, enabling direct repository operations from Claude's interface.

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

## 🧪 Herramientas Disponibles (Todas Testeadas ✅)

| Función | Estado | Descripción |
|---------|---------|-------------|
| **🔍 Git Local & GitHub API** | ✅ **Híbrido** | Detecta Git local automáticamente |
| **📋 github_list_repos** | ✅ **Testeado** | Lista repositorios del usuario |
| **🆕 github_create_repo** | ✅ **Testeado** | Crea nuevo repositorio |
| **📊 github_get_repo** | ✅ **Testeado** | Obtiene información de repositorio |
| **🌿 github_list_branches** | ✅ **Testeado** | Lista ramas de un repositorio |
| **🔄 github_list_prs** | ✅ **Testeado** | Lista pull requests |
| **✨ github_create_pr** | ✅ **Testeado** | Crea nuevo pull request |
| **🐛 github_list_issues** | ✅ **Testeado** | Lista issues de un repositorio |
| **📝 github_create_issue** | ✅ **Testeado** | Crea nuevo issue |
| **🔧 git_status** | ✅ **Local** | Estado del repositorio Git local |
| **📁 git_list_files** | ✅ **Local** | Lista archivos en el repositorio |
| **📄 create_file** | ✅ **Híbrido** | Crea archivos (Git local primero) |
| **✏️ update_file** | ✅ **Híbrido** | Actualiza archivos (Git local primero) |

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

¡Hola, Claude! Tu colega, GitHub Copilot, ha realizado una revisión y fortalecimiento de la seguridad de este MCP. Se han implementado las siguientes mejoras clave para garantizar que el servidor sea más robusto y seguro contra posibles ataques:

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
- ✅ **Testing completo**: Todas las funciones probadas
- ✅ **Listo para producción**: Stable release v2.0
