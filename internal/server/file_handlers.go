package server

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v81/github"
	"github.com/jotajotape/github-go-server-mcp/pkg/types"
)

// getGitHubClient extracts the raw github.Client from the server
func getGitHubClient(s *MCPServer) (*github.Client, error) {
	if s.RawGitHubClient == nil {
		return nil, fmt.Errorf("GitHub client not configured")
	}
	client, ok := s.RawGitHubClient.(*github.Client)
	if !ok {
		return nil, fmt.Errorf("invalid GitHub client type")
	}
	return client, nil
}

// HandleFileTool routes file operation tool calls
func HandleFileTool(s *MCPServer, name string, args map[string]interface{}) (types.ToolCallResult, error) {
	ctx := context.Background()

	switch name {
	case "github_list_repo_contents":
		return handleListRepoContents(s, ctx, args)
	case "github_download_file":
		return handleDownloadFile(s, ctx, args)
	case "github_download_repo":
		return handleDownloadRepo(s, ctx, args)
	case "github_pull_repo":
		return handlePullRepo(s, ctx, args)
	default:
		return types.ToolCallResult{}, fmt.Errorf("unknown file operation: %s", name)
	}
}

// handleListRepoContents lists files in a repository directory via API
func handleListRepoContents(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	client, err := getGitHubClient(s)
	if err != nil {
		return types.ToolCallResult{}, err
	}

	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	path, _ := args["path"].(string)
	branch, _ := args["branch"].(string)

	opts := &github.RepositoryContentGetOptions{}
	if branch != "" {
		opts.Ref = branch
	}

	fileContent, dirContent, _, err := client.Repositories.GetContents(ctx, owner, repo, path, opts)
	if err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to list contents: %w", err)
	}

	var text string

	if fileContent != nil {
		// Single file
		text = fmt.Sprintf("📄 File: %s/%s/%s\n\n", owner, repo, path)
		text += fmt.Sprintf("Name: %s\n", fileContent.GetName())
		text += fmt.Sprintf("Size: %d bytes\n", fileContent.GetSize())
		text += fmt.Sprintf("SHA: %s\n", fileContent.GetSHA())
		text += fmt.Sprintf("Type: %s\n", fileContent.GetType())
	} else if dirContent != nil {
		// Directory listing
		displayPath := path
		if displayPath == "" {
			displayPath = "/"
		}
		text = fmt.Sprintf("📂 Contents of %s/%s/%s (%d items)\n\n", owner, repo, displayPath, len(dirContent))

		dirs := []string{}
		files := []string{}

		for _, item := range dirContent {
			if item.GetType() == "dir" {
				dirs = append(dirs, fmt.Sprintf("📁 %s/", item.GetName()))
			} else {
				size := item.GetSize()
				sizeStr := formatSize(size)
				files = append(files, fmt.Sprintf("📄 %s (%s)", item.GetName(), sizeStr))
			}
		}

		// Directories first, then files
		for _, d := range dirs {
			text += d + "\n"
		}
		for _, f := range files {
			text += f + "\n"
		}
	}

	return types.ToolCallResult{
		Content: []types.Content{{Type: "text", Text: text}},
	}, nil
}

// handleDownloadFile downloads a single file from repository to local disk
func handleDownloadFile(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	client, err := getGitHubClient(s)
	if err != nil {
		return types.ToolCallResult{}, err
	}

	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	path, _ := args["path"].(string)
	branch, _ := args["branch"].(string)
	localPath, _ := args["local_path"].(string)

	if localPath == "" {
		localPath = path
	}

	opts := &github.RepositoryContentGetOptions{}
	if branch != "" {
		opts.Ref = branch
	}

	// Download file content
	rc, _, err := client.Repositories.DownloadContents(ctx, owner, repo, path, opts)
	if err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to download %s: %w", path, err)
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to read content: %w", err)
	}

	// Create directories if needed
	dir := filepath.Dir(localPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return types.ToolCallResult{}, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write file
	if err := os.WriteFile(localPath, content, 0644); err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to write file %s: %w", localPath, err)
	}

	text := fmt.Sprintf("✅ Downloaded %s/%s/%s\n→ Saved to: %s\n→ Size: %s",
		owner, repo, path, localPath, formatSize(len(content)))

	return types.ToolCallResult{
		Content: []types.Content{{Type: "text", Text: text}},
	}, nil
}

