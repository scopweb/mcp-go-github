package github

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-github/v77/github"
	"github.com/stretchr/testify/assert"
)

// mockRepositoriesService es una implementación simulada de RepositoriesService para pruebas.
type mockRepositoriesService struct {
	// Sobrescribe los métodos que necesitas simular en tus pruebas.
	// Si un método no se sobrescribe, se llamará a la implementación base (que devuelve nil).
	ListFunc       func(ctx context.Context, user string, opts *github.RepositoryListOptions) ([]*github.Repository, *github.Response, error)
	CreateFunc     func(ctx context.Context, org string, repo *github.Repository) (*github.Repository, *github.Response, error)
	CreateFileFunc func(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error)
	UpdateFileFunc func(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error)
}

func (m *mockRepositoriesService) List(ctx context.Context, user string, opts *github.RepositoryListOptions) ([]*github.Repository, *github.Response, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, user, opts)
	}
	return nil, nil, nil
}

func (m *mockRepositoriesService) Create(ctx context.Context, org string, repo *github.Repository) (*github.Repository, *github.Response, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, org, repo)
	}
	return nil, nil, nil
}

func (m *mockRepositoriesService) CreateFile(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error) {
	if m.CreateFileFunc != nil {
		return m.CreateFileFunc(ctx, owner, repo, path, opts)
	}
	return nil, nil, nil
}

func (m *mockRepositoriesService) UpdateFile(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error) {
	if m.UpdateFileFunc != nil {
		return m.UpdateFileFunc(ctx, owner, repo, path, opts)
	}
	return nil, nil, nil
}

// mockPullRequestsService es una implementación simulada de PullRequestsService para pruebas.
type mockPullRequestsService struct {
	ListFunc   func(ctx context.Context, owner, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)
	CreateFunc func(ctx context.Context, owner, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error)
}

func (m *mockPullRequestsService) List(ctx context.Context, owner, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, owner, repo, opts)
	}
	return nil, nil, nil
}

func (m *mockPullRequestsService) Create(ctx context.Context, owner, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, owner, repo, pull)
	}
	return nil, nil, nil
}

