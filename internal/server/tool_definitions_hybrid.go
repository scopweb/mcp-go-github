package server

import "github.com/scopweb/mcp-go-github/pkg/types"

// ListHybridTools returns the hybrid file tools (local Git first, GitHub API fallback).
//
// Naming: prefixed with "gh_" to disambiguate from generic filesystem MCP tools that
// also expose create_file / update_file. Without the prefix, an MCP host with both
// servers active would face name collisions and route calls ambiguously.
func ListHybridTools() []types.Tool {
	return []types.Tool{
		{
			Name:        "gh_create_file",
			Description: "Create a file in the current Git workspace and commit it. Falls back to GitHub API (owner/repo required) if no local Git repo is detected. Prefer this for repos cloned locally — it costs zero tokens vs. the API path.",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"path":    {Type: "string", Description: "File path"},
					"content": {Type: "string", Description: "File content"},
					"message": {Type: "string", Description: "Commit message (optional for local Git)"},
					"owner":   {Type: "string", Description: "Repository owner (required ONLY if local Git is unavailable)"},
					"repo":    {Type: "string", Description: "Repository name (required ONLY if local Git is unavailable)"},
				},
				Required: []string{"path", "content"},
			},
		},
		{
			Name:        "gh_update_file",
			Description: "Update an existing file in the current Git workspace and commit it. Falls back to GitHub API (owner/repo/sha required) if no local Git repo is detected. Prefer this for repos cloned locally.",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"path":    {Type: "string", Description: "File path"},
					"content": {Type: "string", Description: "New content"},
					"message": {Type: "string", Description: "Commit message (optional for local Git)"},
					"owner":   {Type: "string", Description: "Repository owner (required ONLY if local Git is unavailable)"},
					"repo":    {Type: "string", Description: "Repository name (required ONLY if local Git is unavailable)"},
					"sha":     {Type: "string", Description: "File SHA (required ONLY if local Git is unavailable)"},
				},
				Required: []string{"path", "content"},
			},
		},
		{
			Name:        "gh_push_files",
			Description: "Write multiple files and run git add/commit/push in a single call (local Git only). Supports 3 modes: inline content (files: [{path, content}]), copy from disk (files: [{path, source_path}]) which avoids sending content over the wire, and paths-only (paths: [...]) for files already present in the workspace where only git add/commit/push is needed.",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"files":   {Type: "array", Description: "File list: [{path, content}] or [{path, source_path}]. source_path reads from disk without sending content."},
					"paths":   {Type: "array", Description: "Files already present in the workspace: only git add/commit/push (no content transferred)."},
					"message": {Type: "string", Description: "Commit message"},
					"branch":  {Type: "string", Description: "Branch to push to (optional, uses current branch if omitted)"},
				},
				Required: []string{"message"},
			},
		},
	}
}
