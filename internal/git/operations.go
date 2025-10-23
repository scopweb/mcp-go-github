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

	// Paso 1: Fetch para tener info actualizada del remoto
	fetchCmd := c.executor.Command("git", "fetch", "origin", effectiveBranch)
	if output, err := fetchCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("error en fetch inicial: %v, Output: %s", err, output)
	}

	// Paso 2: Detectar divergencias
	divergenceInfo, err := c.checkDivergence(effectiveBranch)
	if err != nil {
		return "", fmt.Errorf("error detectando divergencias: %v", err)
	}

	// Paso 3: Si no hay divergencias, hacer fast-forward
	if divergenceInfo["canFastForward"].(bool) {
		cmd := c.executor.Command("git", "pull", "--ff-only", "origin", effectiveBranch)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("error en pull (ff-only): %v, Output: %s", err, output)
		}
		return fmt.Sprintf("Pull exitoso (fast-forward) desde rama: %s", effectiveBranch), nil
	}

	// Paso 4: Si hay divergencias, intentar merge regular
	if divergenceInfo["aheadCount"].(int) > 0 && divergenceInfo["behindCount"].(int) > 0 {
		// Divergencia: usar merge strategy
		cmd := c.executor.Command("git", "pull", "--no-rebase", "origin", effectiveBranch)
		output, err := cmd.CombinedOutput()
		if err != nil {
			if strings.Contains(string(output), "CONFLICT") {
				return "", fmt.Errorf("conflictos detectados durante pull. Divergencia: %d commits locales, %d commits remotos. Resuelve manualmente o usa PullWithStrategy con 'rebase'", divergenceInfo["aheadCount"].(int), divergenceInfo["behindCount"].(int))
			}
			return "", fmt.Errorf("error en pull: %v, Output: %s", err, output)
		}
		return fmt.Sprintf("Pull exitoso (merge) desde rama: %s. Divergencia resuelta: %d commits locales + %d commits remotos", effectiveBranch, divergenceInfo["aheadCount"].(int), divergenceInfo["behindCount"].(int)), nil
	}

	// Solo cambios remotos
	cmd := c.executor.Command("git", "pull", "origin", effectiveBranch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error ejecutando git pull: %v, Output: %s", err, output)
	}

	return fmt.Sprintf("Pull realizado desde rama: %s", effectiveBranch), nil
}

// checkDivergence detecta si hay divergencias entre rama local y remota
func (c *Client) checkDivergence(branch string) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"canFastForward": false,
		"aheadCount":     0,
		"behindCount":    0,
		"isDiverged":     false,
	}

	// Contar commits adelante y atrás
	countCmd := c.executor.Command("git", "rev-list", "--left-right", "--count", fmt.Sprintf("HEAD...origin/%s", branch))
	output, err := countCmd.Output()
	if err != nil {
		return result, fmt.Errorf("error contando divergencias: %v", err)
	}

	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) >= 2 {
		aheadCount := 0
		behindCount := 0
		fmt.Sscanf(parts[0], "%d", &aheadCount)
		fmt.Sscanf(parts[1], "%d", &behindCount)

		result["aheadCount"] = aheadCount
		result["behindCount"] = behindCount
		result["isDiverged"] = aheadCount > 0 && behindCount > 0
		result["canFastForward"] = aheadCount == 0 && behindCount > 0
	}

	return result, nil
}