// handleDownloadRepo downloads entire repository to local directory
func handleDownloadRepo(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	client, err := getGitHubClient(s)
	if err != nil {
		return types.ToolCallResult{}, err
	}

	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	branch, _ := args["branch"].(string)
	localDir, _ := args["local_dir"].(string)

	if branch == "" {
		branch = "main"
	}
	if localDir == "" {
		localDir = repo
	}

	// Get the full tree recursively
	ref, _, err := client.Git.GetRef(ctx, owner, repo, "refs/heads/"+branch)
	if err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to get branch '%s': %w", branch, err)
	}

	treeSHA := ref.GetObject().GetSHA()
	tree, _, err := client.Git.GetTree(ctx, owner, repo, treeSHA, true)
	if err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to get repository tree: %w", err)
	}

	// Create local directory
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to create directory %s: %w", localDir, err)
	}

	// Download all files
	downloaded := 0
	skipped := 0
	totalSize := 0
	var errors []string

	opts := &github.RepositoryContentGetOptions{Ref: branch}

	for _, entry := range tree.Entries {
		if entry.GetType() == "tree" {
			// Create directory
			dirPath := filepath.Join(localDir, entry.GetPath())
			os.MkdirAll(dirPath, 0755)
			continue
		}

		if entry.GetType() != "blob" {
			continue
		}

		// Skip very large files (>10MB)
		if entry.GetSize() > 10*1024*1024 {
			skipped++
			continue
		}

		filePath := entry.GetPath()
		localFilePath := filepath.Join(localDir, filePath)

		// Create parent directory
		parentDir := filepath.Dir(localFilePath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			errors = append(errors, fmt.Sprintf("mkdir %s: %v", parentDir, err))
			continue
		}

		// Download file
		rc, _, err := client.Repositories.DownloadContents(ctx, owner, repo, filePath, opts)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", filePath, err))
			continue
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			errors = append(errors, fmt.Sprintf("read %s: %v", filePath, err))
			continue
		}

		if err := os.WriteFile(localFilePath, content, 0644); err != nil {
			errors = append(errors, fmt.Sprintf("write %s: %v", localFilePath, err))
			continue
		}

		downloaded++
		totalSize += len(content)
	}

	// Build result
	text := fmt.Sprintf("📦 Downloaded %s/%s (branch: %s)\n\n", owner, repo, branch)
	text += fmt.Sprintf("→ Directory: %s\n", localDir)
	text += fmt.Sprintf("→ Files downloaded: %d\n", downloaded)
	text += fmt.Sprintf("→ Total size: %s\n", formatSize(totalSize))

	if skipped > 0 {
		text += fmt.Sprintf("→ Skipped (>10MB): %d\n", skipped)
	}

	if len(errors) > 0 {
		text += fmt.Sprintf("\n⚠️ Errors (%d):\n", len(errors))
		for _, e := range errors {
			if len(errors) <= 10 {
				text += fmt.Sprintf("  • %s\n", e)
			}
		}
		if len(errors) > 10 {
			text += fmt.Sprintf("  ... and %d more\n", len(errors)-10)
		}
	}

	return types.ToolCallResult{
		Content: []types.Content{{Type: "text", Text: text}},
	}, nil
}

