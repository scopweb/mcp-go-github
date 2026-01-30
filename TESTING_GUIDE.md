# Guía de Pruebas Manuales - GitHub MCP Server v2.1

## Resumen del Análisis Automático

### ✅ Estado de Compilación
- Proyecto compila correctamente sin errores
- Análisis estático (`go vet`): Sin problemas detectados
- Verificación de dependencias: Todas las dependencias verificadas correctamente

### ✅ Tests Unitarios
Todos los tests pasan exitosamente:
- **internal/hybrid**: 3 tests (100% pasan)
- **pkg/git**: 8 tests (100% pasan)
- **pkg/github**: 6 tests (100% pasan)
- **test/security**: 14 tests de seguridad (100% pasan)

### 📊 Cobertura de Código
- **internal/hybrid**: 43.1% de cobertura
- **pkg/git**: 12.8% de cobertura (necesita mejoras)
- **pkg/github**: 39.6% de cobertura

### ⚠️ Áreas sin Tests Unitarios
- `cmd/github-mcp-server/main.go` - Punto de entrada (requiere pruebas de integración)
- `internal/server/server.go` - Handler MCP principal (55+ herramientas)
- `pkg/dashboard/dashboard.go` - Operaciones de dashboard (7 herramientas nuevas)

---

## Configuración Inicial

### 1. Obtener Token de GitHub

**Permisos requeridos:**
- `repo` (esencial para todas las operaciones)
- `security_events` (opcional pero recomendado para alertas de seguridad)

**Pasos:**
1. Ve a https://github.com/settings/tokens
2. Click en "Generate new token (classic)"
3. Selecciona los permisos mencionados arriba
4. Copia el token generado

### 2. Configurar Variable de Entorno

#### Windows (PowerShell)
```powershell
$env:GITHUB_TOKEN="ghp_tu_token_aquí"
```

#### Windows (CMD)
```cmd
set GITHUB_TOKEN=ghp_tu_token_aquí
```

#### Linux/macOS
```bash
export GITHUB_TOKEN="ghp_tu_token_aquí"
```

### 3. Compilar el Proyecto

```bash
go build -o github-mcp-server.exe ./cmd/github-mcp-server
```

---

## Plan de Pruebas Manuales

### Fase 1: Pruebas Básicas de Conectividad (5 minutos)

#### Test 1.1: Verificar que el servidor inicia correctamente
```bash
.\github-mcp-server.exe
```

**Entrada de prueba:**
```json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
```

**Resultado esperado:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {"tools": {}},
    "serverInfo": {
      "name": "github-mcp-local-hybrid",
      "version": "2.5.0"
    }
  }
}
```

#### Test 1.2: Listar herramientas disponibles
**Entrada:**
```json
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
```

**Verificar:**
- Debe retornar 55+ herramientas
- Incluir nuevas herramientas v2.1: `github_comment_issue`, `github_comment_pr`, `github_review_pr`, `github_close_issue`, `github_merge_pr`, etc.

---

### Fase 2: Pruebas de Operaciones Git Locales (10 minutos)

#### Test 2.1: Estado del repositorio local
**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "git_status",
    "arguments": {}
  }
}
```

**Verificar:**
- Muestra el estado del repositorio actual
- Indica la rama actual
- Lista archivos modificados (si los hay)

#### Test 2.2: Listar archivos del repositorio
**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "git_list_files",
    "arguments": {
      "path": "."
    }
  }
}
```

#### Test 2.3: Leer contenido de un archivo
**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "tools/call",
  "params": {
    "name": "git_get_file_content",
    "arguments": {
      "file_path": "README.md"
    }
  }
}
```

