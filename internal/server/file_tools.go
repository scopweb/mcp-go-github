package server

import "github.com/jotajotape/github-go-server-mcp/pkg/types"

// ListFileTools returns tools for GitHub file operations (no Git required)
// Consolidated into 1 tool using the operation parameter pattern.
func ListFileTools() []types.Tool {
	return []types.Tool{
		{
			Name:  "github_files",
			Title: "GitHub File Operations",
			Description: "File operations via GitHub API (no Git required). Operations: " +
				"list (list files and directories at a repository path), " +
				"download (download a single file to local disk, requires path), " +
				"download_repo (download entire repository to a local directory), " +
				"pull_repo (update local directory from repository, like a pull via API).",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation":  {Type: "string", Description: "Operation to perform: list, download, download_repo, pull_repo"},
					"owner":      {Type: "string", Description: "Repository owner"},
					"repo":       {Type: "string", Description: "Repository name"},
					"path":       {Type: "string", Description: "File or directory path in the repository (for list and download)"},
					"branch":     {Type: "string", Description: "Branch name (default: main)"},
					"local_path": {Type: "string", Description: "Local path to save file (for download, default: same as repo path)"},
					"local_dir":  {Type: "string", Description: "Local directory to save files (for download_repo/pull_repo, default: ./<repo>)"},
				},
				Required: []string{"operation", "owner", "repo"},
			},
		},
	}
}

// IsFileOperation checks if a tool name is a file operation
func IsFileOperation(name string) bool {
	return name == "github_files"
}
