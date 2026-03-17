package server

import "github.com/jotajotape/github-go-server-mcp/pkg/types"

// ListGitAdvancedTools retorna las herramientas Git avanzadas
func ListGitAdvancedTools() []types.Tool {
	return []types.Tool{
		{
			Name:        "git_history",
			Description: "Consolidated Git history tool. Operations: log (commit history with analysis), diff (modified files with statistics). Use 'log' to view commit history and 'diff' to see file changes.",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation": {Type: "string", Description: "Operation to perform: log, diff"},
					"limit":     {Type: "string", Description: "Number of commits to show (for log, default: 20)"},
					"staged":    {Type: "boolean", Description: "Show staged files (for diff, default: false)"},
				},
				Required: []string{"operation"},
			},
		},
		{
			Name:        "git_branch",
			Description: "Consolidated Git branch management tool. Operations: checkout (switch or create branch), checkout_remote (checkout remote branch with local tracking), list (list all branches), merge (merge branches with safety validations), rebase (rebase onto specified branch), backup (create backup tag of current state).",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation":     {Type: "string", Description: "Operation to perform: checkout, checkout_remote, list, merge, rebase, backup"},
					"branch":        {Type: "string", Description: "Branch name (for checkout, rebase)"},
					"create":        {Type: "boolean", Description: "Create new branch (for checkout)"},
					"remote_branch": {Type: "string", Description: "Remote branch name (for checkout_remote)"},
					"local_branch":  {Type: "string", Description: "Local branch name (for checkout_remote, optional)"},
					"source_branch": {Type: "string", Description: "Source branch for merge (for merge)"},
					"target_branch": {Type: "string", Description: "Target branch for merge (for merge, optional - uses current)"},
					"remote":        {Type: "boolean", Description: "Include remote branches (for list, default: false)"},
					"name":          {Type: "string", Description: "Backup name (for backup)"},
				},
				Required: []string{"operation"},
			},
		},
		{
			Name:        "git_sync",
			Description: "Consolidated Git sync and push/pull tool. Operations: push (push to remote), pull (pull from remote), force_push (force push with --force-with-lease and automatic backup), push_upstream (push setting upstream tracking), sync (fetch + intelligent merge with remote), pull_strategy (pull with specific strategy: merge, rebase, ff-only).",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation":     {Type: "string", Description: "Operation to perform: push, pull, force_push, push_upstream, sync, pull_strategy"},
					"branch":        {Type: "string", Description: "Branch name (optional, uses current branch)"},
					"force":         {Type: "boolean", Description: "Use --force-with-lease (for force_push)"},
					"remote_branch": {Type: "string", Description: "Remote branch name (for sync, optional)"},
					"strategy":      {Type: "string", Description: "Pull strategy: merge, rebase, ff-only (for pull_strategy)"},
				},
				Required: []string{"operation"},
			},
		},
		{
			Name:        "git_conflict",
			Description: "Consolidated Git conflict management tool. Operations: status (detailed conflict state in merge/rebase), resolve (automatic conflict resolution with strategies: theirs, ours, abort, manual), detect (detect potential conflicts between branches before merging), safe_merge (merge with automatic backup and conflict detection).",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation":     {Type: "string", Description: "Operation to perform: status, resolve, detect, safe_merge"},
					"strategy":      {Type: "string", Description: "Resolution strategy: theirs, ours, abort, manual (for resolve)"},
					"source_branch": {Type: "string", Description: "Source branch (for detect)"},
					"target_branch": {Type: "string", Description: "Target branch (for detect)"},
					"source":        {Type: "string", Description: "Source branch (for safe_merge)"},
					"target":        {Type: "string", Description: "Target branch (for safe_merge, optional - uses current)"},
				},
				Required: []string{"operation"},
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
			Name:        "git_reset",
			Description: "Undo commits by moving HEAD to a specific commit (soft/mixed/hard). Dangerous operation - use with caution.",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"mode":   {Type: "string", Description: "Modo: soft (mantiene staging), mixed (deshace staging), hard (descarta todo)"},
					"target": {Type: "string", Description: "Commit/ref destino (ej: HEAD~1, abc123, main)"},
					"files":  {Type: "string", Description: "Archivos específicos a resetear (opcional, separados por comas)"},
				},
				Required: []string{"mode", "target"},
			},
		},
	}
}
