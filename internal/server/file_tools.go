package server

import "github.com/jotajotape/github-go-server-mcp/pkg/types"

// ListFileTools returns tools for GitHub file operations (no Git required)
func ListFileTools() []types.Tool {
	return []types.Tool{
		{
			Name:        "github_list_repo_contents",
			Description: "📂 List files and directories in a repository path (no Git required)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":  {Type: "string", Description: "Repository owner"},
					"repo":   {Type: "string", Description: "Repository name"},
					"path":   {Type: "string", Description: "Directory path (empty for root)"},
					"branch": {Type: "string", Description: "Branch name (default: main)"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_download_file",
			Description: "⬇️ Download a single file from a repository to local disk (no Git required)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":      {Type: "string", Description: "Repository owner"},
					"repo":       {Type: "string", Description: "Repository name"},
					"path":       {Type: "string", Description: "File path in the repository"},
					"branch":     {Type: "string", Description: "Branch name (default: main)"},
					"local_path": {Type: "string", Description: "Local path to save file (default: same as repo path)"},
				},
				Required: []string{"owner", "repo", "path"},
			},
		},
		{
			Name:        "github_download_repo",
			Description: "📦 Download entire repository to local directory (clone via API, no Git required)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":     {Type: "string", Description: "Repository owner"},
					"repo":      {Type: "string", Description: "Repository name"},
					"branch":    {Type: "string", Description: "Branch to download (default: main)"},
					"local_dir": {Type: "string", Description: "Local directory to save files (default: ./<repo>)"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_pull_repo",
			Description: "🔄 Update local directory from repository (pull via API, no Git required)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":     {Type: "string", Description: "Repository owner"},
					"repo":      {Type: "string", Description: "Repository name"},
					"branch":    {Type: "string", Description: "Branch to pull from (default: main)"},
					"local_dir": {Type: "string", Description: "Local directory to update (default: ./<repo>)"},
				},
				Required: []string{"owner", "repo"},
			},
		},
	}
}

// IsFileOperation checks if a tool name is a file operation
func IsFileOperation(name string) bool {
	switch name {
	case "github_list_repo_contents", "github_download_file",
		"github_download_repo", "github_pull_repo":
		return true
	}
	return false
}
