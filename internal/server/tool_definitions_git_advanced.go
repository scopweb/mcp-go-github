package server

import "github.com/jotajotape/github-go-server-mcp/pkg/types"

// ListGitAdvancedTools retorna las herramientas Git avanzadas
func ListGitAdvancedTools() []types.Tool {
	return []types.Tool{
		{
			Name:        "git_log_analysis",
			Description: "Análisis completo del historial de commits",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"limit": {Type: "string", Description: "Número de commits a mostrar (default: 20)"},
				},
			},
		},
		{
			Name:        "git_diff_files",
			Description: "Muestra archivos modificados con estadísticas",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"staged": {Type: "boolean", Description: "Mostrar archivos en staging (default: false)"},
				},
			},
		},
		{
			Name:        "git_branch_list",
			Description: "Lista todas las ramas con información detallada",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"remote": {Type: "boolean", Description: "Incluir ramas remotas (default: false)"},
				},
			},
		},
		{
			Name:        "git_stash",
			Description: "Operaciones de stash (guardar cambios temporalmente)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation": {Type: "string", Description: "Operación: list, push, pop, apply, drop, clear"},
					"name":      {Type: "string", Description: "Nombre del stash (opcional)"},
				},
				Required: []string{"operation"},
			},
		},
		{
			Name:        "git_remote",
			Description: "Gestión de repositorios remotos",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation": {Type: "string", Description: "Operación: list, add, remove, show, fetch"},
					"name":      {Type: "string", Description: "Nombre del remoto"},
					"url":       {Type: "string", Description: "URL del remoto (para add)"},
				},
				Required: []string{"operation"},
			},
		},
		{
			Name:        "git_tag",
			Description: "Gestión de tags/etiquetas",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation": {Type: "string", Description: "Operación: list, create, delete, push, show"},
					"tag_name":  {Type: "string", Description: "Nombre del tag"},
					"message":   {Type: "string", Description: "Mensaje del tag (para create)"},
				},
				Required: []string{"operation"},
			},
		},
		{
			Name:        "git_clean",
			Description: "Limpieza de archivos sin seguimiento",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation": {Type: "string", Description: "Tipo: untracked, untracked_dirs, ignored, all"},
					"dry_run":   {Type: "boolean", Description: "Vista previa sin ejecutar (default: true)"},
				},
				Required: []string{"operation"},
			},
		},
		{
			Name:        "git_context",
			Description: "🔧 Auto-detecta contexto Git para optimizar tokens (Git local vs GitHub API)",
			InputSchema: types.ToolInputSchema{
				Type:       "object",
				Properties: map[string]types.Property{},
			},
		},
		{
			Name:        "git_checkout_remote",
			Description: "🚀 Hace checkout de una rama remota creando tracking local",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"remote_branch": {Type: "string", Description: "Nombre de la rama remota (ej: main, develop)"},
					"local_branch":  {Type: "string", Description: "Nombre local (opcional, usa el mismo de la remota)"},
				},
				Required: []string{"remote_branch"},
			},
		},
		{
			Name:        "git_merge",
			Description: "🔀 Merge de ramas con validaciones de seguridad",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"source_branch": {Type: "string", Description: "Rama origen del merge"},
					"target_branch": {Type: "string", Description: "Rama destino (opcional, usa actual)"},
				},
				Required: []string{"source_branch"},
			},
		},
		{
			Name:        "git_rebase",
			Description: "⚡ Rebase con rama especificada",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"branch": {Type: "string", Description: "Rama base para el rebase"},
				},
				Required: []string{"branch"},
			},
		},
		{
			Name:        "git_pull_with_strategy",
			Description: "⬇️ Pull avanzado con estrategias específicas",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"branch":   {Type: "string", Description: "Rama a actualizar (opcional, usa actual)"},
					"strategy": {Type: "string", Description: "Estrategia: merge, rebase, ff-only"},
				},
				Required: []string{"strategy"},
			},
		},
		{
			Name:        "git_force_push",
			Description: "⬆️ Push con opción force (con backup automático)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"branch": {Type: "string", Description: "Rama a subir (opcional, usa actual)"},
					"force":  {Type: "boolean", Description: "Usar --force-with-lease"},
				},
				Required: []string{"force"},
			},
		},
		{
			Name:        "git_push_upstream",
			Description: "⬆️ Push configurando upstream tracking",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"branch": {Type: "string", Description: "Rama a subir (opcional, usa actual)"},
				},
			},
		},
		{
			Name:        "git_sync_with_remote",
			Description: "🔄 Sincronización automática con rama remota (fetch + merge inteligente)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"remote_branch": {Type: "string", Description: "Rama remota (opcional, usa actual)"},
				},
			},
		},
		{
			Name:        "git_safe_merge",
			Description: "🛡️ Merge seguro con backup automático y detección de conflicts",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"source": {Type: "string", Description: "Rama origen"},
					"target": {Type: "string", Description: "Rama destino (opcional, usa actual)"},
				},
				Required: []string{"source"},
			},
		},
		{
			Name:        "git_conflict_status",
			Description: "⚠️ Estado detallado de conflicts en merge/rebase",
			InputSchema: types.ToolInputSchema{
				Type:       "object",
				Properties: map[string]types.Property{},
			},
		},
		{
			Name:        "git_resolve_conflicts",
			Description: "🔧 Resolución automática de conflicts con estrategias",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"strategy": {Type: "string", Description: "Estrategia: theirs, ours, abort, manual"},
				},
				Required: []string{"strategy"},
			},
		},
		{
			Name:        "git_validate_clean_state",
			Description: "✅ Valida que el working directory esté limpio",
			InputSchema: types.ToolInputSchema{
				Type:       "object",
				Properties: map[string]types.Property{},
			},
		},
		{
			Name:        "git_detect_conflicts",
			Description: "🔍 Detecta conflicts potenciales entre ramas",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"source_branch": {Type: "string", Description: "Rama origen"},
					"target_branch": {Type: "string", Description: "Rama destino"},
				},
				Required: []string{"source_branch", "target_branch"},
			},
		},
		{
			Name:        "git_create_backup",
			Description: "💾 Crea backup/tag del estado actual",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"name": {Type: "string", Description: "Nombre del backup"},
				},
				Required: []string{"name"},
			},
		},
	}
}