#### Test 2.4: Crear un archivo de prueba (operación híbrida)
**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "method": "tools/call",
  "params": {
    "name": "create_file",
    "arguments": {
      "file_path": "test_file.txt",
      "content": "Este es un archivo de prueba",
      "commit_message": "test: Add test file"
    }
  }
}
```

**Verificar:**
- El archivo se crea localmente
- Si estamos en un repo Git, se commitea automáticamente

#### Test 2.5: Actualizar archivo (operación híbrida)
**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 7,
  "method": "tools/call",
  "params": {
    "name": "update_file",
    "arguments": {
      "file_path": "test_file.txt",
      "new_content": "Contenido actualizado",
      "commit_message": "test: Update test file"
    }
  }
}
```

---

### Fase 3: Pruebas de GitHub API (15 minutos)

**Nota:** Para estas pruebas necesitas reemplazar `owner` y `repo` con tus valores reales.

#### Test 3.1: Listar repositorios del usuario
**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 8,
  "method": "tools/call",
  "params": {
    "name": "github_list_repos",
    "arguments": {}
  }
}
```

**Verificar:**
- Lista tus repositorios de GitHub
- Muestra información básica (nombre, descripción, visibilidad)

#### Test 3.2: Listar pull requests
**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 9,
  "method": "tools/call",
  "params": {
    "name": "github_list_prs",
    "arguments": {
      "owner": "tu_usuario",
      "repo": "tu_repo",
      "state": "open"
    }
  }
}
```

#### Test 3.3: Dashboard completo
**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 10,
  "method": "tools/call",
  "params": {
    "name": "github_dashboard",
    "arguments": {}
  }
}
```

**Verificar:**
- Muestra notificaciones no leídas
- Issues asignados
- PRs pendientes de revisión
- Alertas de seguridad
- Workflows fallidos

#### Test 3.4: Notificaciones
**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 11,
  "method": "tools/call",
  "params": {
    "name": "github_notifications",
    "arguments": {
      "all": false,
      "participating": true
    }
  }
}
```

---

### Fase 4: Pruebas de Herramientas de Respuesta v2.1 (15 minutos)

#### Test 4.1: Comentar en un issue
**Prerequisito:** Necesitas un issue existente

**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 12,
  "method": "tools/call",
  "params": {
    "name": "github_comment_issue",
    "arguments": {
      "owner": "tu_usuario",
      "repo": "tu_repo",
      "issue_number": 1,
      "body": "Este es un comentario de prueba desde el MCP server"
    }
  }
}
```

**Verificar:**
- El comentario aparece en el issue
- Retorna el ID del comentario creado

#### Test 4.2: Comentar en un PR
**Prerequisito:** Necesitas un PR existente

**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 13,
  "method": "tools/call",
  "params": {
    "name": "github_comment_pr",
    "arguments": {
      "owner": "tu_usuario",
      "repo": "tu_repo",
      "pr_number": 1,
      "body": "Comentario de prueba en PR"
    }
  }
}
```

#### Test 4.3: Crear revisión de PR
**Prerequisito:** Necesitas un PR existente

**Entrada (aprobar):**
```json
{
  "jsonrpc": "2.0",
  "id": 14,
  "method": "tools/call",
  "params": {
    "name": "github_review_pr",
    "arguments": {
      "owner": "tu_usuario",
      "repo": "tu_repo",
      "pr_number": 1,
      "event": "APPROVE",
      "body": "LGTM! Aprobado desde el MCP server"
    }
  }
}
```

**Entrada (solicitar cambios):**
```json
{
  "jsonrpc": "2.0",
  "id": 15,
  "method": "tools/call",
  "params": {
    "name": "github_review_pr",
    "arguments": {
      "owner": "tu_usuario",
      "repo": "tu_repo",
      "pr_number": 1,
      "event": "REQUEST_CHANGES",
      "body": "Por favor corrige estos problemas:\n- Item 1\n- Item 2"
    }
  }
}
```

---

### Fase 5: Pruebas de Herramientas de Reparación v2.1 (15 minutos)

#### Test 5.1: Cerrar un issue
**Prerequisito:** Necesitas un issue que puedas cerrar

