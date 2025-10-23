package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/scopweb/mcp-go-github/internal/hybrid"
	"github.com/scopweb/mcp-go-github/internal/interfaces"
	"github.com/scopweb/mcp-go-github/internal/types"
)

// MCPServer representa el servidor MCP principal
type MCPServer struct {
	GithubClient interfaces.GitHubOperations
	GitClient    interfaces.GitOperations
}

// HandleRequest procesa las peticiones JSON-RPC del protocolo MCP
func HandleRequest(s *MCPServer, req types.JSONRPCRequest) types.JSONRPCResponse {
	id := req.ID
	if id == nil {
		id = 0
	}

	response := types.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
	}

	if req.JSONRPC != "2.0" {
		response.Error = &types.JSONRPCError{
			Code:    -32600,
			Message: "Invalid Request: jsonrpc must be '2.0'",
		}
		return response
	}

	if req.Method == "" {
		response.Error = &types.JSONRPCError{
			Code:    -32600,
			Message: "Invalid Request: method is required",
		}
		return response
	}

	switch req.Method {
	case "initialize":
		response.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "github-mcp-local-hybrid",
				"version": "2.0.0",
			},
		}
	case "initialized":
		response.Result = map[string]interface{}{}
	case "tools/list":
		response.Result = ListTools()
	case "tools/call":
		result, err := CallTool(s, req.Params)
		if err != nil {
			response.Error = &types.JSONRPCError{
				Code:    -32603,
				Message: err.Error(),
			}
		} else {
			response.Result = result
		}
	default:
		response.Error = &types.JSONRPCError{
			Code:    -32601,
			Message: "Method not found",
		}
	}

	return response
}

