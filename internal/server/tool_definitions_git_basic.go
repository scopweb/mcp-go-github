package server

import "github.com/jotajotape/github-go-server-mcp/pkg/types"

// ListGitBasicTools retorna las herramientas Git locales básicas
func ListGitBasicTools() []types.Tool {
	return []types.Tool{
		{
			Name:        "git_init",
			Description: "Inicializa un nuevo repositorio Git en el directorio especificado",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"path":           {Type: "string", Description: "Ruta del directorio donde inicializar el repo (debe existir)"},
					"initial_branch": {Type: "string", Description: "Nombre de la rama inicial (defecto: main)"},
				},
				Required: []string{"path"},
			},
			Annotations: IdempotentAnnotation(),
		},
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
	}
}
