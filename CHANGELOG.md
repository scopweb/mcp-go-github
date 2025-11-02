# Changelog

Todos los cambios importantes del proyecto GitHub MCP Server serán documentados en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
y este proyecto sigue [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.4.0] - 2025-01-02

### 🎨 Code Quality
- **PHASE 3 COMPLETE:** Implementación completa de linting profesional con golangci-lint
- Resueltos 50+ issues de código identificados por múltiples linters (errcheck, revive, staticcheck, misspell, gocritic, gosimple, gosec)
- Convertidas cadenas if-else complejas a declaraciones switch para mejor legibilidad
- Corregidos errores de ortografía en español a inglés en strings de usuario y comentarios
- Actualizadas funciones deprecated de GitHub API (github.String/Bool → github.Ptr)
- Eliminadas llamadas innecesarias a fmt.Sprintf para strings literales
- Marcados parámetros no utilizados como `_` en funciones de test mock
- Resuelto issue de seguridad G204 eliminando ejecución dinámica de comandos en tests
- **CLEAN LINTING:** golangci-lint ejecuta sin errores ni warnings
- Código preparado para estándares profesionales de desarrollo Go

### 🔧 Technical Improvements
- Mejorada robustez del manejo de errores con validaciones apropiadas de os.Chdir
- Optimizada estructura de control de flujo en funciones de parsing de conflictos
- Eliminadas dependencias innecesarias en expresiones de formato
- Mejorada mantenibilidad del código siguiendo mejores prácticas de Go

### 🧪 Testing
- Tests de linting pasan completamente sin issues
- Validación de calidad de código automatizada con CI-ready configuration
- Preparación para integración continua con estándares profesionales

### 📚 Documentation
- CHANGELOG actualizado con completación de Phase 3
- Documentación de mejoras de calidad de código

## [2.3.0] - 2025-11-02

### 🚀 Added
- Reestructuración completa del proyecto siguiendo mejores prácticas de Go
- Nuevo directorio `pkg/` para código reutilizable y bibliotecas compartidas
- Nuevo directorio `cmd/github-mcp-server/` para punto de entrada de la aplicación
- Movidos paquetes `interfaces`, `types`, `github`, `git` a `pkg/` para mejor organización

### 🔧 Changed
- **BREAKING:** Reorganización de estructura de directorios para alinearse con estándares Go
- Actualización de rutas de importación en todo el proyecto
- Mejor separación entre código interno (`internal/`) y público (`pkg/`)

### 🔄 Updated
- `github.com/google/go-github` de v74.0.0 a v76.0.0 (últimas características y correcciones)
- Sincronización completa del directorio vendor con nuevas dependencias

### 🧪 Testing
- Corregidos todos los tests unitarios que estaban fallando
- Completada implementación de mocks para interfaces `GitOperations`
- Actualizados mocks de comandos Git en tests de integración
- Todos los tests pasan exitosamente (100% funcionalidad validada)
- Tests de seguridad pasan sin issues críticos

### 🎨 Code Quality
- Estructura del proyecto completamente reestructurada
- Mejor organización modular del código
- Eliminadas inconsistencias en tests y mocks
- Código preparado para futuras expansiones siguiendo patrones estándar de Go

### 📚 Documentation
- CHANGELOG actualizado con cambios recientes
- Documentación de estructura del proyecto actualizada

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

[Unreleased]: https://github.com/scopweb/mcp-go-github/compare/v2.4.0...HEAD
[2.4.0]: https://github.com/scopweb/mcp-go-github/compare/v2.3.0...v2.4.0
[2.3.0]: https://github.com/scopweb/mcp-go-github/compare/v2.2.1...v2.3.0
[2.2.1]: https://github.com/scopweb/mcp-go-github/compare/v2.2.0...v2.2.1
[2.2.0]: https://github.com/scopweb/mcp-go-github/compare/v2.1.0...v2.2.0
[2.1.0]: https://github.com/scopweb/mcp-go-github/compare/v2.0.0...v2.1.0
[2.0.0]: https://github.com/scopweb/mcp-go-github/compare/v1.0.0...v2.0.0
[1.0.0]: https://github.com/scopweb/mcp-go-github/releases/tag/v1.0.0