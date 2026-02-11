package server

import "github.com/jotajotape/github-go-server-mcp/pkg/types"

// ListHybridTools retorna las herramientas híbridas (Git local primero, GitHub API fallback)
func ListHybridTools() []types.Tool {
	return []types.Tool{
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
	}
}
