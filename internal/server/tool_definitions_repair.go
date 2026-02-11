package server

import "github.com/jotajotape/github-go-server-mcp/pkg/types"

// ListRepairTools retorna las herramientas de reparación (cerrar issues, merge, rerun workflows, dismiss alerts)
func ListRepairTools() []types.Tool {
	return []types.Tool{
		{
			Name:        "github_close_issue",
			Description: "🔒 Cerrar un issue con un comentario opcional",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":   {Type: "string", Description: "Propietario del repositorio"},
					"repo":    {Type: "string", Description: "Nombre del repositorio"},
					"number":  {Type: "number", Description: "Número del issue"},
					"comment": {Type: "string", Description: "Comentario de cierre opcional"},
				},
				Required: []string{"owner", "repo", "number"},
			},
		},
		{
			Name:        "github_merge_pr",
			Description: "🔀 Mergear un pull request",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":          {Type: "string", Description: "Propietario del repositorio"},
					"repo":           {Type: "string", Description: "Nombre del repositorio"},
					"number":         {Type: "number", Description: "Número del PR"},
					"commit_message": {Type: "string", Description: "Mensaje de commit de merge opcional"},
					"merge_method":   {Type: "string", Description: "Método de merge: merge, squash, rebase (default: merge)"},
				},
				Required: []string{"owner", "repo", "number"},
			},
		},
		{
			Name:        "github_rerun_workflow",
			Description: "🔄 Re-ejecutar un workflow fallido de GitHub Actions",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":            {Type: "string", Description: "Propietario del repositorio"},
					"repo":             {Type: "string", Description: "Nombre del repositorio"},
					"run_id":           {Type: "number", Description: "ID del workflow run"},
					"failed_jobs_only": {Type: "boolean", Description: "Re-ejecutar solo jobs fallidos (default: false)"},
				},
				Required: []string{"owner", "repo", "run_id"},
			},
		},
		{
			Name:        "github_dismiss_dependabot_alert",
			Description: "🛡️ Dismissar una alerta de seguridad de Dependabot",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":   {Type: "string", Description: "Propietario del repositorio"},
					"repo":    {Type: "string", Description: "Nombre del repositorio"},
					"number":  {Type: "number", Description: "Número de la alerta"},
					"reason":  {Type: "string", Description: "Razón: fix_started, inaccurate, no_bandwidth, not_used, tolerable_risk"},
					"comment": {Type: "string", Description: "Comentario explicando el dismissal (opcional)"},
				},
				Required: []string{"owner", "repo", "number", "reason"},
			},
		},
		{
			Name:        "github_dismiss_code_alert",
			Description: "🔍 Dismissar una alerta de code scanning",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":   {Type: "string", Description: "Propietario del repositorio"},
					"repo":    {Type: "string", Description: "Nombre del repositorio"},
					"number":  {Type: "number", Description: "Número de la alerta"},
					"reason":  {Type: "string", Description: "Razón: false positive, won't fix, used in tests"},
					"comment": {Type: "string", Description: "Comentario explicando el dismissal (opcional)"},
				},
				Required: []string{"owner", "repo", "number", "reason"},
			},
		},
		{
			Name:        "github_dismiss_secret_alert",
			Description: "🔑 Dismissar una alerta de secret scanning",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":      {Type: "string", Description: "Propietario del repositorio"},
					"repo":       {Type: "string", Description: "Nombre del repositorio"},
					"number":     {Type: "number", Description: "Número de la alerta"},
					"resolution": {Type: "string", Description: "Resolución: false_positive, wont_fix, revoked, used_in_tests"},
				},
				Required: []string{"owner", "repo", "number", "resolution"},
			},
		},
	}
}