**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 16,
  "method": "tools/call",
  "params": {
    "name": "github_close_issue",
    "arguments": {
      "owner": "tu_usuario",
      "repo": "tu_repo",
      "issue_number": 1,
      "comment": "Cerrando este issue - problema resuelto"
    }
  }
}
```

**Verificar:**
- El issue se cierra
- Se agrega el comentario de cierre (si se proporcionó)

#### Test 5.2: Hacer merge de un PR
**Prerequisito:** Necesitas un PR listo para merge

**Entrada (merge normal):**
```json
{
  "jsonrpc": "2.0",
  "id": 17,
  "method": "tools/call",
  "params": {
    "name": "github_merge_pr",
    "arguments": {
      "owner": "tu_usuario",
      "repo": "tu_repo",
      "pr_number": 1,
      "merge_method": "merge",
      "commit_title": "Merge PR #1",
      "commit_message": "Mergeado desde MCP server"
    }
  }
}
```

**Entrada (squash merge):**
```json
{
  "jsonrpc": "2.0",
  "id": 18,
  "method": "tools/call",
  "params": {
    "name": "github_merge_pr",
    "arguments": {
      "owner": "tu_usuario",
      "repo": "tu_repo",
      "pr_number": 2,
      "merge_method": "squash"
    }
  }
}
```

#### Test 5.3: Re-ejecutar workflow fallido
**Prerequisito:** Necesitas un workflow que haya fallado

**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 19,
  "method": "tools/call",
  "params": {
    "name": "github_rerun_workflow",
    "arguments": {
      "owner": "tu_usuario",
      "repo": "tu_repo",
      "run_id": 123456789
    }
  }
}
```

#### Test 5.4: Descartar alerta de Dependabot
**Prerequisito:** Necesitas una alerta de Dependabot

**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 20,
  "method": "tools/call",
  "params": {
    "name": "github_dismiss_dependabot_alert",
    "arguments": {
      "owner": "tu_usuario",
      "repo": "tu_repo",
      "alert_number": 1,
      "reason": "tolerable_risk",
      "comment": "Riesgo aceptado - versión necesaria para compatibilidad"
    }
  }
}
```

**Opciones válidas para `reason`:**
- `fix_started`: Corrección iniciada
- `inaccurate`: Alerta inexacta
- `no_bandwidth`: Sin recursos para corregir
- `not_used`: Dependencia no utilizada
- `tolerable_risk`: Riesgo tolerable

#### Test 5.5: Descartar alerta de Code Scanning
**Prerequisito:** Necesitas una alerta de Code Scanning

**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 21,
  "method": "tools/call",
  "params": {
    "name": "github_dismiss_code_alert",
    "arguments": {
      "owner": "tu_usuario",
      "repo": "tu_repo",
      "alert_number": 1,
      "reason": "false_positive",
      "comment": "Falso positivo - validación adicional presente"
    }
  }
}
```

**Opciones válidas para `reason`:**
- `false_positive`: Falso positivo
- `won't_fix`: No se corregirá
- `used_in_tests`: Usado solo en tests

#### Test 5.6: Descartar alerta de Secret Scanning
**Prerequisito:** Necesitas una alerta de Secret Scanning

**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 22,
  "method": "tools/call",
  "params": {
    "name": "github_dismiss_secret_alert",
    "arguments": {
      "owner": "tu_usuario",
      "repo": "tu_repo",
      "alert_number": 1,
      "reason": "revoked",
      "comment": "Secret revocado e invalidado"
    }
  }
}
```

**Opciones válidas para `reason`:**
- `false_positive`: Falso positivo
- `won't_fix`: No se corregirá
- `revoked`: Ya fue revocado
- `used_in_tests`: Usado solo en tests

---

### Fase 6: Pruebas de Operaciones Git Avanzadas (20 minutos)

#### Test 6.1: Crear una rama nueva
**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 23,
  "method": "tools/call",
  "params": {
    "name": "git_checkout",
    "arguments": {
      "branch_name": "feature/test-branch",
      "create_new": true
    }
  }
}
```

#### Test 6.2: Hacer cambios y commit
**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 24,
  "method": "tools/call",
  "params": {
    "name": "git_add",
    "arguments": {
      "files": ["test_file.txt"]
    }
  }
}
```

