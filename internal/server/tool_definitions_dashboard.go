package server

import "github.com/jotajotape/github-go-server-mcp/pkg/types"

// ListDashboardTools retorna las herramientas de GitHub Dashboard
func ListDashboardTools() []types.Tool {
	return []types.Tool{
		{
			Name:        "github_dashboard",
			Description: "📊 Dashboard completo: notificaciones, issues asignadas, PRs pendientes, alertas de seguridad, workflows fallidos",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner": {Type: "string", Description: "Propietario del repositorio (opcional para alertas de seguridad)"},
					"repo":  {Type: "string", Description: "Nombre del repositorio (opcional para alertas de seguridad)"},
				},
			},
		},
		{
			Name:        "github_notifications",
			Description: "🔔 Lista notificaciones pendientes de GitHub",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"all":           {Type: "boolean", Description: "Incluir notificaciones leídas"},
					"participating": {Type: "boolean", Description: "Solo notificaciones donde participas"},
				},
			},
		},
		{
			Name:        "github_assigned_issues",
			Description: "📋 Issues asignadas a ti pendientes de resolver",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"state": {Type: "string", Description: "Estado: open, closed, all (default: open)"},
				},
			},
		},
		{
			Name:        "github_prs_to_review",
			Description: "👀 Pull Requests pendientes de tu revisión",
			InputSchema: types.ToolInputSchema{
				Type:       "object",
				Properties: map[string]types.Property{},
			},
		},
		{
			Name:        "github_security_alerts",
			Description: "🛡️ Alertas de seguridad: Dependabot, Secret Scanning, Code Scanning",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner": {Type: "string", Description: "Propietario del repositorio"},
					"repo":  {Type: "string", Description: "Nombre del repositorio"},
					"type":  {Type: "string", Description: "Tipo: dependabot, secret, code, all (default: all)"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_failed_workflows",
			Description: "❌ Workflows de GitHub Actions fallidos recientemente",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner": {Type: "string", Description: "Propietario del repositorio"},
					"repo":  {Type: "string", Description: "Nombre del repositorio"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_mark_notification_read",
			Description: "✅ Marca una notificación como leída",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"thread_id": {Type: "string", Description: "ID del thread de la notificación"},
				},
				Required: []string{"thread_id"},
			},
		},
	}
}
