package interfaces

import (
	"context"

	"github.com/google/go-github/v74/github"
	"github.com/scopweb/mcp-go-github/internal/types"
)

// GitOperations define la interfaz para las operaciones de Git.
type GitOperations interface {
	HasGit() bool
	IsGitRepo() bool
	GetRepoPath() string
	GetCurrentBranch() string
	GetRemoteURL() string
	Status() (string, error)
	Add(path string) (string, error)
	Commit(message string) (string, error)
	Push(branch string) (string, error)
	Pull(branch string) (string, error)
	Checkout(branch string, create bool) (string, error)
	CreateFile(path, content string) (string, error)
	UpdateFile(path, content, sha string) (string, error)
	BranchList(remote bool) ([]types.BranchInfo, error)
	SetWorkspace(path string) (string, error)
	GetFileSHA(path string) (string, error)
	GetLastCommit() (string, error)
	GetFileContent(path, ref string) (string, error)
	GetChangedFiles(staged bool) (string, error)
	ValidateRepo(path string) (string, error)
	ListFiles(ref string) (string, error)
	LogAnalysis(limit string) (string, error)
	DiffFiles(staged bool) (string, error)
	Stash(operation, name string) (string, error)
	Remote(operation, name, url string) (string, error)
	Tag(operation, tagName, message string) (string, error)
	Clean(operation string, dryRun bool) (string, error)
}

// GitHubOperations define la interfaz para las operaciones de GitHub.
type GitHubOperations interface {
	ListRepositories(ctx context.Context, listType string) ([]*github.Repository, error)
	CreateRepository(ctx context.Context, name, description string, private bool) (*github.Repository, error)
	ListPullRequests(ctx context.Context, owner, repo, state string) ([]*github.PullRequest, error)
	CreatePullRequest(ctx context.Context, owner, repo, title, head, base, body string) (*github.PullRequest, error)
	CreateFile(ctx context.Context, owner, repo, path, content, message, branch string) (*github.RepositoryContentResponse, error)
	UpdateFile(ctx context.Context, owner, repo, path, content, message, sha, branch string) (*github.RepositoryContentResponse, error)
}
