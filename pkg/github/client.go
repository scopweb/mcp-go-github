package github

import (
	"context"

	"github.com/google/go-github/v77/github"
	"github.com/jotajotape/github-go-server-mcp/pkg/interfaces"
)

// RepositoriesService define la interfaz para interactuar con la API de repositorios de GitHub.
// Se utiliza para permitir la simulación del cliente de GitHub en las pruebas.
type RepositoriesService interface {
	List(ctx context.Context, user string, opts *github.RepositoryListOptions) ([]*github.Repository, *github.Response, error)
	Create(ctx context.Context, org string, repo *github.Repository) (*github.Repository, *github.Response, error)
	CreateFile(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error)
	UpdateFile(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error)
}

// PullRequestsService define la interfaz para interactuar con la API de pull requests de GitHub.
type PullRequestsService interface {
	List(ctx context.Context, owner, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)
	Create(ctx context.Context, owner, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error)
	Merge(ctx context.Context, owner, repo string, number int, commitMessage string, opts *github.PullRequestOptions) (*github.PullRequestMergeResult, *github.Response, error)
	CreateReview(ctx context.Context, owner, repo string, number int, review *github.PullRequestReviewRequest) (*github.PullRequestReview, *github.Response, error)
}

// IssuesService define la interfaz para interactuar con la API de issues de GitHub.
type IssuesService interface {
	CreateComment(ctx context.Context, owner, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
	Edit(ctx context.Context, owner, repo string, number int, issue *github.IssueRequest) (*github.Issue, *github.Response, error)
}

// ActionsService define la interfaz para interactuar con GitHub Actions.
type ActionsService interface {
	RerunWorkflowByID(ctx context.Context, owner, repo string, runID int64) (*github.Response, error)
	RerunFailedJobsByID(ctx context.Context, owner, repo string, runID int64) (*github.Response, error)
}

// DependabotService define la interfaz para alertas de Dependabot.
type DependabotService interface {
	UpdateAlert(ctx context.Context, owner, repo string, number int, stateInfo *github.DependabotAlertState) (*github.DependabotAlert, *github.Response, error)
}

// CodeScanningService define la interfaz para alertas de Code Scanning.
type CodeScanningService interface {
	UpdateAlert(ctx context.Context, owner, repo string, id int64, stateInfo *github.CodeScanningAlertState) (*github.Alert, *github.Response, error)
}

// SecretScanningService define la interfaz para alertas de Secret Scanning.
type SecretScanningService interface {
	UpdateAlert(ctx context.Context, owner, repo string, number int64, opts *github.SecretScanningAlertUpdateOptions) (*github.SecretScanningAlert, *github.Response, error)
}

// Client implementa la interfaz GitHubOperations.
type Client struct {
	Repositories RepositoriesService
	PullRequests PullRequestsService
	Issues       IssuesService
	Actions      ActionsService
	Dependabot   DependabotService
	CodeScanning CodeScanningService
	SecretScanning SecretScanningService
}

// NewClient crea un nuevo cliente para interactuar con la API de GitHub.
// Acepta un cliente de go-github y extrae los servicios necesarios.
func NewClient(ghClient *github.Client) interfaces.GitHubOperations {
	return &Client{
		Repositories:  ghClient.Repositories,
		PullRequests:  ghClient.PullRequests,
		Issues:        ghClient.Issues,
		Actions:       ghClient.Actions,
		Dependabot:    ghClient.Dependabot,
		CodeScanning:  ghClient.CodeScanning,
		SecretScanning: ghClient.SecretScanning,
	}
}

// NewClientForTest crea un cliente con servicios simulados para pruebas.
func NewClientForTest(repos RepositoriesService, prs PullRequestsService) interfaces.GitHubOperations {
	return &Client{
		Repositories: repos,
		PullRequests: prs,
	}
}

// ListRepositories lista los repositorios del usuario.
func (c *Client) ListRepositories(ctx context.Context, listType string) ([]*github.Repository, error) {
	opt := &github.RepositoryListOptions{
		Type: listType,
	}
	repos, _, err := c.Repositories.List(ctx, "", opt)
	return repos, err
}

// CreateRepository crea un nuevo repositorio.
func (c *Client) CreateRepository(ctx context.Context, name, description string, private bool) (*github.Repository, error) {
	repo := &github.Repository{
		Name:        github.Ptr(name),
		Description: github.Ptr(description),
		Private:     github.Ptr(private),
	}
	newRepo, _, err := c.Repositories.Create(ctx, "", repo)
	return newRepo, err
}

// ListPullRequests lista los pull requests de un repositorio.
func (c *Client) ListPullRequests(ctx context.Context, owner, repo, state string) ([]*github.PullRequest, error) {
	opt := &github.PullRequestListOptions{
		State: state,
	}
	prs, _, err := c.PullRequests.List(ctx, owner, repo, opt)
	return prs, err
}

// CreatePullRequest crea un nuevo pull request.
func (c *Client) CreatePullRequest(ctx context.Context, owner, repo, title, head, base, body string) (*github.PullRequest, error) {
	newPR := &github.NewPullRequest{
		Title: github.Ptr(title),
		Head:  github.Ptr(head),
		Base:  github.Ptr(base),
		Body:  github.Ptr(body),
	}
	pr, _, err := c.PullRequests.Create(ctx, owner, repo, newPR)
	return pr, err
}

// CreateFile crea un nuevo archivo en un repositorio.
func (c *Client) CreateFile(ctx context.Context, owner, repo, path, content, message, branch string) (*github.RepositoryContentResponse, error) {
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(message),
		Content: []byte(content),
		Branch:  github.Ptr(branch),
	}
	res, _, err := c.Repositories.CreateFile(ctx, owner, repo, path, opts)
	return res, err
}

