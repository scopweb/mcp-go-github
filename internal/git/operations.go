package git

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	exec_pkg "os/exec"
	"path/filepath"
	"strings"

	"github.com/scopweb/mcp-go-github/internal/interfaces"
	"github.com/scopweb/mcp-go-github/internal/types"
)

// executor define la interfaz para ejecutar comandos.
type executor interface {
	Command(name string, arg ...string) cmdWrapper
	LookPath(file string) (string, error)
}

// cmdWrapper define la interfaz para un comando.
type cmdWrapper interface {
	Output() ([]byte, error)
	CombinedOutput() ([]byte, error)
	SetDir(dir string)
}

// realCmd es la implementación real de la interfaz cmdWrapper.
type realCmd struct {
	*exec_pkg.Cmd
}

// SetDir establece el directorio de trabajo para el comando.
func (c *realCmd) SetDir(dir string) {
	c.Dir = dir
}

// realExecutor es la implementación real de la interfaz executor.
type realExecutor struct{}

// Command crea un nuevo comando para ejecutar.
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
	defer os.Chdir(originalDir)
	os.Chdir(repoPath)

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

// Implementación de los métodos de la interfaz GitOperations

func (c *Client) Status() (string, error) {
	result := map[string]interface{}{
		"gitConfig": c.Config,
	}

	if !c.Config.HasGit {
		result["message"] = "Git no está disponible en el sistema"
		output, _ := json.MarshalIndent(result, "", "  ")
		return string(output), nil
	}

	if !c.Config.IsGitRepo {
		result["message"] = "No se detectó repositorio Git en el directorio actual"
		output, _ := json.MarshalIndent(result, "", "  ")
		return string(output), nil
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	if output, err := c.executor.Command("git", "status", "--porcelain").Output(); err == nil {
		result["status"] = strings.TrimSpace(string(output))
	}

	if output, err := c.executor.Command("git", "log", "--oneline", "-5").Output(); err == nil {
		result["recentCommits"] = strings.TrimSpace(string(output))
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}

// Add añade archivos al staging area.
func (c *Client) Add(files string) (string, error) {
	cmd := c.executor.Command("git", "add", files)
	cmd.SetDir(c.Config.RepoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error ejecutando git add: %s", string(output))
	}
	return fmt.Sprintf("Archivos '%s' añadidos al staging. Salida: %s", files, string(output)), nil
}

// Commit realiza un commit con el mensaje proporcionado.
func (c *Client) Commit(message string) (string, error) {
	cmd := c.executor.Command("git", "commit", "-m", message)
	cmd.SetDir(c.Config.RepoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error ejecutando git commit: %s", string(output))
	}
	return fmt.Sprintf("Commit realizado: %s. Salida: %s", message, string(output)), nil
}

// Push realiza un push a la rama especificada.
func (c *Client) Push(branch string) (string, error) {
	if branch == "" {
		branch = c.Config.CurrentBranch
	}
	// Obtener el nombre del control remoto
	cmdRemote := c.executor.Command("git", "remote")
	cmdRemote.SetDir(c.Config.RepoPath)
	remoteOutput, err := cmdRemote.Output()
	if err != nil {
		return "", fmt.Errorf("error al obtener remotos de Git: %w", err)
	}
	remotes := strings.Fields(string(remoteOutput))
	if len(remotes) == 0 {
		return "", errors.New("no se encontraron remotos de Git")
	}
	remote := remotes[0] // Usar el primer remoto encontrado

	// Ejecutar git push
	cmdPush := c.executor.Command("git", "push", remote, branch)
	cmdPush.SetDir(c.Config.RepoPath)
	output, err := cmdPush.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error ejecutando git push: %s", string(output))
	}

	return fmt.Sprintf("Push a '%s' en la rama '%s' realizado con éxito. Salida: %s", remote, branch, string(output)), nil
}

func (c *Client) Pull(branch string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	effectiveBranch := branch
	if effectiveBranch == "" {
		effectiveBranch = c.Config.CurrentBranch
	}

	cmd := c.executor.Command("git", "pull", "origin", "--", effectiveBranch)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error ejecutando git pull: %v, Output: %s", err, output)
	}

	return fmt.Sprintf("Pull realizado desde rama: %s", effectiveBranch), nil
}

