package hybrid

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v81/github"
	"github.com/scopweb/mcp-go-github/pkg/types"
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
	addFunc             func(path string) (string, error)
	commitFunc          func(message string) (string, error)
	pushFunc            func(branch string) (string, error)
	branchListFunc      func(remote bool) ([]types.BranchInfo, error)
	setWorkspaceFunc    func(path string) (string, error)
	getFileSHAFunc      func(path string) (string, error)
	GetChangedFilesFunc func(staged bool) (string, error)
	initFunc            func(path string, branch string) (string, error)
}

func (m *mockGitOperations) HasGit() bool             { return m.hasGit }
func (m *mockGitOperations) IsGitRepo() bool          { return m.isGitRepo }
func (m *mockGitOperations) GetRepoPath() string      { return m.repoPath }
func (m *mockGitOperations) GetCurrentBranch() string { return m.currentBranch }
func (m *mockGitOperations) GetRemoteURL() string     { return m.remoteURL }
func (m *mockGitOperations) Status() (string, error)  { return "mock status", nil }
func (m *mockGitOperations) Add(path string) (string, error) {
	if m.addFunc != nil {
		return m.addFunc(path)
	}
	return "mock add", nil
}
func (m *mockGitOperations) Commit(message string) (string, error) {
	if m.commitFunc != nil {
		return m.commitFunc(message)
	}
	return "mock commit", nil
}
func (m *mockGitOperations) Push(branch string) (string, error) {
	if m.pushFunc != nil {
		return m.pushFunc(branch)
	}
	return "mock push", nil
}
func (m *mockGitOperations) Pull(_ string) (string, error) { return "mock pull", nil }
func (m *mockGitOperations) Checkout(_ string, _ bool) (string, error) {
	return "mock checkout", nil
}
func (m *mockGitOperations) LogAnalysis(_ string) (string, error) { return "mock log", nil }
func (m *mockGitOperations) DiffFiles(_ bool) (string, error)     { return "mock diff", nil }
func (m *mockGitOperations) Stash(_, _ string) (string, error)    { return "mock stash", nil }
func (m *mockGitOperations) Remote(_, _, _ string) (string, error) {
	return "mock remote", nil
}
func (m *mockGitOperations) Tag(_, _, _ string) (string, error) {
	return "mock tag", nil
}
func (m *mockGitOperations) Clean(_ string, _ bool) (string, error) {
	return "mock clean", nil
}
func (m *mockGitOperations) GetLastCommit() (string, error) { return "mock last commit", nil }
func (m *mockGitOperations) GetFileContent(_, _ string) (string, error) {
	return "mock content", nil
}
func (m *mockGitOperations) ValidateRepo(_ string) (string, error) { return "mock validate", nil }
func (m *mockGitOperations) ListFiles(_ string) (string, error)    { return "mock list files", nil }

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

// Repository initialization
func (m *mockGitOperations) Init(path string, initialBranch string) (string, error) {
	if m.initFunc != nil {
		return m.initFunc(path, initialBranch)
	}
	return "mock init", nil
}

// Advanced branch operations
func (m *mockGitOperations) CheckoutRemote(_ string, _ string) (string, error) {
	return "mock checkout remote", nil
}
func (m *mockGitOperations) Merge(_ string, _ string) (string, error) {
	return "mock merge", nil
}
func (m *mockGitOperations) Rebase(_ string) (string, error) {
	return "mock rebase", nil
}

// Enhanced pull/push operations
func (m *mockGitOperations) PullWithStrategy(_ string, _ string) (string, error) {
	return "mock pull with strategy", nil
}
func (m *mockGitOperations) ForcePush(_ string, _ bool) (string, error) {
	return "mock force push", nil
}
func (m *mockGitOperations) PushUpstream(_ string) (string, error) {
	return "mock push upstream", nil
}

// Batch operations
func (m *mockGitOperations) SyncWithRemote(_ string) (string, error) {
	return "mock sync with remote", nil
}
func (m *mockGitOperations) SafeMerge(_ string, _ string) (string, error) {
	return "mock safe merge", nil
}

// Conflict management
func (m *mockGitOperations) ConflictStatus() (string, error) {
	return "mock conflict status", nil
}
func (m *mockGitOperations) ResolveConflicts(_ string) (string, error) {
	return "mock resolve conflicts", nil
}

// Validation operations
func (m *mockGitOperations) ValidateCleanState() (bool, error) {
	return true, nil
}
func (m *mockGitOperations) DetectPotentialConflicts(_ string, _ string) (string, error) {
	return "mock detect conflicts", nil
}
func (m *mockGitOperations) CreateBackup(_ string) (string, error) {
	return "mock create backup", nil
}