func TestListRepositories(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful repository listing", func(t *testing.T) {
		mockRepos := []*github.Repository{
			{Name: github.Ptr("repo1")},
			{Name: github.Ptr("repo2")},
		}

		mockRepoService := &mockRepositoriesService{
			ListFunc: func(_ context.Context, _ string, opts *github.RepositoryListOptions) ([]*github.Repository, *github.Response, error) {
				assert.Equal(t, "owner", opts.Type)
				return mockRepos, nil, nil
			},
		}

		client := NewClientForTest(mockRepoService, nil)

		repos, err := client.ListRepositories(ctx, "owner")

		assert.NoError(t, err)
		assert.Len(t, repos, 2)
		assert.Equal(t, "repo1", *repos[0].Name)
	})

	t.Run("GitHub API returns an error", func(t *testing.T) {
		expectedError := errors.New("github api error")

		mockRepoService := &mockRepositoriesService{
			ListFunc: func(_ context.Context, _ string, _ *github.RepositoryListOptions) ([]*github.Repository, *github.Response, error) {
				return nil, nil, expectedError
			},
		}

		client := NewClientForTest(mockRepoService, nil)

		_, err := client.ListRepositories(ctx, "all")

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}

func TestCreateRepository(t *testing.T) {
	ctx := context.Background()
	repoName := "new-repo"
	repoDescription := "A new repository."
	isPrivate := true

	t.Run("Successful repository creation", func(t *testing.T) {
		mockRepoService := &mockRepositoriesService{
			CreateFunc: func(_ context.Context, _ string, repo *github.Repository) (*github.Repository, *github.Response, error) {
				assert.Equal(t, repoName, *repo.Name)
				assert.Equal(t, repoDescription, *repo.Description)
				assert.Equal(t, isPrivate, *repo.Private)
				return repo, nil, nil
			},
		}

		client := NewClientForTest(mockRepoService, nil)
		repo, err := client.CreateRepository(ctx, repoName, repoDescription, isPrivate)

		assert.NoError(t, err)
		assert.NotNil(t, repo)
		assert.Equal(t, repoName, *repo.Name)
	})

	t.Run("GitHub API returns an error on creation", func(t *testing.T) {
		expectedError := errors.New("github api create error")

		mockRepoService := &mockRepositoriesService{
			CreateFunc: func(_ context.Context, _ string, _ *github.Repository) (*github.Repository, *github.Response, error) {
				return nil, nil, expectedError
			},
		}

		client := NewClientForTest(mockRepoService, nil)
		_, err := client.CreateRepository(ctx, repoName, repoDescription, isPrivate)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}

func TestListPullRequests(t *testing.T) {
	ctx := context.Background()
	owner := "owner"
	repo := "repo"

	t.Run("Successful pull request listing", func(t *testing.T) {
		mockPRs := []*github.PullRequest{
			{Title: github.Ptr("pr1")},
			{Title: github.Ptr("pr2")},
		}

		mockPRService := &mockPullRequestsService{
			ListFunc: func(_ context.Context, _ string, _ string, _ *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
				return mockPRs, nil, nil
			},
		}

		client := NewClientForTest(nil, mockPRService)
		prs, err := client.ListPullRequests(ctx, owner, repo, "all")

		assert.NoError(t, err)
		assert.Len(t, prs, 2)
		assert.Equal(t, "pr1", *prs[0].Title)
	})

	t.Run("GitHub API returns an error on PR list", func(t *testing.T) {
		expectedError := errors.New("github api pr list error")

		mockPRService := &mockPullRequestsService{
			ListFunc: func(_ context.Context, _ string, _ string, _ *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
				return nil, nil, expectedError
			},
		}

		client := NewClientForTest(nil, mockPRService)
		_, err := client.ListPullRequests(ctx, owner, repo, "all")

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}

func TestCreatePullRequest(t *testing.T) {
	ctx := context.Background()
	owner := "owner"
	repo := "repo"
	title := "New PR"
	head := "feature-branch"
	base := "main"
	body := "This is the body of the PR."

	t.Run("Successful pull request creation", func(t *testing.T) {
		mockPRService := &mockPullRequestsService{
			CreateFunc: func(_ context.Context, _ string, _ string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
				assert.Equal(t, title, *pull.Title)
				assert.Equal(t, head, *pull.Head)
				assert.Equal(t, base, *pull.Base)
				assert.Equal(t, body, *pull.Body)
				return &github.PullRequest{Title: pull.Title}, nil, nil
			},
		}

		client := NewClientForTest(nil, mockPRService)
		pr, err := client.CreatePullRequest(ctx, owner, repo, title, head, base, body)

		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Equal(t, title, *pr.Title)
	})

	t.Run("GitHub API returns an error on PR creation", func(t *testing.T) {
		expectedError := errors.New("github api pr create error")

		mockPRService := &mockPullRequestsService{
			CreateFunc: func(_ context.Context, _ string, _ string, _ *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
				return nil, nil, expectedError
			},
		}

		client := NewClientForTest(nil, mockPRService)
		_, err := client.CreatePullRequest(ctx, owner, repo, title, head, base, body)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}

func TestCreateFile(t *testing.T) {
	ctx := context.Background()
	owner := "owner"
	repo := "repo"
	path := "new-file.txt"
	content := "Hello world"
	message := "feat: add new-file.txt"
	branch := "main"

	t.Run("Successful file creation", func(t *testing.T) {
		mockRepoService := &mockRepositoriesService{
			CreateFileFunc: func(_ context.Context, _ string, _ string, _ string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error) {
				assert.Equal(t, message, *opts.Message)
				assert.Equal(t, []byte(content), opts.Content)
				assert.Equal(t, branch, *opts.Branch)
				return &github.RepositoryContentResponse{}, nil, nil
			},
		}

		client := NewClientForTest(mockRepoService, nil)
		_, err := client.CreateFile(ctx, owner, repo, path, content, message, branch)

		assert.NoError(t, err)
	})

	t.Run("GitHub API returns an error on file creation", func(t *testing.T) {
		expectedError := errors.New("github api create file error")

		mockRepoService := &mockRepositoriesService{
			CreateFileFunc: func(_ context.Context, _ string, _ string, _ string, _ *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error) {
				return nil, nil, expectedError
			},
		}

		client := NewClientForTest(mockRepoService, nil)
		_, err := client.CreateFile(ctx, owner, repo, path, content, message, branch)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}

func TestUpdateFile(t *testing.T) {
	ctx := context.Background()
	owner := "owner"
	repo := "repo"
	path := "existing-file.txt"
	content := "Hello updated world"
	message := "feat: update existing-file.txt"
	sha := "some-sha"
	branch := "main"

	t.Run("Successful file update", func(t *testing.T) {
		mockRepoService := &mockRepositoriesService{
			UpdateFileFunc: func(_ context.Context, _ string, _ string, _ string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error) {
				assert.Equal(t, message, *opts.Message)
				assert.Equal(t, []byte(content), opts.Content)
				assert.Equal(t, sha, *opts.SHA)
				assert.Equal(t, branch, *opts.Branch)
				return &github.RepositoryContentResponse{}, nil, nil
			},
		}

		client := NewClientForTest(mockRepoService, nil)
		_, err := client.UpdateFile(ctx, owner, repo, path, content, message, sha, branch)

		assert.NoError(t, err)
	})

	t.Run("GitHub API returns an error on file update", func(t *testing.T) {
		expectedError := errors.New("github api update file error")

		mockRepoService := &mockRepositoriesService{
			UpdateFileFunc: func(_ context.Context, _ string, _ string, _ string, _ *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error) {
				return nil, nil, expectedError
			},
		}

		client := NewClientForTest(mockRepoService, nil)
		_, err := client.UpdateFile(ctx, owner, repo, path, content, message, sha, branch)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}
