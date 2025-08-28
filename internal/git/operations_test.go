package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jotajotape/github-go-server-mcp/internal/types"
)

// mockCmd is a mock for cmdWrapper
type mockCmd struct {
	output []byte
	err    error
}

func (c *mockCmd) Output() ([]byte, error) {
	return c.output, c.err
}

func (c *mockCmd) CombinedOutput() ([]byte, error) {
	return c.output, c.err
}

func (c *mockCmd) SetDir(dir string) {
	// No-op for mock. This method is required to satisfy the cmdWrapper interface.
}

// mockExecutor is a mock for the executor interface
type mockExecutor struct {
	t *testing.T
	// A map where keys are command strings (e.g., "git status --porcelain")
	// and values are the desired outputs.
	mockOutputs map[string]string
	mockErrors  map[string]error
}

func (e *mockExecutor) Command(name string, arg ...string) cmdWrapper {
	cmdStr := strings.TrimSpace(name + " " + strings.Join(arg, " "))

	mockOutput, okOutput := e.mockOutputs[cmdStr]
	mockError, okError := e.mockErrors[cmdStr]

	if !okOutput && !okError {
		e.t.Logf("No mock for command: '%s'", cmdStr)
		// Return a mock that produces no output and no error,
		// so the test fails on assertion rather than panic.
		return &mockCmd{output: []byte(""), err: nil}
	}

	return &mockCmd{
		output: []byte(mockOutput),
		err:    mockError,
	}
}

func (e *mockExecutor) LookPath(file string) (string, error) {
	if file == "git" {
		return "/fake/path/to/git", nil
	}
	return "", errors.New("not found")
}

// newTestClient is a helper to create a client with a mock executor.
func newTestClient(t *testing.T, config *types.GitConfig, mockOutputs map[string]string, mockErrors map[string]error) *Client {
	mockExec := &mockExecutor{
		t:           t,
		mockOutputs: mockOutputs,
		mockErrors:  mockErrors,
	}
	// We return a concrete *Client here for testing internal state if needed,
	// but it still satisfies the GitOperations interface.
	return &Client{
		Config:   config,
		executor: mockExec,
	}
}

// createTestRepo creates a temporary directory, initializes a git repository in it,
// and returns the path to the directory.
func createTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git user.email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git user.name: %v", err)
	}

	// Create an initial commit so the repo is not empty and has a HEAD
	err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("initial commit"), 0644)
	if err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}
	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add initial file: %v", err)
	}
	cmd = exec.Command("git", "commit", "-m", "initial commit")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	return dir
}

// TestHelperProcess isn't a real test. It's a helper that's executed by
// the mockExecutor. It prints the desired output to stdout/stderr and exits.
func TestHelperProcess(t *testing.T) {
	for cmdStr, output := range map[string]string{
		"git status --porcelain":         " M modified_file.go",
		"git log --oneline -5":           "abcde123 Fix stuff",
		"git remote":                     "origin\nother-remote",
		"git push origin feature-branch": "Everything up-to-date",
		"git push origin main":           "Everything up-to-date",
	} {
		cmd := exec.Command("sh", "-c", fmt.Sprintf("echo -n '%s'", output))
		cmd.Env = append(os.Environ(), "GIT_TEST_CMD="+cmdStr)
		_ = cmd.Start()
	}
}

func TestStatus(t *testing.T) {
	t.Run("In a valid git repo", func(t *testing.T) {
		repoPath := createTestRepo(t)
		config := &types.GitConfig{
			HasGit:        true,
			IsGitRepo:     true,
			RepoPath:      repoPath,
			RemoteURL:     "git@github.com:fake/repo.git",
			CurrentBranch: "main",
		}

		mockOutputs := map[string]string{
			"git status --porcelain": " M modified_file.go",
			"git log --oneline -5":   "abcde123 Fix stuff",
		}

		client := newTestClient(t, config, mockOutputs, nil)

		status, err := client.Status()
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		if !strings.Contains(status, `"status": "M modified_file.go"`) {
			t.Errorf("Expected status to contain modified file, but it didn't. Got:\n%s", status)
		}
		if !strings.Contains(status, `"recentCommits": "abcde123 Fix stuff"`) {
			t.Errorf("Expected status to contain recent commits, but it didn't. Got:\n%s", status)
		}
		if !strings.Contains(status, `"currentBranch": "main"`) {
			t.Errorf("Expected status to contain current branch, but it didn't. Got:\n%s", status)
		}
	})

	t.Run("Git not installed", func(t *testing.T) {
		config := &types.GitConfig{
			HasGit: false,
		}
		client := newTestClient(t, config, nil, nil)

		status, err := client.Status()
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		expectedMsg := "Git no está disponible en el sistema"
		if !strings.Contains(status, expectedMsg) {
			t.Errorf("Expected status to contain '%s', but it didn't. Got:\n%s", expectedMsg, status)
		}
	})

	t.Run("Not a git repo", func(t *testing.T) {
		config := &types.GitConfig{
			HasGit:    true,
			IsGitRepo: false,
		}
		client := newTestClient(t, config, nil, nil)

		status, err := client.Status()
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		expectedMsg := "No se detectó repositorio Git en el directorio actual"
		if !strings.Contains(status, expectedMsg) {
			t.Errorf("Expected status to contain '%s', but it didn't. Got:\n%s", expectedMsg, status)
		}
	})
}