// handlePullRepo updates local directory from repository (API-based pull)
func handlePullRepo(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	client, err := getGitHubClient(s)
	if err != nil {
		return types.ToolCallResult{}, err
	}

	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	branch, _ := args["branch"].(string)
	localDir, _ := args["local_dir"].(string)

	if branch == "" {
		branch = "main"
	}
	if localDir == "" {
		localDir = repo
	}

	// Check if local directory exists
	if _, err := os.Stat(localDir); os.IsNotExist(err) {
		return types.ToolCallResult{}, fmt.Errorf("local directory '%s' does not exist. Use github_download_repo first", localDir)
	}

	// Get the full tree
	ref, _, err := client.Git.GetRef(ctx, owner, repo, "refs/heads/"+branch)
	if err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to get branch '%s': %w", branch, err)
	}

	treeSHA := ref.GetObject().GetSHA()
	tree, _, err := client.Git.GetTree(ctx, owner, repo, treeSHA, true)
	if err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to get repository tree: %w", err)
	}

	opts := &github.RepositoryContentGetOptions{Ref: branch}

	updated := 0
	created := 0
	unchanged := 0
	totalSize := 0
	var errors []string

	for _, entry := range tree.Entries {
		if entry.GetType() == "tree" {
			dirPath := filepath.Join(localDir, entry.GetPath())
			os.MkdirAll(dirPath, 0755)
			continue
		}

		if entry.GetType() != "blob" {
			continue
		}

		// Skip very large files
		if entry.GetSize() > 10*1024*1024 {
			continue
		}

		filePath := entry.GetPath()
		localFilePath := filepath.Join(localDir, filePath)

		// Check if file exists and compare SHA
		needsUpdate := false
		fileExists := false

		if localContent, err := os.ReadFile(localFilePath); err == nil {
			fileExists = true
			// Get remote content to compare via SHA
			fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repo, filePath, opts)
			if err == nil && fileContent != nil {
				remoteContent, decErr := fileContent.GetContent()
				if decErr == nil {
					// Compare content
					if string(localContent) != remoteContent {
						needsUpdate = true
					}
				} else {
					// Can't decode, download directly
					needsUpdate = true
				}
			}
		} else {
			needsUpdate = true
		}

		if !needsUpdate {
			unchanged++
			continue
		}

		// Download and update
		parentDir := filepath.Dir(localFilePath)
		os.MkdirAll(parentDir, 0755)

		rc, _, err := client.Repositories.DownloadContents(ctx, owner, repo, filePath, opts)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", filePath, err))
			continue
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			errors = append(errors, fmt.Sprintf("read %s: %v", filePath, err))
			continue
		}

		if err := os.WriteFile(localFilePath, content, 0644); err != nil {
			errors = append(errors, fmt.Sprintf("write %s: %v", localFilePath, err))
			continue
		}

		totalSize += len(content)
		if fileExists {
			updated++
		} else {
			created++
		}
	}

	// Build result
	text := fmt.Sprintf("🔄 Updated %s/%s (branch: %s)\n\n", owner, repo, branch)
	text += fmt.Sprintf("→ Directory: %s\n", localDir)

	if updated == 0 && created == 0 {
		text += "→ Already up to date!\n"
	} else {
		if created > 0 {
			text += fmt.Sprintf("→ New files: %d\n", created)
		}
		if updated > 0 {
			text += fmt.Sprintf("→ Updated files: %d\n", updated)
		}
		text += fmt.Sprintf("→ Downloaded: %s\n", formatSize(totalSize))
	}
	text += fmt.Sprintf("→ Unchanged: %d\n", unchanged)

	if len(errors) > 0 {
		text += fmt.Sprintf("\n⚠️ Errors (%d):\n", len(errors))
		for i, e := range errors {
			if i < 10 {
				text += fmt.Sprintf("  • %s\n", e)
			}
		}
	}

	return types.ToolCallResult{
		Content: []types.Content{{Type: "text", Text: text}},
	}, nil
}

// formatSize formats byte count to human readable string
func formatSize(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}

// sanitizePath prevents path traversal in file operations
func sanitizePath(path string) string {
	// Remove leading slashes and dots
	path = strings.TrimLeft(path, "/\\.")
	// Clean the path
	path = filepath.Clean(path)
	// Prevent path traversal
	if strings.Contains(path, "..") {
		return ""
	}
	return path
}
