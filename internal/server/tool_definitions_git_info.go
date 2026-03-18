package server

import "github.com/scopweb/mcp-go-github/pkg/types"

// ListGitInfoTools returns consolidated Git information tools
func ListGitInfoTools() []types.Tool {
	return []types.Tool{
		{
			Name:        "git_info",
			Description: "Git repository information and queries. Operations: status (repo state and config), file_sha (get SHA of a file), last_commit (latest commit SHA), file_content (read file at ref), changed_files (modified files list), validate_repo (check if valid git repo), list_files (all tracked files), context (auto-detect Git local vs API mode), validate_clean (check for uncommitted changes)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation": {Type: "string", Description: "Operation: status, file_sha, last_commit, file_content, changed_files, validate_repo, list_files, context, validate_clean"},
					"path":      {Type: "string", Description: "File or directory path (for file_sha, file_content, validate_repo)"},
					"ref":       {Type: "string", Description: "Git reference - branch, commit, tag (for file_content, list_files). Default: HEAD"},
					"staged":    {Type: "boolean", Description: "Show staged files instead of working directory (for changed_files)"},
				},
				Required: []string{"operation"},
			},
		},
		{
			Name:        "git_set_workspace",
			Description: "Set working directory for all Git operations",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"path": {Type: "string", Description: "Path to Git repository directory"},
				},
				Required: []string{"path"},
			},
		},
	}
}
