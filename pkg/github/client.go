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
}

// Client implementa la interfaz GitHubOperations.
type Client struct {
	Repositories RepositoriesService
	PullRequests PullRequestsService
}

// NewClient crea un nuevo cliente para interactuar con la API de GitHub.
// Acepta un cliente de go-github y extrae los servicios necesarios.
func NewClient(ghClient *github.Client) interfaces.GitHubOperations {
	return &Client{
		Repositories: ghClient.Repositories,
		PullRequests: ghClient.PullRequests,
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
