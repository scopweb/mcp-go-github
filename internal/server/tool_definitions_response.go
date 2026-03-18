package server

import "github.com/scopweb/mcp-go-github/pkg/types"

// ListResponseTools returns the consolidated GitHub response tool
func ListResponseTools() []types.Tool {
	return []types.Tool{
		{
			Name: "github_respond",
			Description: "Respond to GitHub issues and PRs. Operations: " +
				"comment_issue (add comment to issue; requires owner, repo, number, body), " +
				"comment_pr (add comment to PR; requires owner, repo, number, body), " +
				"review_pr (create PR review; requires owner, repo, number, event: APPROVE/REQUEST_CHANGES/COMMENT; optional body).",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation": {Type: "string", Description: "Operation: comment_issue, comment_pr, review_pr"},
					"owner":     {Type: "string", Description: "Repository owner"},
					"repo":      {Type: "string", Description: "Repository name"},
					"number":    {Type: "number", Description: "Issue or PR number"},
					"body":      {Type: "string", Description: "Comment text or review body (supports Markdown)"},
					"event":     {Type: "string", Description: "Review type: APPROVE, REQUEST_CHANGES, COMMENT (for review_pr)"},
				},
				Required: []string{"operation", "owner", "repo", "number"},
			},
		},
	}
}
