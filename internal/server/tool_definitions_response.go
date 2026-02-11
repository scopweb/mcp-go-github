package server

import "github.com/jotajotape/github-go-server-mcp/pkg/types"

// ListResponseTools retorna las herramientas de respuesta (comentarios, reviews)
func ListResponseTools() []types.Tool {
	return []types.Tool{
		{
			Name:        "github_comment_issue",
			Description: "💬 Agregar un comentario a un issue",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":  {Type: "string", Description: "Propietario del repositorio"},
					"repo":   {Type: "string", Description: "Nombre del repositorio"},
					"number": {Type: "number", Description: "Número del issue"},
					"body":   {Type: "string", Description: "Texto del comentario (soporta Markdown)"},
				},
				Required: []string{"owner", "repo", "number", "body"},
			},
		},
		{
			Name:        "github_comment_pr",
			Description: "💬 Agregar un comentario a un pull request",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":  {Type: "string", Description: "Propietario del repositorio"},
					"repo":   {Type: "string", Description: "Nombre del repositorio"},
					"number": {Type: "number", Description: "Número del PR"},
					"body":   {Type: "string", Description: "Texto del comentario (soporta Markdown)"},
				},
				Required: []string{"owner", "repo", "number", "body"},
			},
		},
		{
			Name:        "github_review_pr",
			Description: "✅ Crear una review en un pull request (aprobar, solicitar cambios o comentar)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":  {Type: "string", Description: "Propietario del repositorio"},
					"repo":   {Type: "string", Description: "Nombre del repositorio"},
					"number": {Type: "number", Description: "Número del PR"},
					"event":  {Type: "string", Description: "Tipo de review: APPROVE, REQUEST_CHANGES, COMMENT"},
					"body":   {Type: "string", Description: "Comentario de la review (requerido para REQUEST_CHANGES y COMMENT)"},
				},
				Required: []string{"owner", "repo", "number", "event"},
			},
		},
	}
}
