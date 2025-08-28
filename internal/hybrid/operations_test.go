package hybrid

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v74/github"
	"github.com/jotajotape/github-go-server-mcp/internal/types"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

// mockGitOperations es una implementación simulada de interfaces.GitOperations
type mockGitOperations struct {
	hasGit              bool
	isGitRepo           bool
	repoPath            string
	currentBranch       string
	remoteURL           string
	createFileFunc      func(path, content string) (string, error)
	updateFileFunc      func(path, content, sha string) (string, error)
	branchListFunc      func(remote bool) ([]types.BranchInfo, error)
	setWorkspaceFunc    func(path string) (string, error)
	getFileSHAFunc      func(path string) (string, error)
	GetChangedFilesFunc func(staged bool) (string, error)
}

func (m *mockGitOperations) HasGit() bool                          { return m.hasGit }
func (m *mockGitOperations) IsGitRepo() bool                       { return m.isGitRepo }
func (m *mockGitOperations) GetRepoPath() string                   { return m.repoPath }
func (m *mockGitOperations) GetCurrentBranch() string              { return m.currentBranch }
func (m *mockGitOperations) GetRemoteURL() string                  { return m.remoteURL }
func (m *mockGitOperations) Status() (string, error)               { return "mock status", nil }
func (m *mockGitOperations) Add(path string) (string, error)       { return "mock add", nil }
func (m *mockGitOperations) Commit(message string) (string, error) { return "mock commit", nil }
func (m *mockGitOperations) Push(branch string) (string, error)    { return "mock push", nil }
func (m *mockGitOperations) Pull(branch string) (string, error)    { return "mock pull", nil }
func (m *mockGitOperations) Checkout(branch string, create bool) (string, error) {
	return "mock checkout", nil
}
func (m *mockGitOperations) LogAnalysis(limit string) (string, error)     { return "mock log", nil }
func (m *mockGitOperations) DiffFiles(staged bool) (string, error)        { return "mock diff", nil }
func (m *mockGitOperations) Stash(operation, name string) (string, error) { return "mock stash", nil }
func (m *mockGitOperations) Remote(operation, name, url string) (string, error) {
	return "mock remote", nil
}
func (m *mockGitOperations) Tag(operation, tagName, message string) (string, error) {
	return "mock tag", nil
}
func (m *mockGitOperations) Clean(operation string, dryRun bool) (string, error) {
	return "mock clean", nil
}
func (m *mockGitOperations) GetLastCommit() (string, error) { return "mock last commit", nil }
func (m *mockGitOperations) GetFileContent(path, ref string) (string, error) {
	return "mock content", nil
}
func (m *mockGitOperations) ValidateRepo(path string) (string, error) { return "mock validate", nil }
func (m *mockGitOperations) ListFiles(ref string) (string, error)     { return "mock list files", nil }

func (m *mockGitOperations) CreateFile(path, content string) (string, error) {
	if m.createFileFunc != nil {
		return m.createFileFunc(path, content)
	}
	return "mock create file", nil
}
func (m *mockGitOperations) UpdateFile(path, content, sha string) (string, error) {
	if m.updateFileFunc != nil {
		return m.updateFileFunc(path, content, sha)
	}
	return "mock update file", nil
}
func (m *mockGitOperations) BranchList(remote bool) ([]types.BranchInfo, error) {
	if m.branchListFunc != nil {
		return m.branchListFunc(remote)
	}
	return []types.BranchInfo{{Name: "main"}}, nil
}
func (m *mockGitOperations) SetWorkspace(path string) (string, error) {
	if m.setWorkspaceFunc != nil {
		return m.setWorkspaceFunc(path)
	}
	return "mock set workspace", nil
}
func (m *mockGitOperations) GetFileSHA(path string) (string, error) {
	if m.getFileSHAFunc != nil {
		return m.getFileSHAFunc(path)
	}
	return "mock sha", nil
}
func (m *mockGitOperations) GetChangedFiles(staged bool) (string, error) {
	if m.GetChangedFilesFunc != nil {
		return m.GetChangedFilesFunc(staged)
	}
	return "mock changed files", nil
}

// mockGitHubOperations es una implementación simulada de interfaces.GitHubOperations
type mockGitHubOperations struct {
	createFileFunc func(ctx context.Context, owner, repo, path, content, message, branch string) (*github.RepositoryContentResponse, error)
	updateFileFunc func(ctx context.Context, owner, repo, path, content, message, sha, branch string) (*github.RepositoryContentResponse, error)
}