// Phase 1: Essential commands
func (m *mockGitOperations) Reset(_ string, _ string, _ []string) (string, error) {
	return "mock reset", nil
}

// Phase 2: Conflict management
func (m *mockGitOperations) ShowConflict(_ string) (string, error) {
	return "mock show conflict", nil
}
func (m *mockGitOperations) ResolveFile(_ string, _ string, _ *string) (string, error) {
	return "mock resolve file", nil
}

// mockGitHubOperations es una implementación simulada de interfaces.GitHubOperations
type mockGitHubOperations struct {
	createFileFunc func(ctx context.Context, owner, repo, path, content, message, branch string) (*github.RepositoryContentResponse, error)
	updateFileFunc func(ctx context.Context, owner, repo, path, content, message, sha, branch string) (*github.RepositoryContentResponse, error)
}

func (m *mockGitHubOperations) ListRepositories(_ context.Context, _ string) ([]*github.Repository, error) {
	return nil, nil
}
func (m *mockGitHubOperations) CreateRepository(_ context.Context, _ string, _ string, _ bool) (*github.Repository, error) {
	return nil, nil
}
func (m *mockGitHubOperations) ListPullRequests(_ context.Context, _ string, _ string, _ string) ([]*github.PullRequest, error) {
	return nil, nil
}
func (m *mockGitHubOperations) CreatePullRequest(_ context.Context, _ string, _ string, _ string, _ string, _ string, _ string) (*github.PullRequest, error) {
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

// New methods for v2.1
func (m *mockGitHubOperations) CreateIssueComment(_ context.Context, _ string, _ string, _ int, _ string) (*github.IssueComment, error) {
	return nil, nil
}
func (m *mockGitHubOperations) CloseIssue(_ context.Context, _ string, _ string, _ int, _ string) (*github.Issue, error) {
	return nil, nil
}
func (m *mockGitHubOperations) CreatePRComment(_ context.Context, _ string, _ string, _ int, _ string) (*github.IssueComment, error) {
	return nil, nil
}
func (m *mockGitHubOperations) CreatePRReview(_ context.Context, _ string, _ string, _ int, _ string, _ string) (*github.PullRequestReview, error) {
	return nil, nil
}
func (m *mockGitHubOperations) MergePullRequest(_ context.Context, _ string, _ string, _ int, _ string, _ string) (*github.PullRequestMergeResult, error) {
	return nil, nil
}
func (m *mockGitHubOperations) RerunWorkflow(_ context.Context, _ string, _ string, _ int64) error {
	return nil
}
func (m *mockGitHubOperations) RerunFailedJobs(_ context.Context, _ string, _ string, _ int64) error {
	return nil
}
func (m *mockGitHubOperations) DismissDependabotAlert(_ context.Context, _ string, _ string, _ int, _ string, _ string) (*github.DependabotAlert, error) {
	return nil, nil
}
func (m *mockGitHubOperations) DismissCodeScanningAlert(_ context.Context, _ string, _ string, _ int64, _ string, _ string) (*github.Alert, error) {
	return nil, nil
}
func (m *mockGitHubOperations) DismissSecretScanningAlert(_ context.Context, _ string, _ string, _ int64, _ string) (*github.SecretScanningAlert, error) {
	return nil, nil
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

✅ RECOMENDACIÓN: Usar commands git_* para operaciones sin costo de tokens
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
			createFileFunc: func(path string, content string) (string, error) {
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
			createFileFunc: func(_ string, _ string) (string, error) {
				return "", errors.New("local git error")
			},
		}
		mockGithub := &mockGitHubOperations{
			createFileFunc: func(_ context.Context, _ string, _ string, _ string, _ string, _ string, _ string) (*github.RepositoryContentResponse, error) {
				resp := &github.RepositoryContentResponse{Commit: github.Commit{SHA: github.Ptr("12345")}}
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
			createFileFunc: func(_ context.Context, _ string, _ string, _ string, _ string, _ string, _ string) (*github.RepositoryContentResponse, error) {
				resp := &github.RepositoryContentResponse{Commit: github.Commit{SHA: github.Ptr("abcdef")}}
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
		stat = func(_ string) (os.FileInfo, error) {
			return os.Stat(filepath.Join(tmpDir, "test.txt"))
		}
		defer func() { stat = originalStat }()

		err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("old content"), 0600)
		assert.NoError(t, err)

		mockGit := &mockGitOperations{
			hasGit:    true,
			isGitRepo: true,
			repoPath:  tmpDir,
			updateFileFunc: func(_ string, _ string, _ string) (string, error) {
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
		stat = func(_ string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}
		defer func() { stat = originalStat }()

		mockGit := &mockGitOperations{
			hasGit:    true,
			isGitRepo: true,
			repoPath:  tmpDir,
		}
		mockGithub := &mockGitHubOperations{
			updateFileFunc: func(_ context.Context, _ string, _ string, _ string, _ string, _ string, _ string, _ string) (*github.RepositoryContentResponse, error) {
				resp := &github.RepositoryContentResponse{Commit: github.Commit{SHA: github.Ptr("67890")}}
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
			updateFileFunc: func(_ context.Context, _ string, _ string, _ string, _ string, _ string, _ string, _ string) (*github.RepositoryContentResponse, error) {
				resp := &github.RepositoryContentResponse{Commit: github.Commit{SHA: github.Ptr("uvwxyz")}}
				return resp, nil
			},
		}

		result, err := SmartUpdateFile(mockGit, mockGithub, args)
		assert.NoError(t, err)
		assert.Contains(t, result, "📡 ARCHIVO ACTUALIZADO CON GITHUB API")
		assert.Contains(t, result, "uvwxyz")
	})
}

func TestPushFiles(t *testing.T) {
	t.Run("Batch write, commit and push", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "pushfiles")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		existingPath := filepath.Join(tmpDir, "existing.txt")
		assert.NoError(t, os.MkdirAll(filepath.Dir(existingPath), 0755))
		assert.NoError(t, os.WriteFile(existingPath, []byte("old"), 0600))

		mockGit := &mockGitOperations{
			hasGit:        true,
			isGitRepo:     true,
			repoPath:      tmpDir,
			currentBranch: "main",
			createFileFunc: func(path, content string) (string, error) {
				assert.Equal(t, "new/new.txt", path)
				assert.Equal(t, "new content", content)
				return "created", nil
			},
			updateFileFunc: func(path, content, _ string) (string, error) {
				assert.Equal(t, "existing.txt", path)
				assert.Equal(t, "updated content", content)
				return "updated", nil
			},
			addFunc: func(path string) (string, error) {
				assert.Equal(t, "-A", path)
				return "added", nil
			},
			commitFunc: func(message string) (string, error) {
				assert.Equal(t, "feat: batch push", message)
				return "committed", nil
			},
			pushFunc: func(branch string) (string, error) {
				assert.Equal(t, "main", branch)
				return "pushed", nil
			},
		}

		args := map[string]interface{}{
			"files": []interface{}{
				map[string]interface{}{"path": "new/new.txt", "content": "new content"},
				map[string]interface{}{"path": "existing.txt", "content": "updated content"},
			},
			"message": "feat: batch push",
		}

		result, err := PushFiles(mockGit, args)
		assert.NoError(t, err)
		assert.Contains(t, result, "Creados: new/new.txt")
		assert.Contains(t, result, "Actualizados: existing.txt")
		assert.Contains(t, result, "git push (main)")
	})

	t.Run("Fails when Git is not available", func(t *testing.T) {
		mockGit := &mockGitOperations{}
		args := map[string]interface{}{
			"files":   []interface{}{map[string]interface{}{"path": "a.txt", "content": "x"}},
			"message": "test",
		}

		_, err := PushFiles(mockGit, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "git no disponible")
	})

	t.Run("Validates files payload", func(t *testing.T) {
		mockGit := &mockGitOperations{hasGit: true, isGitRepo: true, repoPath: "/tmp"}
		args := map[string]interface{}{
			"files":   []interface{}{},
			"message": "test",
		}

		_, err := PushFiles(mockGit, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "se requiere 'files' o 'paths'")
	})

	t.Run("source_path reads file from disk", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "pushfiles-src")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create source file on disk
		srcFile := filepath.Join(tmpDir, "source.php")
		assert.NoError(t, os.WriteFile(srcFile, []byte("<?php echo 'hello'; ?>"), 0600))

		repoDir, err := os.MkdirTemp("", "pushfiles-repo")
		assert.NoError(t, err)
		defer os.RemoveAll(repoDir)

		var capturedContent string
		mockGit := &mockGitOperations{
			hasGit:        true,
			isGitRepo:     true,
			repoPath:      repoDir,
			currentBranch: "main",
			createFileFunc: func(path, content string) (string, error) {
				capturedContent = content
				return "created", nil
			},
			addFunc:    func(_ string) (string, error) { return "added", nil },
			commitFunc: func(_ string) (string, error) { return "committed", nil },
			pushFunc:   func(_ string) (string, error) { return "pushed", nil },
		}

		args := map[string]interface{}{
			"files": []interface{}{
				map[string]interface{}{
					"path":        "includes/plugin.php",
					"source_path": srcFile,
				},
			},
			"message": "feat: add plugin from source_path",
		}

		result, err := PushFiles(mockGit, args)
		assert.NoError(t, err)
		assert.Equal(t, "<?php echo 'hello'; ?>", capturedContent)
		assert.Contains(t, result, "Creados: includes/plugin.php")
	})

	t.Run("source_path error when file not found", func(t *testing.T) {
		mockGit := &mockGitOperations{
			hasGit:        true,
			isGitRepo:     true,
			repoPath:      "/tmp",
			currentBranch: "main",
		}

		args := map[string]interface{}{
			"files": []interface{}{
				map[string]interface{}{
					"path":        "dest.txt",
					"source_path": "/nonexistent/file.txt",
				},
			},
			"message": "test",
		}

		_, err := PushFiles(mockGit, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source_path")
	})

	t.Run("paths mode stages existing files", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "pushfiles-paths")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create files in the workspace
		assert.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "includes"), 0755))
		assert.NoError(t, os.WriteFile(filepath.Join(tmpDir, "includes", "main.php"), []byte("php"), 0600))
		assert.NoError(t, os.WriteFile(filepath.Join(tmpDir, "style.css"), []byte("css"), 0600))

		var addedPaths string
		mockGit := &mockGitOperations{
			hasGit:        true,
			isGitRepo:     true,
			repoPath:      tmpDir,
			currentBranch: "main",
			addFunc: func(path string) (string, error) {
				addedPaths = path
				return "added", nil
			},
			commitFunc: func(_ string) (string, error) { return "committed", nil },
			pushFunc:   func(_ string) (string, error) { return "pushed", nil },
		}

		args := map[string]interface{}{
			"paths":   []interface{}{"includes/main.php", "style.css"},
			"message": "feat: stage existing files",
		}

		result, err := PushFiles(mockGit, args)
		assert.NoError(t, err)
		assert.Contains(t, addedPaths, "includes/main.php")
		assert.Contains(t, addedPaths, "style.css")
		assert.Contains(t, result, "Staged: includes/main.php, style.css")
		assert.Contains(t, result, "2 archivo(s) procesados")
	})

	t.Run("paths mode error when file not in workspace", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "pushfiles-paths-err")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		mockGit := &mockGitOperations{
			hasGit:        true,
			isGitRepo:     true,
			repoPath:      tmpDir,
			currentBranch: "main",
		}

		args := map[string]interface{}{
			"paths":   []interface{}{"nonexistent.txt"},
			"message": "test",
		}

		_, err = PushFiles(mockGit, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no existe en el workspace")
	})

	t.Run("files and paths combined", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "pushfiles-combined")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Existing file for paths mode
		assert.NoError(t, os.WriteFile(filepath.Join(tmpDir, "existing.css"), []byte("css"), 0600))

		mockGit := &mockGitOperations{
			hasGit:        true,
			isGitRepo:     true,
			repoPath:      tmpDir,
			currentBranch: "main",
			createFileFunc: func(path, content string) (string, error) {
				return "created", nil
			},
			addFunc:    func(_ string) (string, error) { return "added", nil },
			commitFunc: func(_ string) (string, error) { return "committed", nil },
			pushFunc:   func(_ string) (string, error) { return "pushed", nil },
		}

		args := map[string]interface{}{
			"files": []interface{}{
				map[string]interface{}{"path": "new.txt", "content": "hello"},
			},
			"paths":   []interface{}{"existing.css"},
			"message": "feat: combined mode",
		}

		result, err := PushFiles(mockGit, args)
		assert.NoError(t, err)
		assert.Contains(t, result, "Creados: new.txt")
		assert.Contains(t, result, "Staged: existing.css")
		assert.Contains(t, result, "2 archivo(s) procesados")
	})

	t.Run("file requires content or source_path", func(t *testing.T) {
		mockGit := &mockGitOperations{
			hasGit:        true,
			isGitRepo:     true,
			repoPath:      "/tmp",
			currentBranch: "main",
		}

		args := map[string]interface{}{
			"files": []interface{}{
				map[string]interface{}{"path": "test.txt"},
			},
			"message": "test",
		}

		_, err := PushFiles(mockGit, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "'content' o 'source_path'")
	})
}