func (c *Client) Checkout(branch string, create bool) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	if branch == "" {
		return "", fmt.Errorf("nombre de rama requerido")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	// Paso 1: Validar que la rama existe (si no es creación de rama nueva)
	if !create {
		checkCmd := c.executor.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
		if _, err := checkCmd.CombinedOutput(); err != nil {
			// Intentar desde remoto
			checkRemoteCmd := c.executor.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch)
			if _, err := checkRemoteCmd.CombinedOutput(); err != nil {
				return "", fmt.Errorf("rama '%s' no existe (ni local ni remota). Crea con 'create: true' o usa 'CheckoutRemote'", branch)
			}
		}
	}

	// Paso 2: Validar estado del working directory
	clean, err := c.ValidateCleanState()
	if err != nil {
		return "", fmt.Errorf("error validando estado: %v", err)
	}

	// Paso 3: Si hay cambios sin commitear, hacer stash automático
	stashApplied := false
	stashName := ""
	if !clean {
		stashName = fmt.Sprintf("auto-stash-before-checkout-%s", branch)
		if _, err := c.Stash("push", stashName); err != nil {
			return "", fmt.Errorf("error guardando cambios con stash: %v. Debes commitear o descartar los cambios primero", err)
		}
		stashApplied = true
	}

	// Paso 4: Ejecutar checkout
	var cmd cmdWrapper
	if create {
		cmd = c.executor.Command("git", "checkout", "-b", branch)
	} else {
		cmd = c.executor.Command("git", "checkout", branch)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Si falló, restaurar stash si fue aplicado
		if stashApplied {
			c.Stash("pop", stashName)
		}
		return "", fmt.Errorf("error ejecutando git checkout: %v, Output: %s", err, output)
	}

	c.Config.CurrentBranch = branch

	result := fmt.Sprintf("Checkout exitoso a rama: %s", branch)
	if create {
		result += " (nueva rama creada)"
	}
	if stashApplied {
		result += fmt.Sprintf(" [cambios guardados en stash: %s]", stashName)
	}
	return result, nil
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

// Advanced branch operations

func (c *Client) CheckoutRemote(remoteBranch string, localBranch string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	// Ensure we have the latest remote info
	fetchCmd := c.executor.Command("git", "fetch", "origin")
	if output, err := fetchCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("error en fetch: %v, Output: %s", err, output)
	}

	// If no local branch specified, use remote branch name without origin/
	if localBranch == "" {
		parts := strings.Split(remoteBranch, "/")
		if len(parts) > 1 {
			localBranch = parts[len(parts)-1]
		} else {
			localBranch = remoteBranch
		}
	}

	// Check if local branch already exists
	checkCmd := c.executor.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+localBranch)
	if _, err := checkCmd.CombinedOutput(); err == nil {
		// Local branch exists, just checkout and pull
		checkoutCmd := c.executor.Command("git", "checkout", localBranch)
		if output, err := checkoutCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error en checkout: %v, Output: %s", err, output)
		}
		
		pullCmd := c.executor.Command("git", "pull", "origin", remoteBranch)
		if output, err := pullCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error en pull: %v, Output: %s", err, output)
		}
	} else {
		// Create new local branch tracking remote
		cmd := c.executor.Command("git", "checkout", "-b", localBranch, "origin/"+remoteBranch)
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error en checkout remoto: %v, Output: %s", err, output)
		}
	}

	c.Config.CurrentBranch = localBranch
	return fmt.Sprintf("Checkout remoto exitoso: %s -> %s", remoteBranch, localBranch), nil
}

func (c *Client) Merge(sourceBranch string, targetBranch string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	// Validate clean state
	if clean, err := c.ValidateCleanState(); err != nil {
		return "", fmt.Errorf("error validando estado: %v", err)
	} else if !clean {
		return "", fmt.Errorf("el directorio de trabajo debe estar limpio para hacer merge")
	}

	// If no target branch specified, use current branch
	if targetBranch == "" {
		targetBranch = c.Config.CurrentBranch
	} else if targetBranch != c.Config.CurrentBranch {
		// Checkout to target branch
		checkoutCmd := c.executor.Command("git", "checkout", targetBranch)
		if output, err := checkoutCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error cambiando a rama %s: %v, Output: %s", targetBranch, err, output)
		}
		c.Config.CurrentBranch = targetBranch
	}

	// Perform merge
	mergeCmd := c.executor.Command("git", "merge", sourceBranch)
	output, err := mergeCmd.CombinedOutput()
	if err != nil {
		// Check if it's a conflict
		statusCmd := c.executor.Command("git", "status", "--porcelain")
		statusOut, _ := statusCmd.Output()
		if strings.Contains(string(statusOut), "UU") || strings.Contains(string(output), "CONFLICT") {
			return "", fmt.Errorf("conflictos de merge detectados. Usa 'ConflictStatus' para ver detalles y 'ResolveConflicts' para resolverlos: %s", output)
		}
		return "", fmt.Errorf("error en merge: %v, Output: %s", err, output)
	}

	return fmt.Sprintf("Merge exitoso: %s -> %s", sourceBranch, targetBranch), nil
}

