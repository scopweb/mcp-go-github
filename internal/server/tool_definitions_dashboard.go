package server

import "github.com/scopweb/mcp-go-github/pkg/types"

// ListDashboardTools returns the consolidated GitHub dashboard tool
func ListDashboardTools() []types.Tool {
	return []types.Tool{
		{
			Name: "github_dashboard",
			Description: "GitHub dashboard and notifications. Operations: " +
				"full (complete dashboard: notifications, issues, PRs, security, workflows; optional owner, repo), " +
				"notifications (pending notifications; optional all, participating), " +
				"issues (issues assigned to you; optional state: open/closed/all), " +
				"prs_review (PRs pending your review), " +
				"security (security alerts: Dependabot, Secret, Code scanning; requires owner, repo; optional type: dependabot/secret/code/all), " +
				"workflows (failed GitHub Actions workflows; requires owner, repo), " +
				"mark_read (mark notification as read; requires thread_id).",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation":     {Type: "string", Description: "Operation: full, notifications, issues, prs_review, security, workflows, mark_read"},
					"owner":         {Type: "string", Description: "Repository owner (for full, security, workflows)"},
					"repo":          {Type: "string", Description: "Repository name (for full, security, workflows)"},
					"all":           {Type: "boolean", Description: "Include read notifications (for notifications)"},
					"participating": {Type: "boolean", Description: "Only participating notifications (for notifications)"},
					"state":         {Type: "string", Description: "Issue state: open, closed, all (for issues, default: open)"},
					"type":          {Type: "string", Description: "Alert type: dependabot, secret, code, all (for security, default: all)"},
					"thread_id":     {Type: "string", Description: "Notification thread ID (for mark_read)"},
				},
				Required: []string{"operation"},
			},
		},
	}
}