// ListTools retorna la lista de herramientas disponibles
func ListTools() types.ToolsListResult {
	tools := []types.Tool{
		// Herramientas de información
		{
			Name:        "git_status",
			Description: "Muestra el estado del repositorio Git local y configuración",
			InputSchema: types.ToolInputSchema{
				Type:       "object",
				Properties: map[string]types.Property{},
			},
		},
		{
			Name:        "git_set_workspace",
			Description: "🔧 Configura el directorio de trabajo para operaciones Git",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"path": {Type: "string", Description: "Ruta del directorio del repositorio Git"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "git_get_file_sha",
			Description: "🔑 Obtiene el SHA de un archivo específico desde Git",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"path": {Type: "string", Description: "Ruta del archivo"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "git_get_last_commit",
			Description: "🔑 Obtiene el SHA del último commit",
			InputSchema: types.ToolInputSchema{
				Type:       "object",
				Properties: map[string]types.Property{},
			},
		},
		{
			Name:        "git_get_file_content",
			Description: "📄 Obtiene el contenido de un archivo desde Git",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"path": {Type: "string", Description: "Ruta del archivo"},
					"ref":  {Type: "string", Description: "Referencia Git (branch, commit, tag). Default: HEAD"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "git_get_changed_files",
			Description: "📋 Lista archivos modificados en working directory o staging area",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"staged": {Type: "boolean", Description: "Mostrar archivos en staging (true) o working directory (false)"},
				},
			},
		},
		{
			Name:        "git_validate_repo",
			Description: "✅ Valida si un directorio es un repositorio Git válido",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"path": {Type: "string", Description: "Ruta del directorio a validar"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "git_list_files",
			Description: "📄 Lista todos los archivos en el repositorio Git",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"ref": {Type: "string", Description: "Referencia Git (branch, commit, tag). Default: HEAD"},
				},
			},
		},

		// Herramientas Git locales básicas
		{
			Name:        "git_add",
			Description: "Agrega archivos al staging area (requiere Git local)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"files": {Type: "string", Description: "Archivos a agregar (. para todos)"},
				},
				Required: []string{"files"},
			},
		},
		{
			Name:        "git_commit",
			Description: "Hace commit de los cambios en staging (requiere Git local)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"message": {Type: "string", Description: "Mensaje del commit"},
				},
				Required: []string{"message"},
			},
		},
		{
			Name:        "git_push",
			Description: "Sube cambios al repositorio remoto (requiere Git local)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"branch": {Type: "string", Description: "Rama a subir (opcional, usa actual)"},
				},
			},
		},
		{
			Name:        "git_pull",
			Description: "Baja cambios del repositorio remoto (requiere Git local)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"branch": {Type: "string", Description: "Rama a bajar (opcional, usa actual)"},
				},
			},
		},
		{
			Name:        "git_checkout",
			Description: "Cambia de rama o crea nueva rama (requiere Git local)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"branch": {Type: "string", Description: "Nombre de la rama"},
					"create": {Type: "boolean", Description: "Crear nueva rama"},
				},
				Required: []string{"branch"},
			},
		},

		// Herramientas Git avanzadas
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

		// Advanced Git Operations
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
			Description: "🛡️ Merge seguro con backup automático y detección de conflictos",
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
			Description: "⚠️ Estado detallado de conflictos en merge/rebase",
			InputSchema: types.ToolInputSchema{
				Type:       "object",
				Properties: map[string]types.Property{},
			},
		},
		{
			Name:        "git_resolve_conflicts",
			Description: "🔧 Resolución automática de conflictos con estrategias",
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
			Description: "🔍 Detecta conflictos potenciales entre ramas",
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

		// Herramientas híbridas
		{
			Name:        "create_file",
			Description: "✅ Crea archivo PRIORIZANDO Git local (0 tokens) sobre GitHub API",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"path":    {Type: "string", Description: "Ruta del archivo"},
					"content": {Type: "string", Description: "Contenido del archivo"},
					"message": {Type: "string", Description: "Mensaje del commit (opcional para Git local)"},
					"owner":   {Type: "string", Description: "Propietario (SOLO si falla Git local)"},
					"repo":    {Type: "string", Description: "Repositorio (SOLO si falla Git local)"},
				},
				Required: []string{"path", "content"},
			},
		},
		{
			Name:        "update_file",
			Description: "✅ Actualiza archivo PRIORIZANDO Git local (0 tokens) sobre GitHub API",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"path":    {Type: "string", Description: "Ruta del archivo"},
					"content": {Type: "string", Description: "Nuevo contenido"},
					"message": {Type: "string", Description: "Mensaje del commit (opcional para Git local)"},
					"owner":   {Type: "string", Description: "Propietario (SOLO si falla Git local)"},
					"repo":    {Type: "string", Description: "Repositorio (SOLO si falla Git local)"},
					"sha":     {Type: "string", Description: "SHA del archivo (SOLO si falla Git local)"},
				},
				Required: []string{"path", "content"},
			},
		},

		// Herramientas API puras
		{
			Name:        "github_list_repos",
			Description: "Lista repositorios del usuario (GitHub API)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"type": {Type: "string", Description: "Tipo: all, owner, member"},
				},
			},
		},
		{
			Name:        "github_create_repo",
			Description: "Crea un nuevo repositorio (GitHub API)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"name":        {Type: "string", Description: "Nombre del repositorio"},
					"description": {Type: "string", Description: "Descripción del repositorio"},
					"private":     {Type: "boolean", Description: "Repositorio privado"},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "github_list_prs",
			Description: "Lista pull requests (GitHub API)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner": {Type: "string", Description: "Propietario del repositorio"},
					"repo":  {Type: "string", Description: "Nombre del repositorio"},
					"state": {Type: "string", Description: "Estado: open, closed, all"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_create_pr",
			Description: "Crea pull request (GitHub API)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner": {Type: "string", Description: "Propietario del repositorio"},
					"repo":  {Type: "string", Description: "Nombre del repositorio"},
					"title": {Type: "string", Description: "Título del PR"},
					"body":  {Type: "string", Description: "Descripción del PR"},
					"head":  {Type: "string", Description: "Rama origen"},
					"base":  {Type: "string", Description: "Rama destino"},
				},
				Required: []string{"owner", "repo", "title", "head", "base"},
			},
		},
	}

	return types.ToolsListResult{Tools: tools}
}