func TestAdd(t *testing.T) {
	repoPath := createTestRepo(t)
	config := &types.GitConfig{
		HasGit:    true,
		IsGitRepo: true,
		RepoPath:  repoPath,
	}

	t.Run("Successful add", func(t *testing.T) {
		filesToAdd := "."
		mockOutputs := map[string]string{
			fmt.Sprintf("git add %s", filesToAdd): "", // git add doesn't produce output on success
		}
		client := newTestClient(t, config, mockOutputs, nil)

		result, err := client.Add(filesToAdd)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		expectedResult := fmt.Sprintf("Archivos '%s' añadidos al staging. Salida: ", filesToAdd)
		if !strings.HasPrefix(result, expectedResult) {
			t.Errorf("Expected result to start with '%s', but got '%s'", expectedResult, result)
		}
	})

	t.Run("Git command fails", func(t *testing.T) {
		filesToAdd := "."
		mockErrors := map[string]error{
			fmt.Sprintf("git add %s", filesToAdd): errors.New("some git error"),
		}
		client := newTestClient(t, config, nil, mockErrors)

		_, err := client.Add(filesToAdd)
		if err == nil {
			t.Fatal("Expected an error, but got none")
		}

		expectedErrorMsg := "error ejecutando git add"
		if !strings.Contains(err.Error(), expectedErrorMsg) {
			t.Errorf("Expected error message to contain '%s', but got '%s'", expectedErrorMsg, err.Error())
		}
	})
}

func TestCommit(t *testing.T) {
	repoPath := createTestRepo(t)
	config := types.GitConfig{
		RepoPath: repoPath,
	}
	var err error

	t.Run("Successful commit", func(t *testing.T) {
		commitMsg := "feat: add new feature"
		mockOutputs := map[string]string{
			fmt.Sprintf(`git commit -m %s`, commitMsg): "[main abcde123] feat: add new feature\n 1 file changed, 1 insertion(+)", // Simulando la salida de un commit exitoso
		}
		client := newTestClient(t, &config, mockOutputs, nil)

		result, err := client.Commit(commitMsg)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		if !strings.Contains(result, fmt.Sprintf("Commit realizado: %s", commitMsg)) {
			t.Errorf("Expected result to contain commit message, but it didn't. Got: %s", result)
		}
	})

	t.Run("Git command fails", func(t *testing.T) {
		commitMsg := "fail commit"
		mockErrors := map[string]error{
			fmt.Sprintf(`git commit -m %s`, commitMsg): errors.New("some git error"),
		}
		client := newTestClient(t, &config, nil, mockErrors)

		_, err = client.Commit(commitMsg)
		if err == nil {
			t.Fatal("Expected an error, but got none")
		}

		expectedErrorMsg := "error ejecutando git commit"
		if !strings.Contains(err.Error(), expectedErrorMsg) {
			t.Errorf("Expected error message to contain '%s', but got '%s'", expectedErrorMsg, err.Error())
		}
	})
}

