# Changelog

Todos los cambios importantes del proyecto GitHub MCP Server serán documentados en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
y este proyecto sigue [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.2.1] - 2024-10-23

### 🔧 Changed
- Añadida declaración `toolchain go1.24.6` para consistencia de builds
- Mejorada reproducibilidad en diferentes entornos de desarrollo

## [2.2.0] - 2024-10-23

### 🚀 Added
- Soporte completo para múltiples perfiles GitHub simultáneos
- Sistema híbrido inteligente que prioriza Git local sobre GitHub API
- Detección automática de contexto Git para optimización de tokens
- Logging mejorado con emojis e información detallada del perfil
- Validación obligatoria de tokens GitHub para mayor seguridad
- Función `NewMCPServer()` para inicialización más robusta

### 🔧 Changed
- **BREAKING:** Perfil ahora es obligatorio con valor por defecto "default"
- **BREAKING:** Token GitHub ahora es obligatorio (no funciona sin token)
- Actualizada versión mínima de Go de 1.19 a 1.24.0
- Mejorada la gestión de errores con validaciones más estrictas
- Optimizada la estructura de inicialización del servidor

### 🔄 Updated
- `golang.org/x/oauth2` de v0.30.0 a v0.32.0
- Versión de Go en go.mod de 1.23.0 a 1.24.0
- Directorio vendor sincronizado con nuevas dependencias
- Documentación actualizada con requisitos del sistema

### 🛡️ Security
- Implementadas mejoras de seguridad sugeridas por GitHub Copilot
- Prevención de inyección de argumentos en comandos Git
- Defensa contra ataques "Path Traversal"
- Validación estricta de todas las entradas del usuario
- Actualización de OAuth2 incluye parches de seguridad

### 🧪 Testing
- Mantenida cobertura de tests al 100%
- Todos los tests pasan después de las actualizaciones
- Verificación de seguridad con `govulncheck` - sin vulnerabilidades
- Tests unitarios completos para todas las funciones críticas

### 🎨 Code Quality
- Formateo automático aplicado a todos los archivos
- Análisis estático limpio con `go vet`
- Código completamente formateado siguiendo estándares de Go
- Eliminadas inconsistencias de formateo

### 📚 Documentation
- README.md completamente reescrito con emojis y mejor estructura
- Tabla de herramientas disponibles con estado de testing
- Instrucciones detalladas para configuración multi-perfil
- Sección de troubleshooting expandida
- Documentación de permisos GitHub requeridos

## [2.1.0] - 2024-10-20

### 🚀 Added
- Sistema de herramientas híbridas (Git local + GitHub API)
- Operaciones Git avanzadas (merge, rebase, stash, etc.)
- Gestión completa de ramas remotas
- Sistema de backups automáticos
- Detección preventiva de conflictos

### 🔧 Changed
- Arquitectura modular mejorada
- Mejor manejo de errores en operaciones Git

### 🔄 Updated
- `github.com/google/go-github` a v74.0.0
- Todas las dependencias a versiones estables

### 🧪 Testing
- Suite completa de tests unitarios
- Cobertura del 100% en funciones críticas

## [2.0.0] - 2024-10-15

### 🚀 Added
- Protocolo JSON-RPC 2.0 completo
- Integración GitHub API
- Operaciones Git locales básicas
- Sistema MCP (Model Context Protocol)

### 🔧 Changed
- Reescritura completa en Go
- Arquitectura modular

### 🛡️ Security
- Autenticación OAuth2 con GitHub
- Validación de tokens

## [1.0.0] - 2024-10-01

### 🚀 Added
- Versión inicial del proyecto
- Funcionalidades básicas de GitHub

---

## Tipos de Cambios

- `🚀 Added` para nuevas funcionalidades
- `🔧 Changed` para cambios en funcionalidades existentes
- `🗑️ Deprecated` para funcionalidades que serán removidas
- `❌ Removed` para funcionalidades removidas
- `🐛 Fixed` para corrección de bugs
- `🛡️ Security` para mejoras de seguridad
- `🔄 Updated` para actualizaciones de dependencias
- `🧪 Testing` para cambios relacionados con tests
- `🎨 Code Quality` para mejoras de calidad de código
- `📚 Documentation` para cambios en documentación

## Links de Comparación

[Unreleased]: https://github.com/scopweb/mcp-go-github/compare/v2.2.0...HEAD
[2.2.0]: https://github.com/scopweb/mcp-go-github/compare/v2.1.0...v2.2.0
[2.1.0]: https://github.com/scopweb/mcp-go-github/compare/v2.0.0...v2.1.0
[2.0.0]: https://github.com/scopweb/mcp-go-github/compare/v1.0.0...v2.0.0
[1.0.0]: https://github.com/scopweb/mcp-go-github/releases/tag/v1.0.0