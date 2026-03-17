package server

import "github.com/jotajotape/github-go-server-mcp/pkg/types"

// ListRepairTools returns the consolidated GitHub repair tool for closing, merging, rerunning, and dismissing
func ListRepairTools() []types.Tool {
	return []types.Tool{
		{
			Name:        "github_repair",
			Description: "GitHub repair operations for closing issues, merging PRs, rerunning workflows, and dismissing security alerts. Operations: close_issue (close an issue with optional comment; requires owner, repo, number), merge_pr (merge a pull request; requires owner, repo, number; optional commit_message, merge_method), rerun_workflow (re-run a failed GitHub Actions workflow; requires owner, repo, run_id; optional failed_jobs_only), dismiss_alert (dismiss a security alert; requires owner, repo, number, alert_type; for dependabot requires reason and optional comment; for code requires reason and optional comment; for secret requires resolution).",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation":       {Type: "string", Description: "Operation to perform: close_issue, merge_pr, rerun_workflow, dismiss_alert"},
					"owner":           {Type: "string", Description: "Repository owner"},
					"repo":            {Type: "string", Description: "Repository name"},
					"number":          {Type: "number", Description: "Issue, PR, or alert number"},
					"comment":         {Type: "string", Description: "Closing comment (for close_issue) or dismissal comment (for dismiss_alert with dependabot or code)"},
					"commit_message":  {Type: "string", Description: "Merge commit message (for merge_pr)"},
					"merge_method":    {Type: "string", Description: "Merge method: merge, squash, rebase (for merge_pr, default: merge)"},
					"run_id":          {Type: "number", Description: "Workflow run ID (for rerun_workflow)"},
					"failed_jobs_only": {Type: "boolean", Description: "Re-run only failed jobs (for rerun_workflow, default: false)"},
					"alert_type":      {Type: "string", Description: "Security alert type: dependabot, code, secret (for dismiss_alert)"},
					"reason":          {Type: "string", Description: "Dismissal reason (for dismiss_alert with dependabot: fix_started, inaccurate, no_bandwidth, not_used, tolerable_risk; for code: false positive, won't fix, used in tests)"},
					"resolution":      {Type: "string", Description: "Resolution for secret scanning alerts: false_positive, wont_fix, revoked, used_in_tests (for dismiss_alert with secret)"},
				},
				Required: []string{"operation"},
			},
		},
	}
}