func TestPush(t *testing.T) {
	repoPath := createTestRepo(t)
	config := &types.GitConfig{
		HasGit:        true,
		IsGitRepo:     true,
		RepoPath:      repoPath,
		CurrentBranch: "main",
	}

	t.Run("Successful push to a specific branch", func(t *testing.T) {
		branch := "feature-branch"
		remote := "origin"
		mockOutputs := map[string]string{
			"git remote": fmt.Sprintf("%s\nother-remote", remote),
			fmt.Sprintf("git push %s %s", remote, branch): "Everything up-to-date",
		}
		client := newTestClient(t, config, mockOutputs, nil)

		result, err := client.Push(branch)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		if !strings.Contains(result, fmt.Sprintf("Push a '%s' en la rama '%s' realizado con éxito", remote, branch)) {
			t.Errorf("Expected result to contain push confirmation for branch %s, but it didn't. Got: %s", branch, result)
		}
	})

	t.Run("Successful push to default branch", func(t *testing.T) {
		remote := "origin"
		mockOutputs := map[string]string{
			"git remote": remote,
			fmt.Sprintf("git push %s %s", remote, config.CurrentBranch): "Everything up-to-date",
		}
		client := newTestClient(t, config, mockOutputs, nil)

		result, err := client.Push("") // Empty string should use default
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		if !strings.Contains(result, fmt.Sprintf("Push a '%s' en la rama '%s' realizado con éxito", remote, config.CurrentBranch)) {
			t.Errorf("Expected result to contain push confirmation for default branch, but it didn't. Got: %s", result)
		}
	})

	t.Run("Git command fails", func(t *testing.T) {
		branch := "main"
		remote := "origin"
		mockOutputs := map[string]string{
			"git remote": remote,
		}
		mockErrors := map[string]error{
			fmt.Sprintf("git push %s %s", remote, branch): errors.New("some git push error"),
		}
		client := newTestClient(t, config, mockOutputs, mockErrors)

		_, err := client.Push(branch)
		if err == nil {
			t.Fatal("Expected an error, but got none")
		}

		expectedErrorMsg := "error ejecutando git push"
		if !strings.Contains(err.Error(), expectedErrorMsg) {
			t.Errorf("Expected error message to contain '%s', but got '%s'", expectedErrorMsg, err.Error())
		}
	})
}

func TestPull(t *testing.T) {
	repoPath := createTestRepo(t)
	config := &types.GitConfig{
		HasGit:        true,
		IsGitRepo:     true,
		RepoPath:      repoPath,
		CurrentBranch: "main",
	}

	t.Run("Successful pull from a specific branch", func(t *testing.T) {
		branch := "feature-branch"
		mockOutputs := map[string]string{
			fmt.Sprintf("git pull origin -- %s", branch): "Already up to date.",
		}
		client := newTestClient(t, config, mockOutputs, nil)

		result, err := client.Pull(branch)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		if !strings.Contains(result, fmt.Sprintf("Pull realizado desde rama: %s", branch)) {
			t.Errorf("Expected result to contain pull confirmation for branch %s, but it didn't. Got: %s", branch, result)
		}
	})

	t.Run("Successful pull from default branch", func(t *testing.T) {
		mockOutputs := map[string]string{
			fmt.Sprintf("git pull origin -- %s", config.CurrentBranch): "Already up to date.",
		}
		client := newTestClient(t, config, mockOutputs, nil)

		result, err := client.Pull("") // Empty string should use default
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		if !strings.Contains(result, fmt.Sprintf("Pull realizado desde rama: %s", config.CurrentBranch)) {
			t.Errorf("Expected result to contain pull confirmation for default branch, but it didn't. Got: %s", result)
		}
	})

	t.Run("Git command fails", func(t *testing.T) {
		branch := "main"
		mockErrors := map[string]error{
			fmt.Sprintf("git pull origin -- %s", branch): errors.New("some git pull error"),
		}
		client := newTestClient(t, config, nil, mockErrors)

		_, err := client.Pull(branch)
		if err == nil {
			t.Fatal("Expected an error, but got none")
		}

		expectedErrorMsg := "error ejecutando git pull"
		if !strings.Contains(err.Error(), expectedErrorMsg) {
			t.Errorf("Expected error message to contain '%s', but got '%s'", expectedErrorMsg, err.Error())
		}
	})
}