func (m *mockGitHubOperations) ListRepositories(ctx context.Context, listType string) ([]*github.Repository, error) {
	return nil, nil
}
func (m *mockGitHubOperations) CreateRepository(ctx context.Context, name, description string, private bool) (*github.Repository, error) {
	return nil, nil
}
func (m *mockGitHubOperations) ListPullRequests(ctx context.Context, owner, repo, state string) ([]*github.PullRequest, error) {
	return nil, nil
}
func (m *mockGitHubOperations) CreatePullRequest(ctx context.Context, owner, repo, title, head, base, body string) (*github.PullRequest, error) {
	return nil, nil
}
func (m *mockGitHubOperations) CreateFile(ctx context.Context, owner, repo, path, content, message, branch string) (*github.RepositoryContentResponse, error) {
	if m.createFileFunc != nil {
		return m.createFileFunc(ctx, owner, repo, path, content, message, branch)
	}
	return nil, errors.New("createFileFunc not implemented")
}
func (m *mockGitHubOperations) UpdateFile(ctx context.Context, owner, repo, path, content, message, sha, branch string) (*github.RepositoryContentResponse, error) {
	if m.updateFileFunc != nil {
		return m.updateFileFunc(ctx, owner, repo, path, content, message, sha, branch)
	}
	return nil, errors.New("updateFileFunc not implemented")
}

// --- Tests ---

func TestAutoDetectContext(t *testing.T) {
	t.Run("Git local detected", func(t *testing.T) {
		mockGit := &mockGitOperations{
			hasGit:        true,
			isGitRepo:     true,
			repoPath:      "/path/to/repo",
			currentBranch: "main",
			remoteURL:     "git@github.com:user/repo.git",
		}
		expected := fmt.Sprintf(`🔧 MODO GIT LOCAL DETECTADO (OPTIMIZACIÓN DE TOKENS)
📁 Repo: %s
🌿 Rama: %s
🔗 Remote: %s

✅ RECOMENDACIÓN: Usar comandos git_* para operaciones sin costo de tokens
- create_file/update_file: 0 tokens (Git local)
- git_add + git_commit: 0 tokens
- git_push: Solo si necesario sincronizar

❌ EVITAR: github_* APIs a menos que sea estrictamente necesario`, mockGit.repoPath, mockGit.currentBranch, mockGit.remoteURL)

		result := AutoDetectContext(mockGit)
		assert.Equal(t, expected, result)
	})

	t.Run("No Git local detected", func(t *testing.T) {
		mockGit := &mockGitOperations{
			hasGit:    false,
			isGitRepo: false,
		}
		expected := `⚠️ MODO GITHUB API (COSTO TOKENS)
❌ No se detectó Git local o repositorio Git
📡 Usando GitHub API (consume tokens)

💡 OPTIMIZACIÓN: Clona el repo localmente para reducir costos`

		result := AutoDetectContext(mockGit)
		assert.Equal(t, expected, result)
	})
}

