package interfaces

import (
	"context"

	"github.com/google/go-github/v81/github"
	"github.com/jotajotape/github-go-server-mcp/pkg/types"
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

	// Advanced branch operations
	CheckoutRemote(remoteBranch string, localBranch string) (string, error)
	Merge(sourceBranch string, targetBranch string) (string, error)
	Rebase(branch string) (string, error)

	// Enhanced pull/push operations
	PullWithStrategy(branch string, strategy string) (string, error)
	ForcePush(branch string, force bool) (string, error)
	PushUpstream(branch string) (string, error)

	// Batch operations
	SyncWithRemote(remoteBranch string) (string, error)
	SafeMerge(source string, target string) (string, error)

	// Conflict management
	ConflictStatus() (string, error)
	ResolveConflicts(strategy string) (string, error)

	// Validation operations
	ValidateCleanState() (bool, error)
	DetectPotentialConflicts(sourceBranch string, targetBranch string) (string, error)
	CreateBackup(name string) (string, error)

	// Phase 1: Essential commands (Fase 1)
	Reset(mode string, target string, files []string) (string, error)

	// Phase 2: Conflict management (Fase 2)
	ShowConflict(filePath string) (string, error)
	ResolveFile(filePath string, strategy string, customContent *string) (string, error)

	// Repository initialization
	Init(path string, initialBranch string) (string, error)
}

// GitHubOperations define la interfaz para las operaciones de GitHub.
type GitHubOperations interface {
	ListRepositories(ctx context.Context, listType string) ([]*github.Repository, error)
	CreateRepository(ctx context.Context, name, description string, private bool) (*github.Repository, error)
	ListPullRequests(ctx context.Context, owner, repo, state string) ([]*github.PullRequest, error)
	CreatePullRequest(ctx context.Context, owner, repo, title, head, base, body string) (*github.PullRequest, error)
	CreateFile(ctx context.Context, owner, repo, path, content, message, branch string) (*github.RepositoryContentResponse, error)
	UpdateFile(ctx context.Context, owner, repo, path, content, message, sha, branch string) (*github.RepositoryContentResponse, error)

	// Issue operations
	CreateIssueComment(ctx context.Context, owner, repo string, number int, body string) (*github.IssueComment, error)
	CloseIssue(ctx context.Context, owner, repo string, number int, comment string) (*github.Issue, error)

	// Pull Request operations
	CreatePRComment(ctx context.Context, owner, repo string, number int, body string) (*github.IssueComment, error)
	CreatePRReview(ctx context.Context, owner, repo string, number int, event, body string) (*github.PullRequestReview, error)
	MergePullRequest(ctx context.Context, owner, repo string, number int, commitMessage, mergeMethod string) (*github.PullRequestMergeResult, error)

	// Workflow operations
	RerunWorkflow(ctx context.Context, owner, repo string, runID int64) error
	RerunFailedJobs(ctx context.Context, owner, repo string, runID int64) error

	// Security alert operations
	DismissDependabotAlert(ctx context.Context, owner, repo string, number int, reason, comment string) (*github.DependabotAlert, error)
	DismissCodeScanningAlert(ctx context.Context, owner, repo string, number int64, reason, comment string) (*github.Alert, error)
	DismissSecretScanningAlert(ctx context.Context, owner, repo string, number int64, resolution string) (*github.SecretScanningAlert, error)
}

// AdminOperations define la interfaz para operaciones administrativas de GitHub (v3.0)
type AdminOperations interface {
	// Repository Settings
	GetRepositorySettings(ctx context.Context, owner, repo string) (*github.Repository, error)
	UpdateRepositorySettings(ctx context.Context, owner, repo string, settings map[string]interface{}) (*github.Repository, error)
	ArchiveRepository(ctx context.Context, owner, repo string) (*github.Repository, error)
	DeleteRepository(ctx context.Context, owner, repo string) error

	// Branch Protection
	GetBranchProtection(ctx context.Context, owner, repo, branch string) (*github.Protection, error)
	UpdateBranchProtection(ctx context.Context, owner, repo, branch string, protection *github.ProtectionRequest) (*github.Protection, error)
	DeleteBranchProtection(ctx context.Context, owner, repo, branch string) error

	// Webhooks
	ListWebhooks(ctx context.Context, owner, repo string) ([]*github.Hook, error)
	CreateWebhook(ctx context.Context, owner, repo string, config map[string]interface{}) (*github.Hook, error)
	UpdateWebhook(ctx context.Context, owner, repo string, hookID int64, config map[string]interface{}) (*github.Hook, error)
	DeleteWebhook(ctx context.Context, owner, repo string, hookID int64) error
	TestWebhook(ctx context.Context, owner, repo string, hookID int64) error

	// Collaborators
	ListCollaborators(ctx context.Context, owner, repo string) ([]*github.User, error)
	AddCollaborator(ctx context.Context, owner, repo, username, permission string) (*github.CollaboratorInvitation, error)
	UpdateCollaboratorPermission(ctx context.Context, owner, repo, username, permission string) (*github.CollaboratorInvitation, error)
	RemoveCollaborator(ctx context.Context, owner, repo, username string) error
	CheckCollaborator(ctx context.Context, owner, repo, username string) (bool, error)

	// Repository Invitations
	ListInvitations(ctx context.Context, owner, repo string) ([]*github.RepositoryInvitation, error)
	AcceptInvitation(ctx context.Context, invitationID int64) error
	CancelInvitation(ctx context.Context, owner, repo string, invitationID int64) error

	// Team Access
	ListRepositoryTeams(ctx context.Context, owner, repo string) ([]*github.Team, error)
	AddRepositoryTeam(ctx context.Context, owner, repo string, teamID int64, permission string) error
}