func (c *Client) Rebase(branch string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	// Validate clean state
	if clean, err := c.ValidateCleanState(); err != nil {
		return "", fmt.Errorf("error validando estado: %v", err)
	} else if !clean {
		return "", fmt.Errorf("el directorio de trabajo debe estar limpio para hacer rebase")
	}

	// Perform rebase
	rebaseCmd := c.executor.Command("git", "rebase", branch)
	output, err := rebaseCmd.CombinedOutput()
	if err != nil {
		// Check if it's a conflict
		if strings.Contains(string(output), "CONFLICT") {
			return "", fmt.Errorf("conflictos de rebase detectados. Usa 'ConflictStatus' para ver detalles: %s", output)
		}
		return "", fmt.Errorf("error en rebase: %v, Output: %s", err, output)
	}

	return fmt.Sprintf("Rebase exitoso en rama: %s", branch), nil
}

// Enhanced pull/push operations

func (c *Client) PullWithStrategy(branch string, strategy string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	if branch == "" {
		branch = c.Config.CurrentBranch
	}

	var cmd cmdWrapper
	switch strategy {
	case "merge":
		cmd = c.executor.Command("git", "pull", "--no-rebase", "origin", branch)
	case "rebase":
		cmd = c.executor.Command("git", "pull", "--rebase", "origin", branch)
	case "ff-only":
		cmd = c.executor.Command("git", "pull", "--ff-only", "origin", branch)
	default:
		return "", fmt.Errorf("estrategia no válida: %s. Usa: merge, rebase, ff-only", strategy)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "CONFLICT") {
			return "", fmt.Errorf("conflictos detectados durante pull con estrategia %s: %s", strategy, output)
		}
		return "", fmt.Errorf("error en pull con estrategia %s: %v, Output: %s", strategy, err, output)
	}

	return fmt.Sprintf("Pull con estrategia '%s' exitoso en rama: %s", strategy, branch), nil
}

func (c *Client) ForcePush(branch string, force bool) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	if branch == "" {
		branch = c.Config.CurrentBranch
	}

	// Get remote name
	remoteCmd := c.executor.Command("git", "remote")
	remoteOutput, err := remoteCmd.Output()
	if err != nil {
		return "", fmt.Errorf("error obteniendo remotos: %v", err)
	}
	remotes := strings.Fields(string(remoteOutput))
	if len(remotes) == 0 {
		return "", errors.New("no se encontraron remotos")
	}
	remote := remotes[0]

	var cmd cmdWrapper
	if force {
		// Create backup before force push
		backupName := fmt.Sprintf("backup-before-force-push-%s", branch)
		if _, err := c.CreateBackup(backupName); err != nil {
			return "", fmt.Errorf("error creando backup antes de force push: %v", err)
		}
		
		cmd = c.executor.Command("git", "push", "--force-with-lease", remote, branch)
	} else {
		cmd = c.executor.Command("git", "push", remote, branch)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error en push: %v, Output: %s", err, output)
	}

	if force {
		return fmt.Sprintf("Force push exitoso (con backup): %s a %s", branch, remote), nil
	}
	return fmt.Sprintf("Push exitoso: %s a %s", branch, remote), nil
}

func (c *Client) PushUpstream(branch string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	if branch == "" {
		branch = c.Config.CurrentBranch
	}

	// Get remote name
	remoteCmd := c.executor.Command("git", "remote")
	remoteOutput, err := remoteCmd.Output()
	if err != nil {
		return "", fmt.Errorf("error obteniendo remotos: %v", err)
	}
	remotes := strings.Fields(string(remoteOutput))
	if len(remotes) == 0 {
		return "", errors.New("no se encontraron remotos")
	}
	remote := remotes[0]

	cmd := c.executor.Command("git", "push", "-u", remote, branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error en push upstream: %v, Output: %s", err, output)
	}

	return fmt.Sprintf("Push upstream exitoso: %s configurado para trackear %s/%s", branch, remote, branch), nil
}

