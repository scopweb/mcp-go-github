package server

import "github.com/scopweb/mcp-go-github/pkg/types"

// ListGitBasicTools returns the core Git workflow tools (kept individual for fast access)
func ListGitBasicTools() []types.Tool {
	return []types.Tool{
		{
			Name:        "git_init",
			Description: "Initialize a new Git repository in specified directory",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"path":           {Type: "string", Description: "Directory path to initialize (must exist)"},
					"initial_branch": {Type: "string", Description: "Initial branch name (default: main)"},
				},
				Required: []string{"path"},
			},
			Annotations: IdempotentAnnotation(),
		},
		{
			Name:        "git_add",
			Description: "Stage files for commit (use . for all files)",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"files": {Type: "string", Description: "Files to stage (. for all)"},
				},
				Required: []string{"files"},
			},
		},
		{
			Name:        "git_commit",
			Description: "Commit staged changes with a message",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"message": {Type: "string", Description: "Commit message"},
				},
				Required: []string{"message"},
			},
		},
	}
}
