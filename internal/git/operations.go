package git

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jotajotape/github-go-server-mcp/internal/types"
)

// DetectGitEnvironment detecta y configura el entorno Git local
func DetectGitEnvironment() types.GitConfig {
	config := types.GitConfig{}
	
	// Verificar si git está disponible
	if _, err := exec.LookPath("git"); err == nil {
		config.HasGit = true
	}

	if !config.HasGit {
		return config
	}

	// Obtener directorio actual
	pwd, err := os.Getwd()
	if err != nil {
		return config
	}

	// Buscar repositorio Git (subir directorios hasta encontrar .git)
	repoPath := findGitRepo(pwd)
	if repoPath == "" {
		return config
	}

	config.IsGitRepo = true
	config.RepoPath = repoPath

	// Cambiar al directorio del repositorio para ejecutar comandos
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repoPath)

	// Obtener URL remota
	if output, err := exec.Command("git", "remote", "get-url", "origin").Output(); err == nil {
		config.RemoteURL = strings.TrimSpace(string(output))
	}

	// Obtener rama actual
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
			break // Llegamos a la raíz
		}
		currentPath = parentPath
	}
	return ""
}

// Status muestra el estado del repositorio Git local
func Status(config types.GitConfig) (string, error) {
	result := map[string]interface{}{
		"gitConfig": config,
	}

	if !config.HasGit {
		result["message"] = "Git no está disponible en el sistema"
		output, _ := json.MarshalIndent(result, "", "  ")
		return string(output), nil
	}

	if !config.IsGitRepo {
		result["message"] = "No se detectó repositorio Git en el directorio actual"
		output, _ := json.MarshalIndent(result, "", "  ")
		return string(output), nil
	}

	// Cambiar al directorio del repositorio
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(config.RepoPath)

	// Obtener status
	if output, err := exec.Command("git", "status", "--porcelain").Output(); err == nil {
		result["status"] = strings.TrimSpace(string(output))
	}

	// Obtener log reciente
	if output, err := exec.Command("git", "log", "--oneline", "-5").Output(); err == nil {
		result["recentCommits"] = strings.TrimSpace(string(output))
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}

// Add agrega archivos al staging area
func Add(config types.GitConfig, files string) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	workingDir := GetEffectiveWorkingDir(config)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	cmd := exec.Command("git", "add", files)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error ejecutando git add: %v\nOutput: %s", err, output)
	}

	return fmt.Sprintf("✅ Archivos agregados al staging: %s\n📁 Directorio: %s", files, workingDir), nil
}

// Commit hace commit de los cambios en staging
func Commit(config types.GitConfig, message string) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	workingDir := GetEffectiveWorkingDir(config)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	cmd := exec.Command("git", "commit", "-m", message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error ejecutando git commit: %v\nOutput: %s", err, output)
	}

	return fmt.Sprintf("✅ Commit realizado: %s\n📁 Directorio: %s\n📝 Output: %s", message, workingDir, output), nil
}

// Push sube cambios al repositorio remoto
func Push(config types.GitConfig, branch string) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	workingDir := GetEffectiveWorkingDir(config)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	if branch == "" {
		branch = config.CurrentBranch
	}

	var cmd *exec.Cmd
	if branch != "" {
		cmd = exec.Command("git", "push", "origin", branch)
	} else {
		cmd = exec.Command("git", "push")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error ejecutando git push: %v\nOutput: %s", err, output)
	}

	return fmt.Sprintf("🚀 Push realizado a rama: %s\n📁 Directorio: %s\n📝 Output: %s", branch, workingDir, output), nil
}

// Pull baja cambios del repositorio remoto
func Pull(config types.GitConfig, branch string) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(config.RepoPath)

	if branch == "" {
		branch = config.CurrentBranch
	}

	var cmd *exec.Cmd
	if branch != "" {
		cmd = exec.Command("git", "pull", "origin", branch)
	} else {
		cmd = exec.Command("git", "pull")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error ejecutando git pull: %v\nOutput: %s", err, output)
	}

	return fmt.Sprintf("Pull realizado desde rama: %s", branch), nil
}

// Checkout cambia de rama o crea nueva rama
func Checkout(config *types.GitConfig, branch string, create bool) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(config.RepoPath)

	var cmd *exec.Cmd
	if create {
		cmd = exec.Command("git", "checkout", "-b", branch)
	} else {
		cmd = exec.Command("git", "checkout", branch)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error ejecutando git checkout: %v\nOutput: %s", err, output)
	}

	config.CurrentBranch = branch
	return fmt.Sprintf("Checkout a rama: %s (crear: %v)", branch, create), nil
}

// CreateFile crea un archivo usando Git local
func CreateFile(config types.GitConfig, path, content string) (string, error) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(config.RepoPath)

	// Crear archivo
	fullPath := filepath.Join(config.RepoPath, path)
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	if err != nil {
		return "", fmt.Errorf("error creando directorio: %v", err)
	}

	err = os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("error escribiendo archivo: %v", err)
	}

	return fmt.Sprintf("Archivo creado: %s\nUsa 'git_add' y 'git_commit' para confirmar cambios", path), nil
}

// UpdateFile actualiza un archivo usando Git local
func UpdateFile(config types.GitConfig, path, content string) (string, error) {
	workingDir := GetEffectiveWorkingDir(config)
	
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	// Actualizar archivo
	fullPath := filepath.Join(workingDir, path)
	err := os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("error actualizando archivo: %v", err)
	}

	return fmt.Sprintf("Archivo actualizado: %s\nUsa 'git_add' y 'git_commit' para confirmar cambios", path), nil
}