// Batch operations

func (c *Client) SyncWithRemote(remoteBranch string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	results := []string{}

	// 1. Fetch from remote
	fetchCmd := c.executor.Command("git", "fetch", "origin")
	if output, err := fetchCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("error en fetch: %v, Output: %s", err, output)
	}
	results = append(results, "Fetch completado")

	// 2. Check if we need to merge
	currentBranch := c.Config.CurrentBranch
	if remoteBranch == "" {
		remoteBranch = currentBranch
	}

	// Check if remote branch exists
	checkCmd := c.executor.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+remoteBranch)
	if _, err := checkCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("rama remota no encontrada: origin/%s", remoteBranch)
	}

	// 3. Check if fast-forward is possible
	mergeBaseCmd := c.executor.Command("git", "merge-base", currentBranch, "origin/"+remoteBranch)
	mergeBase, _ := mergeBaseCmd.Output()
	
	currentCommitCmd := c.executor.Command("git", "rev-parse", currentBranch)
	currentCommit, _ := currentCommitCmd.Output()

	if strings.TrimSpace(string(mergeBase)) == strings.TrimSpace(string(currentCommit)) {
		// Fast-forward possible
		mergeCmd := c.executor.Command("git", "merge", "--ff-only", "origin/"+remoteBranch)
		if output, err := mergeCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error en fast-forward: %v, Output: %s", err, output)
		}
		results = append(results, "Fast-forward merge completado")
	} else {
		// Need regular merge
		if clean, err := c.ValidateCleanState(); err != nil {
			return "", fmt.Errorf("error validando estado: %v", err)
		} else if !clean {
			return "", fmt.Errorf("el directorio debe estar limpio para sincronizar")
		}

		mergeCmd := c.executor.Command("git", "merge", "origin/"+remoteBranch)
		if output, err := mergeCmd.CombinedOutput(); err != nil {
			if strings.Contains(string(output), "CONFLICT") {
				return "", fmt.Errorf("conflictos detectados durante sincronización: %s", output)
			}
			return "", fmt.Errorf("error en merge: %v, Output: %s", err, output)
		}
		results = append(results, "Merge completado")
	}

	return fmt.Sprintf("Sincronización exitosa con origin/%s: %s", remoteBranch, strings.Join(results, ", ")), nil
}

func (c *Client) SafeMerge(source string, target string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	// 1. Create backup
	backupName := fmt.Sprintf("safe-merge-backup-%s", target)
	if _, err := c.CreateBackup(backupName); err != nil {
		return "", fmt.Errorf("error creando backup: %v", err)
	}

	// 2. Validate clean state
	if clean, err := c.ValidateCleanState(); err != nil {
		return "", fmt.Errorf("error validando estado: %v", err)
	} else if !clean {
		return "", fmt.Errorf("el directorio debe estar limpio para safe merge")
	}

	// 3. Check for potential conflicts
	if conflicts, err := c.DetectPotentialConflicts(source, target); err != nil {
		return "", fmt.Errorf("error detectando conflictos: %v", err)
	} else if conflicts != "" {
		return "", fmt.Errorf("conflictos potenciales detectados: %s", conflicts)
	}

	// 4. Perform merge
	originalBranch := c.Config.CurrentBranch
	
	// Switch to target branch if needed
	if target != "" && target != originalBranch {
		checkoutCmd := c.executor.Command("git", "checkout", target)
		if output, err := checkoutCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error cambiando a rama %s: %v, Output: %s", target, err, output)
		}
		c.Config.CurrentBranch = target
	}

	// Perform merge
	mergeCmd := c.executor.Command("git", "merge", "--no-ff", source)
	output, err := mergeCmd.CombinedOutput()
	if err != nil {
		// Rollback on error
		resetCmd := c.executor.Command("git", "reset", "--hard", "HEAD~1")
		resetCmd.CombinedOutput()
		
		return "", fmt.Errorf("safe merge falló, rollback realizado: %v, Output: %s", err, output)
	}

	return fmt.Sprintf("Safe merge exitoso: %s -> %s (backup creado: %s)", source, target, backupName), nil
}

