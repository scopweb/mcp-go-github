package server

import "github.com/jotajotape/github-go-server-mcp/pkg/types"

// ListGitHubAPITools returns the consolidated GitHub API tool
func ListGitHubAPITools() []types.Tool {
	return []types.Tool{
		{
			Name: "github_repo",
			Description: "GitHub repository and PR operations via API. Operations: " +
				"list_repos (list your repositories; optional type: all/owner/member), " +
				"create_repo (create new repository; requires name; optional description, private), " +
				"list_prs (list pull requests; requires owner, repo; optional state: open/closed/all), " +
				"create_pr (create pull request; requires owner, repo, title, head, base; optional body).",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation":   {Type: "string", Description: "Operation to perform: list_repos, create_repo, list_prs, create_pr"},
					"type":        {Type: "string", Description: "Repository type filter: all, owner, member (for list_repos)"},
					"name":        {Type: "string", Description: "Repository name (for create_repo)"},
					"description": {Type: "string", Description: "Repository description (for create_repo)"},
					"private":     {Type: "boolean", Description: "Make repository private (for create_repo)"},
					"owner":       {Type: "string", Description: "Repository owner (for list_prs, create_pr)"},
					"repo":        {Type: "string", Description: "Repository name (for list_prs, create_pr)"},
					"state":       {Type: "string", Description: "PR state: open, closed, all (for list_prs)"},
					"title":       {Type: "string", Description: "PR title (for create_pr)"},
					"body":        {Type: "string", Description: "PR description (for create_pr)"},
					"head":        {Type: "string", Description: "Source branch (for create_pr)"},
					"base":        {Type: "string", Description: "Target branch (for create_pr)"},
				},
				Required: []string{"operation"},
			},
		},
	}
}
