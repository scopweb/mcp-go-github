package git

import (
	"fmt"
	"os"
	exec_pkg "os/exec"
	"path/filepath"
	"strings"

	"github.com/jotajotape/github-go-server-mcp/pkg/interfaces"
	"github.com/jotajotape/github-go-server-mcp/pkg/types"
)

// executor define la interfaz para ejecutar commands.
type executor interface {
	Command(name string, arg ...string) cmdWrapper
	LookPath(file string) (string, error)
}

// cmdWrapper define la interfaz para un command.
type cmdWrapper interface {
	Output() ([]byte, error)
	CombinedOutput() ([]byte, error)
	SetDir(dir string)
}

// realCmd es la implementación real de la interfaz cmdWrapper.
type realCmd struct {
	*exec_pkg.Cmd
}

// SetDir establece el directorio de trabajo para el command.
func (c *realCmd) SetDir(dir string) {
	c.Dir = dir
}

// realExecutor es la implementación real de la interfaz executor.
type realExecutor struct{}

// Command crea un nuevo command para ejecutar.
func (e *realExecutor) Command(name string, arg ...string) cmdWrapper {
	return &realCmd{
		Cmd: exec_pkg.Command(name, arg...),
	}
}

// LookPath busca el ejecutable en el PATH del sistema.
func (e *realExecutor) LookPath(file string) (string, error) {
	return exec_pkg.LookPath(file)
}

// Client es el cliente para interactuar con Git.
type Client struct {
	Config   *types.GitConfig
	executor executor
}

// NewClientForTest crea un cliente con un ejecutor específico para pruebas.
func NewClientForTest(config *types.GitConfig, exec executor) interfaces.GitOperations {
	return &Client{
		Config:   config,
		executor: exec,
	}
}

// NewClient crea un nuevo cliente Git y detecta el entorno.
func NewClient() (interfaces.GitOperations, error) {
	exec := &realExecutor{}
	config := detectGitEnvironment(exec)

	// IMPORTANTE: No fallar si Git no está disponible
	// El servidor debe poder funcionar solo con GitHub API
	return &Client{
		Config:   &config,
		executor: exec,
	}, nil
}

// enterWorkingDir cambia al directorio del repositorio Git y retorna una función
// para restaurar el directorio original. Esto encapsula el patrón común de:
//   originalDir, _ := os.Getwd()
//   defer func() { _ = os.Chdir(originalDir) }()
//   if err := os.Chdir(c.Config.RepoPath); err != nil { ... }
func (c *Client) enterWorkingDir() (restore func(), err error) {
	originalDir, _ := os.Getwd()

	if err := os.Chdir(c.Config.RepoPath); err != nil {
		return nil, fmt.Errorf("error cambiando al directorio del repositorio: %w", err)
	}

	// Retorna función que restaura el directorio original
	return func() { _ = os.Chdir(originalDir) }, nil
}

// enterDir cambia a un directorio específico y retorna una función para restaurar
// el directorio original. Útil para operaciones que necesitan cambiar a un directorio
// arbitrario (no necesariamente c.Config.RepoPath).
func enterDir(dir string) (restore func(), err error) {
	originalDir, _ := os.Getwd()

	if err := os.Chdir(dir); err != nil {
		return nil, fmt.Errorf("error cambiando al directorio del repositorio: %w", err)
	}

	return func() { _ = os.Chdir(originalDir) }, nil
}

// detectGitEnvironment detecta y configura el entorno Git local.
func detectGitEnvironment(exec executor) types.GitConfig {
	config := types.GitConfig{}

	if _, err := exec.LookPath("git"); err == nil {
		config.HasGit = true
	}

	if !config.HasGit {
		return config
	}

	pwd, err := os.Getwd()
	if err != nil {
		return config
	}

	repoPath := findGitRepo(pwd)
	if repoPath == "" {
		return config
	}

	config.IsGitRepo = true
	config.RepoPath = repoPath

	originalDir, _ := os.Getwd()
	if err := os.Chdir(repoPath); err != nil {
		return config
	}
	defer func() { _ = os.Chdir(originalDir) }()

	if output, err := exec.Command("git", "remote", "get-url", "origin").Output(); err == nil {
		config.RemoteURL = strings.TrimSpace(string(output))
	}

	if output, err := exec.Command("git", "branch", "--show-current").Output(); err == nil {
		config.CurrentBranch = strings.TrimSpace(string(output))
	}

	return config
}

func findGitRepo(startPath string) string {
	currentPath := startPath
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return currentPath
		}

		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			break
		}
		currentPath = parentPath
	}
	return ""
}

func (c *Client) getEffectiveWorkingDir() string {
	if c.Config.WorkspacePath != "" {
		return c.Config.WorkspacePath
	}
	if c.Config.RepoPath != "" {
		return c.Config.RepoPath
	}
	pwd, _ := os.Getwd()
	return pwd
}