func (c *Client) Checkout(branch string, create bool) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	var cmd cmdWrapper
	if create {
		cmd = c.executor.Command("git", "checkout", "-b", "--", branch)
	} else {
		cmd = c.executor.Command("git", "checkout", "--", branch)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error ejecutando git checkout: %v, Output: %s", err, output)
	}

	c.Config.CurrentBranch = branch
	return fmt.Sprintf("Checkout a rama: %s (crear: %v)", branch, create), nil
}

func (c *Client) CreateFile(path, content string) (string, error) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	fullPath := filepath.Join(c.Config.RepoPath, path)
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	if err != nil {
		return "", fmt.Errorf("error creando directorio: %v", err)
	}

	err = os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("error escribiendo archivo: %v", err)
	}

	return fmt.Sprintf("Archivo creado: %s. Usa 'git_add' y 'git_commit' para confirmar cambios", path), nil
}

func (c *Client) UpdateFile(path, content, sha string) (string, error) {
	workingDir := c.getEffectiveWorkingDir()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	fullPath := filepath.Join(workingDir, path)
	err := os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("error actualizando archivo: %v", err)
	}

	return fmt.Sprintf("Archivo actualizado: %s. Usa 'git_add' y 'git_commit' para confirmar cambios", path), nil
}

func (c *Client) SetWorkspace(workspacePath string) (string, error) {
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return "", fmt.Errorf("directorio no existe: %s", workspacePath)
	}

	gitPath := filepath.Join(workspacePath, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return "", fmt.Errorf("no es un repositorio Git: %s", workspacePath)
	}

	if _, err := c.executor.LookPath("git"); err != nil {
		return "", fmt.Errorf("git no está disponible en el sistema")
	}

	c.Config.WorkspacePath = workspacePath
	c.Config.IsGitRepo = true
	c.Config.RepoPath = workspacePath
	c.Config.HasGit = true

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workspacePath)

	if output, err := c.executor.Command("git", "remote", "get-url", "origin").Output(); err == nil {
		c.Config.RemoteURL = strings.TrimSpace(string(output))
	}

	if output, err := c.executor.Command("git", "branch", "--show-current").Output(); err == nil {
		c.Config.CurrentBranch = strings.TrimSpace(string(output))
	}

	return fmt.Sprintf("Workspace configurado: %s, Rama: %s, Remote: %s",
		workspacePath, c.Config.CurrentBranch, c.Config.RemoteURL), nil
}

func (c *Client) GetFileSHA(filePath string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	workingDir := c.getEffectiveWorkingDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	cmd := c.executor.Command("git", "rev-parse", fmt.Sprintf("HEAD:%s", filePath))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error obteniendo SHA del archivo %s: %v", filePath, err)
	}

	sha := strings.TrimSpace(string(output))
	return fmt.Sprintf("Archivo: %s, SHA: %s, Directorio: %s", filePath, sha, workingDir), nil
}

func (c *Client) GetLastCommit() (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	workingDir := c.getEffectiveWorkingDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	cmd := c.executor.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error obteniendo SHA del commit: %v", err)
	}

	sha := strings.TrimSpace(string(output))
	return fmt.Sprintf("Último commit SHA: %s, Directorio: %s", sha, workingDir), nil
}

func (c *Client) GetFileContent(filePath, ref string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	workingDir := c.getEffectiveWorkingDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	if ref == "" {
		ref = "HEAD"
	}

	cmd := c.executor.Command("git", "show", fmt.Sprintf("%s:%s", ref, filePath))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error obteniendo contenido del archivo %s en %s: %v", filePath, ref, err)
	}

	return fmt.Sprintf("Archivo: %s, Ref: %s, Contenido: %s", filePath, ref, string(output)), nil
}

func (c *Client) GetChangedFiles(staged bool) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	workingDir := c.getEffectiveWorkingDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	var cmd cmdWrapper
	if staged {
		cmd = c.executor.Command("git", "diff", "--cached", "--name-only")
	} else {
		cmd = c.executor.Command("git", "diff", "--name-only")
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error obteniendo archivos modificados: %v", err)
	}

	status := "working directory"
	if staged {
		status = "staging area"
	}

	return fmt.Sprintf("Directorio: %s, Archivos modificados (%s): %s", workingDir, status, string(output)), nil
}