// Conflict management

func (c *Client) ConflictStatus() (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	result := map[string]interface{}{}

	// Check if in merge state
	mergeHeadPath := filepath.Join(c.Config.RepoPath, ".git", "MERGE_HEAD")
	if _, err := os.Stat(mergeHeadPath); err == nil {
		result["inMergeState"] = true
		
		// Get merge message if exists
		mergeMsgPath := filepath.Join(c.Config.RepoPath, ".git", "MERGE_MSG")
		if msgBytes, err := os.ReadFile(mergeMsgPath); err == nil {
			result["mergeMessage"] = string(msgBytes)
		}
	} else {
		result["inMergeState"] = false
	}

	// Check if in rebase state
	rebaseDirPath := filepath.Join(c.Config.RepoPath, ".git", "rebase-merge")
	rebaseApplyPath := filepath.Join(c.Config.RepoPath, ".git", "rebase-apply")
	if _, err := os.Stat(rebaseDirPath); err == nil {
		result["inRebaseState"] = true
	} else if _, err := os.Stat(rebaseApplyPath); err == nil {
		result["inRebaseState"] = true
	} else {
		result["inRebaseState"] = false
	}

	// Get conflicted files
	statusCmd := c.executor.Command("git", "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return "", fmt.Errorf("error obteniendo status: %v", err)
	}

	conflictedFiles := []string{}
	lines := strings.Split(string(statusOutput), "\n")
	for _, line := range lines {
		if len(line) > 2 && (line[:2] == "UU" || line[:2] == "AA" || line[:2] == "DD" || 
			strings.Contains(line[:2], "U") || strings.Contains(line[:2], "A")) {
			conflictedFiles = append(conflictedFiles, strings.TrimSpace(line[2:]))
		}
	}
	
	result["conflictedFiles"] = conflictedFiles
	result["hasConflicts"] = len(conflictedFiles) > 0

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}

func (c *Client) ResolveConflicts(strategy string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	// Paso 1: Obtener lista de archivos en conflicto
	conflictedFiles, err := c.getConflictedFiles()
	if err != nil {
		return "", fmt.Errorf("error obteniendo archivos en conflicto: %v", err)
	}

	if len(conflictedFiles) == 0 {
		return "No hay conflictos para resolver", nil
	}

	var cmd cmdWrapper
	var result string

	switch strategy {
	case "theirs":
		// Aceptar versión remota para todos los conflictos
		cmd = c.executor.Command("git", "checkout", "--theirs", ".")
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error aceptando cambios remotos: %v, Output: %s", err, output)
		}

		// Agregar archivos resueltos
		addCmd := c.executor.Command("git", "add", ".")
		if output, err := addCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error agregando archivos resueltos: %v, Output: %s", err, output)
		}

		result = fmt.Sprintf("Conflictos resueltos en %d archivo(s) aceptando versión remota (theirs)", len(conflictedFiles))

	case "ours":
		// Aceptar versión local para todos los conflictos
		cmd = c.executor.Command("git", "checkout", "--ours", ".")
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error aceptando cambios locales: %v, Output: %s", err, output)
		}

		// Agregar archivos resueltos
		addCmd := c.executor.Command("git", "add", ".")
		if output, err := addCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error agregando archivos resueltos: %v, Output: %s", err, output)
		}

		result = fmt.Sprintf("Conflictos resueltos en %d archivo(s) aceptando versión local (ours)", len(conflictedFiles))

	case "abort":
		// Abortar merge o rebase
		mergeHeadPath := filepath.Join(c.Config.RepoPath, ".git", "MERGE_HEAD")
		if _, err := os.Stat(mergeHeadPath); err == nil {
			cmd = c.executor.Command("git", "merge", "--abort")
		} else {
			// Verificar si hay rebase activo
			rebaseDirPath := filepath.Join(c.Config.RepoPath, ".git", "rebase-merge")
			rebaseApplyPath := filepath.Join(c.Config.RepoPath, ".git", "rebase-apply")
			_, rebaseDirErr := os.Stat(rebaseDirPath)
			_, rebaseApplyErr := os.Stat(rebaseApplyPath)
			if rebaseDirErr == nil || rebaseApplyErr == nil {
				cmd = c.executor.Command("git", "rebase", "--abort")
			} else {
				return "", fmt.Errorf("no hay operación de merge o rebase activa para abortar")
			}
		}

		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error abortando operación: %v, Output: %s", err, output)
		}

		result = "Operación abortada, repositorio restaurado"

	case "manual":
		// Mostrar información detallada de archivos en conflicto
		var details []string
		details = append(details, fmt.Sprintf("Archivos en conflicto (%d):", len(conflictedFiles)))
		for i, file := range conflictedFiles {
			details = append(details, fmt.Sprintf("  %d. %s", i+1, file))
		}
		details = append(details, "")
		details = append(details, "Usa ResolveFile para resolver archivo por archivo")
		details = append(details, "Ejemplo: ResolveFile('file.txt', 'ours')")
		result = strings.Join(details, "\n")

	default:
		return "", fmt.Errorf("estrategia no válida: %s. Usa: theirs, ours, abort, manual", strategy)
	}

	return result, nil
}

