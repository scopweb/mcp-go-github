package server

import "github.com/jotajotape/github-go-server-mcp/pkg/types"

// ListGitInfoTools retorna las herramientas de información Git
func ListGitInfoTools() []types.Tool {
	return []types.Tool{
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
	}
}