func (c *Client) ValidateRepo(path string) (string, error) {
	gitPath := filepath.Join(path, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return "", fmt.Errorf("no es un repositorio Git: %s", path)
	}

	if _, err := c.executor.LookPath("git"); err != nil {
		return "", fmt.Errorf("git no está disponible en el sistema")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(path)

	cmd := c.executor.Command("git", "remote", "get-url", "origin")
	remoteOutput, _ := cmd.Output()

	cmd = c.executor.Command("git", "branch", "--show-current")
	branchOutput, _ := cmd.Output()

	return fmt.Sprintf("Repositorio Git válido: %s, Rama: %s, Remote: %s",
		path,
		strings.TrimSpace(string(branchOutput)),
		strings.TrimSpace(string(remoteOutput))), nil
}

func (c *Client) ListFiles(ref string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	workingDir := c.getEffectiveWorkingDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	if ref == "" {
		ref = "HEAD"
	}

	cmd := c.executor.Command("git", "ls-tree", "--name-only", "-r", ref)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error listando archivos en %s: %v", ref, err)
	}

	files := strings.TrimSpace(string(output))
	fileCount := strings.Count(files, "\n") + 1
	if files == "" {
		fileCount = 0
	}

	return fmt.Sprintf("Directorio: %s, Ref: %s, Total archivos: %d, Archivos: %s", workingDir, ref, fileCount, files), nil
}

func (c *Client) LogAnalysis(limit string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	if limit == "" {
		limit = "20"
	}

	result := map[string]interface{}{}

	cmd := c.executor.Command("git", "log", "--graph", "--oneline", "--decorate", "--all", "-"+limit)
	if output, err := cmd.Output(); err == nil {
		result["graphLog"] = strings.TrimSpace(string(output))
	}

	cmd = c.executor.Command("git", "shortlog", "-sn", "--all")
	if output, err := cmd.Output(); err == nil {
		result["authorStats"] = strings.TrimSpace(string(output))
	}

	cmd = c.executor.Command("git", "log", "--pretty=format:%h|%an|%ad|%s", "--date=short", "-"+limit)
	if output, err := cmd.Output(); err == nil {
		result["recentCommits"] = strings.TrimSpace(string(output))
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}

func (c *Client) DiffFiles(staged bool) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	result := map[string]interface{}{}

	var cmd cmdWrapper
	if staged {
		cmd = c.executor.Command("git", "diff", "--name-status", "--cached")
	} else {
		cmd = c.executor.Command("git", "diff", "--name-status")
	}

	if output, err := cmd.Output(); err == nil {
		result["files"] = strings.TrimSpace(string(output))
	}

	if staged {
		cmd = c.executor.Command("git", "diff", "--stat", "--cached")
	} else {
		cmd = c.executor.Command("git", "diff", "--stat")
	}

	if output, err := cmd.Output(); err == nil {
		result["stats"] = strings.TrimSpace(string(output))
	}

	if !staged {
		cmd = c.executor.Command("git", "ls-files", "--others", "--exclude-standard")
		if output, err := cmd.Output(); err == nil {
			untracked := strings.TrimSpace(string(output))
			if untracked != "" {
				result["untracked"] = untracked
			}
		}
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}

func (c *Client) BranchList(remote bool) ([]types.BranchInfo, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return nil, fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	workingDir := c.getEffectiveWorkingDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	// Get current branch
	cmdCurrent := c.executor.Command("git", "branch", "--show-current")
	currentBranchBytes, err := cmdCurrent.Output()
	if err != nil {
		// This can fail if in detached HEAD state, not a fatal error
		currentBranchBytes = []byte{}
	}
	currentBranchName := strings.TrimSpace(string(currentBranchBytes))

	// List branches with details
	args := []string{"for-each-ref", "--format=%(refname:short)|%(objectname:short)|%(committerdate:iso)", "refs/heads"}
	if remote {
		args = append(args, "refs/remotes")
	}
	cmdList := c.executor.Command("git", args...)
	output, err := cmdList.Output()
	if err != nil {
		return nil, fmt.Errorf("error listing branches: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var branches []types.BranchInfo

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) < 3 {
			continue
		}
		branchName := parts[0]

		// Skip remote HEAD pointer
		if strings.HasSuffix(branchName, "/HEAD") {
			continue
		}

		branches = append(branches, types.BranchInfo{
			Name:       branchName,
			IsCurrent:  branchName == currentBranchName,
			CommitSHA:  parts[1],
			CommitDate: parts[2],
		})
	}

	return branches, nil
}

func (c *Client) Stash(operation, name string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	var cmd cmdWrapper
	var result string

	switch operation {
	case "list":
		cmd = c.executor.Command("git", "stash", "list")
		if output, err := cmd.Output(); err == nil {
			result = fmt.Sprintf("Stash list: %s", strings.TrimSpace(string(output)))
		} else {
			result = "No hay stashes guardados"
		}

	case "push":
		if name != "" {
			cmd = c.executor.Command("git", "stash", "push", "-m", name)
		} else {
			cmd = c.executor.Command("git", "stash", "push")
		}
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Stash creado: %s", strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error creando stash: %v, Output: %s", err, output)
		}

	case "pop":
		if name != "" {
			cmd = c.executor.Command("git", "stash", "pop", name)
		} else {
			cmd = c.executor.Command("git", "stash", "pop")
		}
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Stash aplicado y eliminado: %s", strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error aplicando stash: %v, Output: %s", err, output)
		}

	case "apply":
		if name != "" {
			cmd = c.executor.Command("git", "stash", "apply", name)
		} else {
			cmd = c.executor.Command("git", "stash", "apply")
		}
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Stash aplicado (mantenido): %s", strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error aplicando stash: %v, Output: %s", err, output)
		}

	case "drop":
		if name != "" {
			cmd = c.executor.Command("git", "stash", "drop", name)
		} else {
			cmd = c.executor.Command("git", "stash", "drop")
		}
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Stash eliminado: %s", strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error eliminando stash: %v, Output: %s", err, output)
		}

	case "clear":
		cmd = c.executor.Command("git", "stash", "clear")
		if output, err := cmd.CombinedOutput(); err == nil {
			result = "Todos los stashes han sido eliminados"
		} else {
			return "", fmt.Errorf("error limpiando stashes: %v, Output: %s", err, output)
		}

	default:
		return "", fmt.Errorf("operación no válida: %s. Usa: list, push, pop, apply, drop, clear", operation)
	}

	return result, nil
}

