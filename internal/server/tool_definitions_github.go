package server

import "github.com/jotajotape/github-go-server-mcp/pkg/types"

// ListGitHubAPITools retorna las herramientas de GitHub API puras
func ListGitHubAPITools() []types.Tool {
	return []types.Tool{
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
}