// getConflictedFiles obtiene la lista de archivos en conflicto
func (c *Client) getConflictedFiles() ([]string, error) {
	statusCmd := c.executor.Command("git", "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo status: %v", err)
	}

	var conflictedFiles []string
	lines := strings.Split(string(statusOutput), "\n")
	for _, line := range lines {
		if len(line) > 2 {
			// Archivos en conflicto tienen estado "UU", "AA", "DD", "AU", "UA", etc.
			status := line[:2]
			if (status[0] == 'U' || status[0] == 'A' || status[0] == 'D') &&
				(status[1] == 'U' || status[1] == 'A' || status[1] == 'D') {
				conflictedFiles = append(conflictedFiles, strings.TrimSpace(line[2:]))
			}
		}
	}
	return conflictedFiles, nil
}

// Validation operations

func (c *Client) ValidateCleanState() (bool, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return false, fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	cmd := c.executor.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("error obteniendo status: %v", err)
	}

	// Empty output means clean state
	return strings.TrimSpace(string(output)) == "", nil
}

func (c *Client) DetectPotentialConflicts(sourceBranch string, targetBranch string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	// Get merge base
	mergeBaseCmd := c.executor.Command("git", "merge-base", sourceBranch, targetBranch)
	mergeBase, err := mergeBaseCmd.Output()
	if err != nil {
		return "", fmt.Errorf("error encontrando merge base: %v", err)
	}
	
	mergeBaseHash := strings.TrimSpace(string(mergeBase))

	// Get files changed in both branches since merge base
	sourceFilesCmd := c.executor.Command("git", "diff", "--name-only", mergeBaseHash, sourceBranch)
	sourceFiles, err := sourceFilesCmd.Output()
	if err != nil {
		return "", fmt.Errorf("error obteniendo archivos de rama origen: %v", err)
	}

	targetFilesCmd := c.executor.Command("git", "diff", "--name-only", mergeBaseHash, targetBranch)
	targetFiles, err := targetFilesCmd.Output()
	if err != nil {
		return "", fmt.Errorf("error obteniendo archivos de rama destino: %v", err)
	}

	sourceSet := make(map[string]bool)
	for _, file := range strings.Split(strings.TrimSpace(string(sourceFiles)), "\n") {
		if file != "" {
			sourceSet[file] = true
		}
	}

	var conflictingFiles []string
	for _, file := range strings.Split(strings.TrimSpace(string(targetFiles)), "\n") {
		if file != "" && sourceSet[file] {
			conflictingFiles = append(conflictingFiles, file)
		}
	}

	if len(conflictingFiles) == 0 {
		return "", nil
	}

	return fmt.Sprintf("Archivos potencialmente conflictivos: %s", strings.Join(conflictingFiles, ", ")), nil
}