func (c *Client) Remote(operation, name, url string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	var cmd cmdWrapper
	var result string

	switch operation {
	case "list":
		cmd = c.executor.Command("git", "remote", "-v")
		if output, err := cmd.Output(); err == nil {
			result = fmt.Sprintf("Remotos configurados: %s", strings.TrimSpace(string(output)))
		} else {
			result = "No hay remotos configurados"
		}

	case "add":
		if name == "" || url == "" {
			return "", fmt.Errorf("nombre y URL requeridos para agregar remoto")
		}
		cmd = c.executor.Command("git", "remote", "add", name, url)
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Remoto '%s' agregado: %s", name, url)
		} else {
			return "", fmt.Errorf("error agregando remoto: %v, Output: %s", err, output)
		}

	case "remove":
		if name == "" {
			return "", fmt.Errorf("nombre del remoto requerido")
		}
		cmd = c.executor.Command("git", "remote", "remove", name)
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Remoto '%s' eliminado", name)
		} else {
			return "", fmt.Errorf("error eliminando remoto: %v, Output: %s", err, output)
		}

	case "show":
		if name == "" {
			name = "origin"
		}
		cmd = c.executor.Command("git", "remote", "show", name)
		if output, err := cmd.Output(); err == nil {
			result = fmt.Sprintf("Información del remoto '%s': %s", name, strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error mostrando remoto: %v", err)
		}

	case "fetch":
		if name == "" {
			cmd = c.executor.Command("git", "fetch", "--all")
			result = "Fetching desde todos los remotos"
		} else {
			cmd = c.executor.Command("git", "fetch", name)
			result = fmt.Sprintf("Fetching desde '%s'", name)
		}
		if output, err := cmd.CombinedOutput(); err == nil {
			result += fmt.Sprintf(": %s", strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error en fetch: %v, Output: %s", err, output)
		}

	default:
		return "", fmt.Errorf("operación no válida: %s. Usa: list, add, remove, show, fetch", operation)
	}

	return result, nil
}