// UpdateFile actualiza un archivo existente en un repositorio.
func (c *Client) UpdateFile(ctx context.Context, owner, repo, path, content, message, sha, branch string) (*github.RepositoryContentResponse, error) {
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(message),
		Content: []byte(content),
		SHA:     github.Ptr(sha),
		Branch:  github.Ptr(branch),
	}
	res, _, err := c.Repositories.UpdateFile(ctx, owner, repo, path, opts)
	return res, err
}

// === ISSUE OPERATIONS ===

// CreateIssueComment crea un comentario en un issue.
func (c *Client) CreateIssueComment(ctx context.Context, owner, repo string, number int, body string) (*github.IssueComment, error) {
	comment := &github.IssueComment{
		Body: github.Ptr(body),
	}
	result, _, err := c.Issues.CreateComment(ctx, owner, repo, number, comment)
	return result, err
}

// CloseIssue cierra un issue.
func (c *Client) CloseIssue(ctx context.Context, owner, repo string, number int, comment string) (*github.Issue, error) {
	req := &github.IssueRequest{
		State: github.Ptr("closed"),
	}
	if comment != "" {
		req.Body = github.Ptr(comment)
	}
	result, _, err := c.Issues.Edit(ctx, owner, repo, number, req)
	return result, err
}

// === PULL REQUEST OPERATIONS ===

// CreatePRComment crea un comentario en un pull request.
func (c *Client) CreatePRComment(ctx context.Context, owner, repo string, number int, body string) (*github.IssueComment, error) {
	return c.CreateIssueComment(ctx, owner, repo, number, body)
}

// CreatePRReview crea una review en un pull request.
func (c *Client) CreatePRReview(ctx context.Context, owner, repo string, number int, event, body string) (*github.PullRequestReview, error) {
	review := &github.PullRequestReviewRequest{
		Event: github.Ptr(event), // "APPROVE", "REQUEST_CHANGES", "COMMENT"
		Body:  github.Ptr(body),
	}
	result, _, err := c.PullRequests.CreateReview(ctx, owner, repo, number, review)
	return result, err
}

// MergePullRequest mergea un pull request.
func (c *Client) MergePullRequest(ctx context.Context, owner, repo string, number int, commitMessage, mergeMethod string) (*github.PullRequestMergeResult, error) {
	opts := &github.PullRequestOptions{
		MergeMethod: mergeMethod, // "merge", "squash", "rebase"
	}
	result, _, err := c.PullRequests.Merge(ctx, owner, repo, number, commitMessage, opts)
	return result, err
}

// === WORKFLOW OPERATIONS ===

// RerunWorkflow re-ejecuta un workflow.
func (c *Client) RerunWorkflow(ctx context.Context, owner, repo string, runID int64) error {
	_, err := c.Actions.RerunWorkflowByID(ctx, owner, repo, runID)
	return err
}

// RerunFailedJobs re-ejecuta solo los jobs fallidos en un workflow.
func (c *Client) RerunFailedJobs(ctx context.Context, owner, repo string, runID int64) error {
	_, err := c.Actions.RerunFailedJobsByID(ctx, owner, repo, runID)
	return err
}

// === SECURITY ALERT OPERATIONS ===

// DismissDependabotAlert dismissa una alerta de Dependabot.
func (c *Client) DismissDependabotAlert(ctx context.Context, owner, repo string, number int, reason, comment string) (*github.DependabotAlert, error) {
	state := &github.DependabotAlertState{
		State:            "dismissed",
		DismissedReason:  github.Ptr(reason), // "fix_started", "inaccurate", "no_bandwidth", "not_used", "tolerable_risk"
		DismissedComment: github.Ptr(comment),
	}
	result, _, err := c.Dependabot.UpdateAlert(ctx, owner, repo, number, state)
	return result, err
}

// DismissCodeScanningAlert dismissa una alerta de code scanning.
func (c *Client) DismissCodeScanningAlert(ctx context.Context, owner, repo string, number int64, reason, comment string) (*github.Alert, error) {
	state := &github.CodeScanningAlertState{
		State:            "dismissed",
		DismissedReason:  github.Ptr(reason), // "false positive", "won't fix", "used in tests"
		DismissedComment: github.Ptr(comment),
	}
	result, _, err := c.CodeScanning.UpdateAlert(ctx, owner, repo, number, state)
	return result, err
}

// DismissSecretScanningAlert dismissa una alerta de secret scanning.
func (c *Client) DismissSecretScanningAlert(ctx context.Context, owner, repo string, number int64, resolution string) (*github.SecretScanningAlert, error) {
	opts := &github.SecretScanningAlertUpdateOptions{
		State:      "resolved",
		Resolution: github.Ptr(resolution), // "false_positive", "wont_fix", "revoked", "used_in_tests"
	}
	result, _, err := c.SecretScanning.UpdateAlert(ctx, owner, repo, number, opts)
	return result, err
}