func TestSmartCreateFile(t *testing.T) {
	args := map[string]interface{}{
		"path":    "test.txt",
		"content": "hello world",
		"owner":   "testuser",
		"repo":    "testrepo",
		"message": "feat: add test.txt",
	}

	t.Run("Success with Git local", func(t *testing.T) {
		mockGit := &mockGitOperations{
			hasGit:    true,
			isGitRepo: true,
			createFileFunc: func(path, content string) (string, error) {
				assert.Equal(t, "test.txt", path)
				assert.Equal(t, "hello world", content)
				return "Local file created", nil
			},
		}
		mockGithub := &mockGitHubOperations{}

		result, err := SmartCreateFile(mockGit, mockGithub, args)
		assert.NoError(t, err)
		assert.Contains(t, result, "✅ ARCHIVO CREADO CON GIT LOCAL")
		assert.Contains(t, result, "Local file created")
		assert.Contains(t, result, "git_add('test.txt') -> git_commit('feat: add test.txt')")
	})

	t.Run("Fallback to GitHub API when Git local fails", func(t *testing.T) {
		mockGit := &mockGitOperations{
			hasGit:    true,
			isGitRepo: true,
			createFileFunc: func(path, content string) (string, error) {
				return "", errors.New("local git error")
			},
		}
		mockGithub := &mockGitHubOperations{
			createFileFunc: func(ctx context.Context, owner, repo, path, content, message, branch string) (*github.RepositoryContentResponse, error) {
				resp := &github.RepositoryContentResponse{Commit: github.Commit{SHA: github.String("12345")}}
				return resp, nil
			},
		}

		result, err := SmartCreateFile(mockGit, mockGithub, args)
		assert.Error(t, err)
		assert.Equal(t, "git_local_failed", err.Error())
		assert.Contains(t, result, "⚠️ Git local falló: local git error")

		apiResult, apiErr := createFileWithAPI(mockGithub, args)
		assert.NoError(t, apiErr)
		assert.Contains(t, apiResult, "12345")
	})

	t.Run("Success with GitHub API (no local git)", func(t *testing.T) {
		mockGit := &mockGitOperations{hasGit: false}
		mockGithub := &mockGitHubOperations{
			createFileFunc: func(ctx context.Context, owner, repo, path, content, message, branch string) (*github.RepositoryContentResponse, error) {
				resp := &github.RepositoryContentResponse{Commit: github.Commit{SHA: github.String("abcdef")}}
				return resp, nil
			},
		}

		result, err := SmartCreateFile(mockGit, mockGithub, args)
		assert.NoError(t, err)
		assert.Contains(t, result, "📡 ARCHIVO CREADO CON GITHUB API")
		assert.Contains(t, result, "abcdef")
	})
}

func TestSmartUpdateFile(t *testing.T) {
	args := map[string]interface{}{
		"path":    "test.txt",
		"content": "new content",
		"owner":   "testuser",
		"repo":    "testrepo",
		"sha":     "file-sha",
		"message": "fix: update test.txt",
	}

	tmpDir, err := os.MkdirTemp("", "testrepo")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("Success with Git local", func(t *testing.T) {
		originalStat := stat
		stat = func(name string) (os.FileInfo, error) {
			return os.Stat(filepath.Join(tmpDir, "test.txt"))
		}
		defer func() { stat = originalStat }()

		err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("old content"), 0644)
		assert.NoError(t, err)

		mockGit := &mockGitOperations{
			hasGit:    true,
			isGitRepo: true,
			repoPath:  tmpDir,
			updateFileFunc: func(path, content, sha string) (string, error) {
				return "Local file updated", nil
			},
		}
		mockGithub := &mockGitHubOperations{}

		result, err := SmartUpdateFile(mockGit, mockGithub, args)
		assert.NoError(t, err)
		assert.Contains(t, result, "✅ ARCHIVO ACTUALIZADO CON GIT LOCAL")
	})

	t.Run("Fallback to GitHub API when local file does not exist", func(t *testing.T) {
		originalStat := stat
		stat = func(name string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}
		defer func() { stat = originalStat }()

		mockGit := &mockGitOperations{
			hasGit:    true,
			isGitRepo: true,
			repoPath:  tmpDir,
		}
		mockGithub := &mockGitHubOperations{
			updateFileFunc: func(ctx context.Context, owner, repo, path, content, message, sha, branch string) (*github.RepositoryContentResponse, error) {
				resp := &github.RepositoryContentResponse{Commit: github.Commit{SHA: github.String("67890")}}
				return resp, nil
			},
		}

		result, err := SmartUpdateFile(mockGit, mockGithub, args)
		assert.Error(t, err)
		assert.Equal(t, "git_local_failed", err.Error())
		assert.Contains(t, result, "⚠️ Archivo no existe localmente o Git local falló")

		apiResult, apiErr := updateFileWithAPI(mockGithub, args)
		assert.NoError(t, apiErr)
		assert.Contains(t, apiResult, "67890")
	})

	t.Run("Success with GitHub API (no local git)", func(t *testing.T) {
		mockGit := &mockGitOperations{hasGit: false}
		mockGithub := &mockGitHubOperations{
			updateFileFunc: func(ctx context.Context, owner, repo, path, content, message, sha, branch string) (*github.RepositoryContentResponse, error) {
				resp := &github.RepositoryContentResponse{Commit: github.Commit{SHA: github.String("uvwxyz")}}
				return resp, nil
			},
		}

		result, err := SmartUpdateFile(mockGit, mockGithub, args)
		assert.NoError(t, err)
		assert.Contains(t, result, "📡 ARCHIVO ACTUALIZADO CON GITHUB API")
		assert.Contains(t, result, "uvwxyz")
	})
}