func (c *Client) Tag(operation, tagName, message string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	var cmd cmdWrapper
	var result string

	switch operation {
	case "list":
		cmd = c.executor.Command("git", "tag", "-l", "--sort=-version:refname")
		if output, err := cmd.Output(); err == nil {
			result = fmt.Sprintf("Tags disponibles: %s", strings.TrimSpace(string(output)))
		} else {
			result = "No hay tags creados"
		}

	case "create":
		if tagName == "" {
			return "", fmt.Errorf("nombre del tag requerido")
		}
		if message != "" {
			cmd = c.executor.Command("git", "tag", "-a", tagName, "-m", message)
		} else {
			cmd = c.executor.Command("git", "tag", tagName)
		}
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Tag '%s' creado", tagName)
		} else {
			return "", fmt.Errorf("error creando tag: %v, Output: %s", err, output)
		}

	case "delete":
		if tagName == "" {
			return "", fmt.Errorf("nombre del tag requerido")
		}
		cmd = c.executor.Command("git", "tag", "-d", tagName)
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Tag '%s' eliminado localmente", tagName)
		} else {
			return "", fmt.Errorf("error eliminando tag: %v, Output: %s", err, output)
		}

	case "push":
		if tagName == "" {
			cmd = c.executor.Command("git", "push", "origin", "--tags")
			result = "Todos los tags enviados al remoto"
		} else {
			cmd = c.executor.Command("git", "push", "origin", tagName)
			result = fmt.Sprintf("Tag '%s' enviado al remoto", tagName)
		}
		if output, err := cmd.CombinedOutput(); err == nil {
			result += fmt.Sprintf(": %s", strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error enviando tags: %v, Output: %s", err, output)
		}

	case "show":
		if tagName == "" {
			return "", fmt.Errorf("nombre del tag requerido")
		}
		cmd = c.executor.Command("git", "show", tagName)
		if output, err := cmd.Output(); err == nil {
			result = fmt.Sprintf("Información del tag '%s': %s", tagName, strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error mostrando tag: %v", err)
		}

	default:
		return "", fmt.Errorf("operación no válida: %s. Usa: list, create, delete, push, show", operation)
	}

	return result, nil
}

func (c *Client) Clean(operation string, dryRun bool) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	var cmd cmdWrapper
	var result string

	switch operation {
	case "untracked":
		if dryRun {
			cmd = c.executor.Command("git", "clean", "-n")
			result = "Vista previa - archivos que se eliminarían:"
		} else {
			cmd = c.executor.Command("git", "clean", "-f")
			result = "Archivos sin seguimiento eliminados:"
		}

	case "untracked_dirs":
		if dryRun {
			cmd = c.executor.Command("git", "clean", "-n", "-d")
			result = "Vista previa - archivos y directorios que se eliminarían:"
		} else {
			cmd = c.executor.Command("git", "clean", "-f", "-d")
			result = "Archivos y directorios sin seguimiento eliminados:"
		}

	case "ignored":
		if dryRun {
			cmd = c.executor.Command("git", "clean", "-n", "-X")
			result = "Vista previa - archivos ignorados que se eliminarían:"
		} else {
			cmd = c.executor.Command("git", "clean", "-f", "-X")
			result = "Archivos ignorados eliminados:"
		}

	case "all":
		if dryRun {
			cmd = c.executor.Command("git", "clean", "-n", "-d", "-x")
			result = "Vista previa - todos los archivos sin seguimiento que se eliminarían:"
		} else {
			cmd = c.executor.Command("git", "clean", "-f", "-d", "-x")
			result = "Todos los archivos sin seguimiento eliminados:"
		}

	default:
		return "", fmt.Errorf("operación no válida: %s. Usa: untracked, untracked_dirs, ignored, all", operation)
	}

	if output, err := cmd.Output(); err == nil {
		cleanResult := strings.TrimSpace(string(output))
		if cleanResult == "" {
			result += " No hay archivos para procesar"
		} else {
			result += fmt.Sprintf(" %s", cleanResult)
		}
	} else {
		return "", fmt.Errorf("error en limpieza: %v", err)
	}

	return result, nil
}

// Métodos para acceder a la configuración
func (c *Client) HasGit() bool {
	return c.Config.HasGit
}

func (c *Client) IsGitRepo() bool {
	return c.Config.IsGitRepo
}

func (c *Client) GetRepoPath() string {
	return c.Config.RepoPath
}

func (c *Client) GetCurrentBranch() string {
	return c.Config.CurrentBranch
}

func (c *Client) GetRemoteURL() string {
	return c.Config.RemoteURL
}