func (c *Client) CreateBackup(name string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	// Create tag as backup
	tagName := fmt.Sprintf("backup/%s", name)

	// Get current commit
	commitCmd := c.executor.Command("git", "rev-parse", "HEAD")
	commitHash, err := commitCmd.Output()
	if err != nil {
		return "", fmt.Errorf("error obteniendo commit actual: %v", err)
	}

	// Create backup tag
	tagCmd := c.executor.Command("git", "tag", tagName, strings.TrimSpace(string(commitHash)))
	output, err := tagCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error creando backup tag: %v, Output: %s", err, output)
	}

	return fmt.Sprintf("Backup creado: %s en commit %s", tagName, strings.TrimSpace(string(commitHash))[:8]), nil
}

// Helper types para manejo de conflictos
type ConflictMarker struct {
	Ours   string `json:"ours"`
	Theirs string `json:"theirs"`
	Base   string `json:"base,omitempty"`
}

type ConflictDetails struct {
	File      string            `json:"file"`
	HasBase   bool              `json:"hasBase"`
	Markers   []ConflictMarker  `json:"markers"`
	Content   string            `json:"content,omitempty"`
	Suggestion string           `json:"suggestion"`
}

// FASE 1: Nuevos comandos esenciales

// Reset realiza un reset a un commit especificado con el modo indicado (hard, soft, mixed)
func (c *Client) Reset(mode string, target string, files []string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	// Validar modo
	validModes := map[string]bool{"hard": true, "soft": true, "mixed": true}
	if !validModes[mode] {
		return "", fmt.Errorf("modo inválido: %s. Usa: hard, soft, mixed", mode)
	}

	// Validar que no sea reset hard con cambios pendientes (advertencia de seguridad)
	if mode == "hard" {
		clean, err := c.ValidateCleanState()
		if err != nil {
			return "", fmt.Errorf("error validando estado: %v", err)
		}
		if !clean {
			// Crear backup automático antes de reset hard
			if _, err := c.CreateBackup(fmt.Sprintf("before-reset-hard-%s", target)); err != nil {
				return "", fmt.Errorf("error creando backup de seguridad: %v", err)
			}
		}
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	// Validar que el target existe
	validateCmd := c.executor.Command("git", "rev-parse", target)
	if _, err := validateCmd.Output(); err != nil {
		return "", fmt.Errorf("target inválido '%s': %v", target, err)
	}

	var cmd cmdWrapper

	// Reset parcial (archivos específicos)
	if len(files) > 0 {
		args := []string{"reset"}
		if mode != "mixed" { // mixed es el modo default, no necesita especificarse
			args = append(args, "--"+mode)
		}
		args = append(args, target)
		args = append(args, files...)
		cmd = c.executor.Command("git", args...)
	} else {
		// Reset completo
		cmd = c.executor.Command("git", "reset", "--"+mode, target)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error ejecutando git reset: %v, Output: %s", err, output)
	}

	if len(files) > 0 {
		return fmt.Sprintf("Reset parcial exitoso (modo %s) en %d archivo(s) a commit %s", mode, len(files), target), nil
	}
	return fmt.Sprintf("Reset exitoso (modo %s) a commit %s", mode, target), nil
}

// FASE 2: Gestión de Conflictos

// ShowConflict muestra los detalles de un conflicto en un archivo específico
func (c *Client) ShowConflict(filePath string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	// Leer contenido del archivo
	fullPath := filepath.Join(c.Config.RepoPath, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("error leyendo archivo '%s': %v", filePath, err)
	}

	contentStr := string(content)

	// Parsear markers de conflicto
	details := c.parseConflictMarkers(filePath, contentStr)

	// Convertir a JSON
	output, _ := json.MarshalIndent(details, "", "  ")
	return string(output), nil
}

// parseConflictMarkers parsea los markers de conflicto en un archivo
func (c *Client) parseConflictMarkers(filePath string, content string) ConflictDetails {
	details := ConflictDetails{
		File:      filePath,
		Content:   content,
		Suggestion: "Resuelve manualmente o usa ResolveFile con estrategia 'ours' o 'theirs'",
	}

	lines := strings.Split(content, "\n")
	var currentMarker ConflictMarker
	inOurs := false
	inBase := false
	inTheirs := false
	hasConflict := false

	for _, line := range lines {
		if strings.HasPrefix(line, "<<<<<<<") {
			// Inicio de conflicto
			hasConflict = true
			inOurs = true
			inBase = false
			inTheirs = false
			if currentMarker.Ours != "" || currentMarker.Theirs != "" {
				// Nuevo conflicto encontrado
				details.Markers = append(details.Markers, currentMarker)
				currentMarker = ConflictMarker{}
			}
		} else if strings.HasPrefix(line, "=======") {
			// Separador entre ours y theirs (o base y theirs)
			if inOurs {
				inOurs = false
				inTheirs = true
			} else if inBase {
				inBase = false
				inTheirs = true
			}
		} else if strings.HasPrefix(line, "||||||| ") {
			// Separador de base (merge 3-way)
			inOurs = false
			inBase = true
			details.HasBase = true
		} else if strings.HasPrefix(line, ">>>>>>>") {
			// Fin de conflicto
			inTheirs = false
			inOurs = false
			inBase = false
		} else {
			// Acumular contenido
			if inOurs {
				if currentMarker.Ours != "" {
					currentMarker.Ours += "\n"
				}
				currentMarker.Ours += line
			} else if inBase {
				if currentMarker.Base != "" {
					currentMarker.Base += "\n"
				}
				currentMarker.Base += line
			} else if inTheirs {
				if currentMarker.Theirs != "" {
					currentMarker.Theirs += "\n"
				}
				currentMarker.Theirs += line
			}
		}
	}

	// Agregar último marker si existe
	if hasConflict && (currentMarker.Ours != "" || currentMarker.Theirs != "") {
		details.Markers = append(details.Markers, currentMarker)
	}

	return details
}

// ResolveFile resuelve un archivo específico en conflicto
func (c *Client) ResolveFile(filePath string, strategy string, customContent *string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	// Validar estrategia
	validStrategies := map[string]bool{"ours": true, "theirs": true, "manual": true}
	if !validStrategies[strategy] {
		return "", fmt.Errorf("estrategia inválida: %s. Usa: ours, theirs, manual", strategy)
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(c.Config.RepoPath)

	// Validar que el archivo existe
	fullPath := filepath.Join(c.Config.RepoPath, filePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("archivo no encontrado: %s", filePath)
	}

	var cmd cmdWrapper
	var result string

	switch strategy {
	case "ours":
		// Aceptar nuestra versión
		cmd = c.executor.Command("git", "checkout", "--ours", filePath)
		result = "Archivo resuelto aceptando nuestra versión (ours)"

	case "theirs":
		// Aceptar su versión
		cmd = c.executor.Command("git", "checkout", "--theirs", filePath)
		result = "Archivo resuelto aceptando su versión (theirs)"

	case "manual":
		// Usar contenido personalizado
		if customContent == nil {
			return "", fmt.Errorf("contenido personalizado requerido para estrategia 'manual'")
		}

		// Escribir contenido personalizado
		if err := os.WriteFile(fullPath, []byte(*customContent), 0644); err != nil {
			return "", fmt.Errorf("error escribiendo contenido personalizado: %v", err)
		}
		result = "Archivo resuelto con contenido personalizado"
	}

	// Ejecutar comando si no es manual
	if strategy != "manual" {
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("error resolviendo archivo: %v, Output: %s", err, output)
		}
	}

	// Agregar archivo resuelto al staging
	addCmd := c.executor.Command("git", "add", filePath)
	if output, err := addCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("error agregando archivo: %v, Output: %s", err, output)
	}

	return fmt.Sprintf("%s y agregado al staging", result), nil
}