func TestCheckout(t *testing.T) {
	repoPath := createTestRepo(t)
	config := &types.GitConfig{
		HasGit:        true,
		IsGitRepo:     true,
		RepoPath:      repoPath,
		CurrentBranch: "main",
	}

	t.Run("Successful checkout to a new branch", func(t *testing.T) {
		branch := "new-feature-branch"
		mockOutputs := map[string]string{
			fmt.Sprintf("git checkout -b -- %s", branch): fmt.Sprintf("Switched to a new branch '%s'", branch),
		}
		client := newTestClient(t, config, mockOutputs, nil)

		result, err := client.Checkout(branch, true)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		if !strings.Contains(result, fmt.Sprintf("Checkout a rama: %s (crear: true)", branch)) {
			t.Errorf("Expected result to contain checkout confirmation for new branch %s, but it didn't. Got: %s", branch, result)
		}
		if config.CurrentBranch != branch {
			t.Errorf("Expected CurrentBranch to be updated to '%s', but got '%s'", branch, config.CurrentBranch)
		}
	})

	t.Run("Successful checkout to an existing branch", func(t *testing.T) {
		// First, create the branch in the test repo so checkout doesn't fail
		existingBranch := "existing-branch"
		cmd := exec.Command("git", "branch", existingBranch)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create existing branch: %v", err)
		}

		mockOutputs := map[string]string{
			fmt.Sprintf("git checkout -- %s", existingBranch): fmt.Sprintf("Switched to branch '%s'", existingBranch),
		}
		client := newTestClient(t, config, mockOutputs, nil)

		result, err := client.Checkout(existingBranch, false)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		if !strings.Contains(result, fmt.Sprintf("Checkout a rama: %s (crear: false)", existingBranch)) {
			t.Errorf("Expected result to contain checkout confirmation for existing branch %s, but it didn't. Got: %s", existingBranch, result)
		}
	})

	t.Run("Git command fails", func(t *testing.T) {
		branch := "non-existent-branch"
		mockErrors := map[string]error{
			fmt.Sprintf("git checkout -- %s", branch): errors.New("some git checkout error"),
		}
		client := newTestClient(t, config, nil, mockErrors)

		_, err := client.Checkout(branch, false)
		if err == nil {
			t.Fatal("Expected an error, but got none")
		}

		expectedErrorMsg := "error ejecutando git checkout"
		if !strings.Contains(err.Error(), expectedErrorMsg) {
			t.Errorf("Expected error message to contain '%s', but got '%s'", expectedErrorMsg, err.Error())
		}
	})
}

func TestCreateFile(t *testing.T) {
	repoPath := createTestRepo(t)
	config := &types.GitConfig{
		HasGit:    true,
		IsGitRepo: true,
		RepoPath:  repoPath,
	}
	client := newTestClient(t, config, nil, nil)

	t.Run("Successful file creation in root", func(t *testing.T) {
		filePath := "new_file.txt"
		content := "Hello, World!"
		fullPath := filepath.Join(repoPath, filePath)

		result, err := client.CreateFile(filePath, content)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		expectedResult := fmt.Sprintf("Archivo creado: %s. Usa 'git_add' y 'git_commit' para confirmar cambios", filePath)
		if result != expectedResult {
			t.Errorf("Expected result '%s', but got '%s'", expectedResult, result)
		}

		// Verify the file was actually created with the correct content
		fileContent, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}
		if string(fileContent) != content {
			t.Errorf("Expected file content '%s', but got '%s'", content, string(fileContent))
		}
	})

	t.Run("Successful file creation in subdirectory", func(t *testing.T) {
		filePath := "new/dir/another_file.txt"
		content := "Subdirectory content"
		fullPath := filepath.Join(repoPath, filePath)

		result, err := client.CreateFile(filePath, content)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		expectedResult := fmt.Sprintf("Archivo creado: %s. Usa 'git_add' y 'git_commit' para confirmar cambios", filePath)
		if result != expectedResult {
			t.Errorf("Expected result '%s', but got '%s'", expectedResult, result)
		}

		// Verify the file was actually created with the correct content
		fileContent, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}
		if string(fileContent) != content {
			t.Errorf("Expected file content '%s', but got '%s'", content, string(fileContent))
		}
	})
}

func TestUpdateFile(t *testing.T) {
	repoPath := createTestRepo(t)
	config := &types.GitConfig{
		HasGit:    true,
		IsGitRepo: true,
		RepoPath:  repoPath,
	}
	client := newTestClient(t, config, nil, nil)

	t.Run("Successful file update", func(t *testing.T) {
		filePath := "file_to_update.txt"
		initialContent := "Initial content"
		updatedContent := "Updated content"
		fullPath := filepath.Join(repoPath, filePath)

		// Create the file first
		err := os.WriteFile(fullPath, []byte(initialContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create initial file for update test: %v", err)
		}

		result, err := client.UpdateFile(filePath, updatedContent, "")
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		expectedResult := fmt.Sprintf("Archivo actualizado: %s. Usa 'git_add' y 'git_commit' para confirmar cambios", filePath)
		if result != expectedResult {
			t.Errorf("Expected result '%s', but got '%s'", expectedResult, result)
		}

		// Verify the file was actually updated with the correct content
		fileContent, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("Failed to read updated file: %v", err)
		}
		if string(fileContent) != updatedContent {
			t.Errorf("Expected file content '%s', but got '%s'", updatedContent, string(fileContent))
		}
	})
}