// SetWorkspace configura el directorio de trabajo para operaciones Git
func SetWorkspace(config *types.GitConfig, workspacePath string) (string, error) {
	// Verificar que el directorio existe
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return "", fmt.Errorf("directorio no existe: %s", workspacePath)
	}

	// Verificar que es un repositorio Git
	gitPath := filepath.Join(workspacePath, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return "", fmt.Errorf("no es un repositorio Git: %s", workspacePath)
	}

	// Verificar que git está disponible
	if _, err := exec.LookPath("git"); err != nil {
		return "", fmt.Errorf("Git no está disponible en el sistema")
	}

	// Configurar workspace
	config.WorkspacePath = workspacePath
	config.IsGitRepo = true
	config.RepoPath = workspacePath
	config.HasGit = true

	// Cambiar temporalmente al directorio para obtener info
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workspacePath)

	// Obtener URL remota
	if output, err := exec.Command("git", "remote", "get-url", "origin").Output(); err == nil {
		config.RemoteURL = strings.TrimSpace(string(output))
	}

	// Obtener rama actual
	if output, err := exec.Command("git", "branch", "--show-current").Output(); err == nil {
		config.CurrentBranch = strings.TrimSpace(string(output))
	}

	return fmt.Sprintf("✅ Workspace configurado: %s\n🌿 Rama: %s\n🔗 Remote: %s", 
		workspacePath, config.CurrentBranch, config.RemoteURL), nil
}

// GetEffectiveWorkingDir retorna el directorio de trabajo efectivo
func GetEffectiveWorkingDir(config types.GitConfig) string {
	if config.WorkspacePath != "" {
		return config.WorkspacePath
	}
	if config.RepoPath != "" {
		return config.RepoPath
	}
	pwd, _ := os.Getwd()
	return pwd
}

// GetFileSHA obtiene el SHA de un archivo específico
func GetFileSHA(config types.GitConfig, filePath string) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	workingDir := GetEffectiveWorkingDir(config)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	// Obtener SHA del archivo
	cmd := exec.Command("git", "rev-parse", fmt.Sprintf("HEAD:%s", filePath))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error obteniendo SHA del archivo %s: %v", filePath, err)
	}

	sha := strings.TrimSpace(string(output))
	return fmt.Sprintf("📄 Archivo: %s\n🔑 SHA: %s\n📁 Directorio: %s", filePath, sha, workingDir), nil
}

// GetLastCommitSHA obtiene el SHA del último commit
func GetLastCommitSHA(config types.GitConfig) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	workingDir := GetEffectiveWorkingDir(config)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error obteniendo SHA del commit: %v", err)
	}

	sha := strings.TrimSpace(string(output))
	return fmt.Sprintf("🔑 Último commit SHA: %s\n📁 Directorio: %s", sha, workingDir), nil
}

// GetFileContent obtiene el contenido de un archivo desde Git
func GetFileContent(config types.GitConfig, filePath, ref string) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	workingDir := GetEffectiveWorkingDir(config)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	if ref == "" {
		ref = "HEAD"
	}

	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", ref, filePath))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error obteniendo contenido del archivo %s en %s: %v", filePath, ref, err)
	}

	return fmt.Sprintf("📄 Archivo: %s\n🌿 Ref: %s\n📝 Contenido:\n%s", filePath, ref, string(output)), nil
}

// GetChangedFiles obtiene lista de archivos modificados
func GetChangedFiles(config types.GitConfig, staged bool) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	workingDir := GetEffectiveWorkingDir(config)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	var cmd *exec.Cmd
	if staged {
		cmd = exec.Command("git", "diff", "--cached", "--name-only")
	} else {
		cmd = exec.Command("git", "diff", "--name-only")
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error obteniendo archivos modificados: %v", err)
	}

	status := "working directory"
	if staged {
		status = "staging area"
	}

	return fmt.Sprintf("📁 Directorio: %s\n📋 Archivos modificados (%s):\n%s", workingDir, status, string(output)), nil
}

// ValidateRepository verifica si el directorio es un repositorio Git válido
func ValidateRepository(path string) (string, error) {
	gitPath := filepath.Join(path, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return "", fmt.Errorf("no es un repositorio Git: %s", path)
	}

	// Verificar si Git está disponible
	if _, err := exec.LookPath("git"); err != nil {
		return "", fmt.Errorf("Git no está disponible en el sistema")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(path)

	// Obtener información del repositorio
	cmd := exec.Command("git", "remote", "get-url", "origin")
	remoteOutput, _ := cmd.Output()
	
	cmd = exec.Command("git", "branch", "--show-current")
	branchOutput, _ := cmd.Output()

	return fmt.Sprintf("✅ Repositorio Git válido: %s\n🌿 Rama: %s\n🔗 Remote: %s", 
		path, 
		strings.TrimSpace(string(branchOutput)), 
		strings.TrimSpace(string(remoteOutput))), nil
}

// ListFiles lista todos los archivos en el repositorio
func ListFiles(config types.GitConfig, ref string) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	workingDir := GetEffectiveWorkingDir(config)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workingDir)

	if ref == "" {
		ref = "HEAD"
	}

	cmd := exec.Command("git", "ls-tree", "--name-only", "-r", ref)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error listando archivos en %s: %v", ref, err)
	}

	files := strings.TrimSpace(string(output))
	fileCount := strings.Count(files, "\n") + 1
	if files == "" {
		fileCount = 0
	}

	return fmt.Sprintf("📁 Directorio: %s\n🌿 Ref: %s\n📊 Total archivos: %d\n\n📄 Archivos:\n%s", workingDir, ref, fileCount, files), nil
}