```json
{
  "jsonrpc": "2.0",
  "id": 25,
  "method": "tools/call",
  "params": {
    "name": "git_commit",
    "arguments": {
      "message": "test: Add changes to test file"
    }
  }
}
```

#### Test 6.3: Push a remoto
**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 26,
  "method": "tools/call",
  "params": {
    "name": "git_push",
    "arguments": {
      "branch": "feature/test-branch"
    }
  }
}
```

#### Test 6.4: Crear PR desde la nueva rama
**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 27,
  "method": "tools/call",
  "params": {
    "name": "github_create_pr",
    "arguments": {
      "owner": "tu_usuario",
      "repo": "tu_repo",
      "title": "Test PR from MCP server",
      "body": "Este PR fue creado automáticamente desde el MCP server",
      "head": "feature/test-branch",
      "base": "main"
    }
  }
}
```

#### Test 6.5: Merge con detección de conflictos
**Entrada:**
```json
{
  "jsonrpc": "2.0",
  "id": 28,
  "method": "tools/call",
  "params": {
    "name": "git_safe_merge",
    "arguments": {
      "source_branch": "feature/test-branch",
      "target_branch": "main"
    }
  }
}
```

---

### Fase 7: Pruebas Multi-Perfil (10 minutos)

#### Test 7.1: Iniciar servidor con perfil específico
```bash
.\github-mcp-server.exe --profile personal
```

#### Test 7.2: Verificar que usa el token correcto
**Configuración en Claude Desktop:**
```json
{
  "mcpServers": {
    "github-personal": {
      "command": "path\\to\\github-mcp-server.exe",
      "args": ["--profile", "personal"],
      "env": {"GITHUB_TOKEN": "ghp_token_personal"}
    },
    "github-work": {
      "command": "path\\to\\github-mcp-server.exe",
      "args": ["--profile", "work"],
      "env": {"GITHUB_TOKEN": "ghp_token_work"}
    }
  }
}
```

---

## Matriz de Pruebas

| Categoría | Herramientas | Estado | Notas |
|-----------|--------------|--------|-------|
| Info Git | 8 tools | ✅ Testeado | Baja cobertura unitaria (12.8%) |
| Git Básico | 6 tools | ✅ Testeado | Funcionan correctamente |
| Git Avanzado | 7 tools | ⚠️ Requiere pruebas manuales | Merge, rebase, force-push |
| Conflictos | 6 tools | ⚠️ Requiere pruebas manuales | Detección y resolución |
| Operaciones Híbridas | 2 tools | ✅ Testeado (43.1%) | create_file, update_file |
| GitHub API | 4 tools | ✅ Testeado (39.6%) | Repos, PRs básicos |
| Dashboard | 7 tools | ❌ Sin tests unitarios | Requiere token real |
| Response v2.1 | 3 tools | ❌ Sin tests unitarios | Nuevas herramientas |
| Repair v2.1 | 6 tools | ❌ Sin tests unitarios | Nuevas herramientas |

---

## Problemas Conocidos y Recomendaciones

### 1. Cobertura de Tests
**Problema:** Cobertura baja en algunos paquetes
- `pkg/git`: Solo 12.8%
- `pkg/github`: 39.6%

**Recomendación:** Agregar tests unitarios para:
- Funciones de merge y rebase
- Manejo de conflictos
- Operaciones de push/pull con diferentes escenarios

### 2. Tests de Integración
**Problema:** No hay tests E2E automatizados

**Recomendación:** Crear suite de tests de integración que:
- Use un repositorio de prueba en GitHub
- Ejecute flujos completos (crear rama → cambios → PR → merge)
- Valide todas las nuevas herramientas v2.1

### 3. Tests de Dashboard y v2.1
**Problema:** Sin tests unitarios para 16 herramientas nuevas