// CallTool ejecuta la herramienta solicitada
func CallTool(s *MCPServer, params map[string]interface{}) (types.ToolCallResult, error) {
	name, ok := params["name"].(string)
	if !ok {
		return types.ToolCallResult{}, fmt.Errorf("tool name required")
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	ctx := context.Background()
	var text string
	var err error

	switch name {
	// Herramientas Git básicas
	case "git_status":
		text, err = s.GitClient.Status()
	case "git_set_workspace":
		path, _ := arguments["path"].(string)
		text, err = s.GitClient.SetWorkspace(path)
	case "git_get_file_sha":
		path, _ := arguments["path"].(string)
		text, err = s.GitClient.GetFileSHA(path)
	case "git_get_last_commit":
		text, err = s.GitClient.GetLastCommit()
	case "git_get_file_content":
		path, _ := arguments["path"].(string)
		ref, _ := arguments["ref"].(string)
		text, err = s.GitClient.GetFileContent(path, ref)
	case "git_get_changed_files":
		staged, _ := arguments["staged"].(bool)
		text, err = s.GitClient.GetChangedFiles(staged)
	case "git_validate_repo":
		path, _ := arguments["path"].(string)
		text, err = s.GitClient.ValidateRepo(path)
	case "git_list_files":
		ref, _ := arguments["ref"].(string)
		text, err = s.GitClient.ListFiles(ref)
	case "git_add":
		files, _ := arguments["files"].(string)
		text, err = s.GitClient.Add(files)
	case "git_commit":
		message, _ := arguments["message"].(string)
		text, err = s.GitClient.Commit(message)
	case "git_push":
		branch, _ := arguments["branch"].(string)
		text, err = s.GitClient.Push(branch)
	case "git_pull":
		branch, _ := arguments["branch"].(string)
		text, err = s.GitClient.Pull(branch)
	case "git_checkout":
		branch, _ := arguments["branch"].(string)
		create, _ := arguments["create"].(bool)
		text, err = s.GitClient.Checkout(branch, create)

	// Herramientas Git avanzadas
	case "git_log_analysis":
		limit, _ := arguments["limit"].(string)
		text, err = s.GitClient.LogAnalysis(limit)
	case "git_diff_files":
		staged, _ := arguments["staged"].(bool)
		text, err = s.GitClient.DiffFiles(staged)
	case "git_branch_list":
		remote, _ := arguments["remote"].(bool)
		branches, branchErr := s.GitClient.BranchList(remote)
		if branchErr != nil {
			err = branchErr
		} else {
			// Convertir a JSON para una salida más estructurada
			jsonOutput, jsonErr := json.MarshalIndent(branches, "", "  ")
			if jsonErr != nil {
				err = fmt.Errorf("failed to marshal branch list: %w", jsonErr)
			} else {
				text = string(jsonOutput)
			}
		}
	case "git_stash":
		operation, _ := arguments["operation"].(string)
		name, _ := arguments["name"].(string)
		text, err = s.GitClient.Stash(operation, name)
	case "git_remote":
		operation, _ := arguments["operation"].(string)
		name, _ := arguments["name"].(string)
		url, _ := arguments["url"].(string)
		text, err = s.GitClient.Remote(operation, name, url)
	case "git_tag":
		operation, _ := arguments["operation"].(string)
		tagName, _ := arguments["tag_name"].(string)
		message, _ := arguments["message"].(string)
		text, err = s.GitClient.Tag(operation, tagName, message)
	case "git_clean":
		operation, _ := arguments["operation"].(string)
		dryRun, exists := arguments["dry_run"].(bool)
		if !exists {
			dryRun = true // default a true para seguridad
		}
		text, err = s.GitClient.Clean(operation, dryRun)

	case "git_context":
		text = hybrid.AutoDetectContext(s.GitClient)
		err = nil

	// Advanced Git Operations
	case "git_checkout_remote":
		remoteBranch, _ := arguments["remote_branch"].(string)
		localBranch, _ := arguments["local_branch"].(string)
		text, err = s.GitClient.CheckoutRemote(remoteBranch, localBranch)
	case "git_merge":
		sourceBranch, _ := arguments["source_branch"].(string)
		targetBranch, _ := arguments["target_branch"].(string)
		text, err = s.GitClient.Merge(sourceBranch, targetBranch)
	case "git_rebase":
		branch, _ := arguments["branch"].(string)
		text, err = s.GitClient.Rebase(branch)
	case "git_pull_with_strategy":
		branch, _ := arguments["branch"].(string)
		strategy, _ := arguments["strategy"].(string)
		text, err = s.GitClient.PullWithStrategy(branch, strategy)
	case "git_force_push":
		branch, _ := arguments["branch"].(string)
		force, _ := arguments["force"].(bool)
		text, err = s.GitClient.ForcePush(branch, force)
	case "git_push_upstream":
		branch, _ := arguments["branch"].(string)
		text, err = s.GitClient.PushUpstream(branch)
	case "git_sync_with_remote":
		remoteBranch, _ := arguments["remote_branch"].(string)
		text, err = s.GitClient.SyncWithRemote(remoteBranch)
	case "git_safe_merge":
		source, _ := arguments["source"].(string)
		target, _ := arguments["target"].(string)
		text, err = s.GitClient.SafeMerge(source, target)
	case "git_conflict_status":
		text, err = s.GitClient.ConflictStatus()
	case "git_resolve_conflicts":
		strategy, _ := arguments["strategy"].(string)
		text, err = s.GitClient.ResolveConflicts(strategy)
	case "git_validate_clean_state":
		clean, validateErr := s.GitClient.ValidateCleanState()
		if validateErr != nil {
			err = validateErr
		} else {
			if clean {
				text = "✅ Working directory is clean"
			} else {
				text = "⚠️ Working directory has uncommitted changes"
			}
		}
	case "git_detect_conflicts":
		sourceBranch, _ := arguments["source_branch"].(string)
		targetBranch, _ := arguments["target_branch"].(string)
		conflictInfo, detectErr := s.GitClient.DetectPotentialConflicts(sourceBranch, targetBranch)
		if detectErr != nil {
			err = detectErr
		} else {
			if conflictInfo == "" {
				text = "✅ No potential conflicts detected between branches"
			} else {
				text = fmt.Sprintf("⚠️ %s", conflictInfo)
			}
		}
	case "git_create_backup":
		name, _ := arguments["name"].(string)
		text, err = s.GitClient.CreateBackup(name)

	// Herramientas híbridas
	case "create_file":
		text, err = hybrid.SmartCreateFile(s.GitClient, s.GithubClient, arguments)
	case "update_file":
		text, err = hybrid.SmartUpdateFile(s.GitClient, s.GithubClient, arguments)

	// Herramientas API puras
	case "github_list_repos":
		listType, _ := arguments["type"].(string)
		repos, listErr := s.GithubClient.ListRepositories(ctx, listType)
		if listErr != nil {
			err = listErr
		} else {
			var repoNames []string
			for _, repo := range repos {
				repoNames = append(repoNames, repo.GetFullName())
			}
			text = fmt.Sprintf("Repositories:\n%s", strings.Join(repoNames, "\n"))
		}
	case "github_create_repo":
		name, _ := arguments["name"].(string)
		description, _ := arguments["description"].(string)
		private, _ := arguments["private"].(bool)
		repo, createErr := s.GithubClient.CreateRepository(ctx, name, description, private)
		if createErr != nil {
			err = createErr
		} else {
			text = fmt.Sprintf("Successfully created repository: %s", repo.GetFullName())
		}
	case "github_list_prs":
		owner, _ := arguments["owner"].(string)
		repo, _ := arguments["repo"].(string)
		state, _ := arguments["state"].(string)
		prs, listErr := s.GithubClient.ListPullRequests(ctx, owner, repo, state)
		if listErr != nil {
			err = listErr
		} else {
			var prInfo []string
			for _, pr := range prs {
				prInfo = append(prInfo, fmt.Sprintf("#%d: %s", pr.GetNumber(), pr.GetTitle()))
			}
			if len(prInfo) == 0 {
				text = "No pull requests found."
			} else {
				text = fmt.Sprintf("Pull Requests:\n%s", strings.Join(prInfo, "\n"))
			}
		}
	case "github_create_pr":
		owner, _ := arguments["owner"].(string)
		repo, _ := arguments["repo"].(string)
		title, _ := arguments["title"].(string)
		body, _ := arguments["body"].(string)
		head, _ := arguments["head"].(string)
		base, _ := arguments["base"].(string)
		pr, createErr := s.GithubClient.CreatePullRequest(ctx, owner, repo, title, body, head, base)
		if createErr != nil {
			err = createErr
		} else {
			text = fmt.Sprintf("Successfully created pull request #%d: %s", pr.GetNumber(), pr.GetHTMLURL())
		}
	default:
		return types.ToolCallResult{}, fmt.Errorf("tool not found")
	}

	if err != nil {
		return types.ToolCallResult{}, err
	}

	return types.ToolCallResult{
		Content: []types.Content{{Type: "text", Text: text}},
	}, nil
}