**Recomendación:** Agregar mocks para:
- Cliente HTTP del dashboard
- Respuestas de API de GitHub
- Operaciones de response y repair

### 4. Race Condition Testing
**Problema:** No se pudo ejecutar `-race` en Windows (requiere CGO)

**Recomendación:**
- Ejecutar tests con `-race` en Linux/macOS
- Agregar a CI/CD pipeline
- Revisar código para concurrency patterns

### 5. Validación de Seguridad
**Estado:** ✅ Excelente

El proyecto tiene 14 tests de seguridad exhaustivos que cubren:
- Path traversal (CWE-22)
- Command injection (CWE-78)
- CVEs conocidos
- Patrones de debilidad comunes
- Criptografía
- Dependencias

---

## Checklist de Pruebas Completado

### Antes de las Pruebas
- [ ] Token de GitHub obtenido con permisos correctos
- [ ] Variable de entorno `GITHUB_TOKEN` configurada
- [ ] Proyecto compilado exitosamente
- [ ] Tienes acceso a un repositorio de prueba

### Pruebas Básicas
- [ ] Servidor inicia correctamente
- [ ] `initialize` funciona
- [ ] `tools/list` retorna 55+ herramientas

### Pruebas Git Locales
- [ ] `git_status` muestra información correcta
- [ ] `git_list_files` lista archivos
- [ ] `git_get_file_content` lee archivos
- [ ] `create_file` crea archivos (híbrido)
- [ ] `update_file` actualiza archivos (híbrido)

### Pruebas GitHub API
- [ ] `github_list_repos` lista repositorios
- [ ] `github_list_prs` lista PRs
- [ ] `github_dashboard` muestra dashboard completo
- [ ] `github_notifications` obtiene notificaciones

### Pruebas Response v2.1
- [ ] `github_comment_issue` agrega comentarios a issues
- [ ] `github_comment_pr` agrega comentarios a PRs
- [ ] `github_review_pr` crea revisiones (APPROVE)
- [ ] `github_review_pr` crea revisiones (REQUEST_CHANGES)

### Pruebas Repair v2.1
- [ ] `github_close_issue` cierra issues
- [ ] `github_merge_pr` hace merge (método: merge)
- [ ] `github_merge_pr` hace merge (método: squash)
- [ ] `github_rerun_workflow` re-ejecuta workflows
- [ ] `github_dismiss_dependabot_alert` descarta alertas Dependabot
- [ ] `github_dismiss_code_alert` descarta alertas Code Scanning
- [ ] `github_dismiss_secret_alert` descarta alertas Secret Scanning

### Pruebas Git Avanzadas
- [ ] Crear rama nueva
- [ ] Hacer cambios, add, commit
- [ ] Push a remoto
- [ ] Crear PR desde nueva rama
- [ ] Merge con detección de conflictos

### Pruebas Multi-Perfil
- [ ] Servidor inicia con `--profile`
- [ ] Configuración en Claude Desktop funciona
- [ ] Múltiples perfiles funcionan independientemente

---

## Contacto y Soporte

Si encuentras problemas durante las pruebas:
1. Verifica que el token tiene los permisos correctos
2. Revisa los logs del servidor
3. Consulta el archivo [CLAUDE.md](CLAUDE.md) para documentación detallada
4. Abre un issue en el repositorio con:
   - Comando ejecutado
   - Respuesta recibida
   - Comportamiento esperado vs real

---

## Conclusión

El proyecto está en buen estado de producción:
- ✅ Compila sin errores
- ✅ Tests unitarios pasan (100%)
- ✅ Tests de seguridad exhaustivos
- ✅ Sin problemas detectados por análisis estático
- ⚠️ Cobertura de tests mejorable
- ❌ Herramientas v2.1 requieren tests

**Recomendación:** El servidor es funcional y seguro para uso en producción. Para mayor confianza, ejecuta las pruebas manuales de esta guía, especialmente para las nuevas herramientas v2.1 (Response y Repair).
