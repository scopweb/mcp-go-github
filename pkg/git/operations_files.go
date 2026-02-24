package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// normalizeWindowsPath convierte rutas WSL (/mnt/c/...) a rutas Windows (C:\...)
// y normaliza separadores de ruta
func normalizeWindowsPath(path string) string {
	// Detectar formato WSL: /mnt/<letra>/<resto>
	if strings.HasPrefix(path, "/mnt/") {
		parts := strings.SplitN(path[5:], "/", 2)
		if len(parts) >= 1 && len(parts[0]) == 1 {
			drive := strings.ToUpper(parts[0])
			rest := ""
			if len(parts) == 2 {
				rest = filepath.FromSlash(parts[1])
			}
			return drive + ":\\" + rest
		}
	}
	// Normalizar separadores de ruta
	return filepath.Clean(filepath.FromSlash(path))
}

func (c *Client) CreateFile(path, content string) (string, error) {
	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

	fullPath := filepath.Join(c.Config.RepoPath, path)
	err = os.MkdirAll(filepath.Dir(fullPath), 0755)
	if err != nil {
		return "", fmt.Errorf("error creando directorio: %v", err)
	}

	err = os.WriteFile(fullPath, []byte(content), 0600)
	if err != nil {
		return "", fmt.Errorf("error escribiendo archivo: %v", err)
	}

	return fmt.Sprintf("Archivo creado: %s. Usa 'git_add' y 'git_commit' para confirmar cambios", path), nil
}

func (c *Client) UpdateFile(path, content, _ string) (string, error) {
	workingDir := c.getEffectiveWorkingDir()

	restore, err := enterDir(workingDir)
	if err != nil {
		return "", err
	}
	defer restore()

	fullPath := filepath.Join(workingDir, path)
	err = os.WriteFile(fullPath, []byte(content), 0600)
	if err != nil {
		return "", fmt.Errorf("error actualizando archivo: %v", err)
	}

	return fmt.Sprintf("Archivo actualizado: %s. Usa 'git_add' y 'git_commit' para confirmar cambios", path), nil
}

func (c *Client) SetWorkspace(workspacePath string) (string, error) {
	// Normalizar ruta (convierte WSL a Windows, limpia separadores)
	workspacePath = normalizeWindowsPath(workspacePath)

	// Verificar que el directorio existe (capturar todos los errores, no solo IsNotExist)
	if _, err := os.Stat(workspacePath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("directorio no existe: %s", workspacePath)
		}
		return "", fmt.Errorf("no se puede acceder al directorio '%s': %v", workspacePath, err)
	}

	// Verificar que .git existe
	gitPath := filepath.Join(workspacePath, ".git")
	if _, err := os.Stat(gitPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no es un repositorio Git válido (falta carpeta .git): %s", workspacePath)
		}
		return "", fmt.Errorf("no se puede acceder a .git en '%s': %v", workspacePath, err)
	}

	// Verificar Git disponible
	if _, err := c.executor.LookPath("git"); err != nil {
		return "", fmt.Errorf("git no está disponible en el sistema: %v", err)
	}

	// Actualizar configuración
	c.Config.WorkspacePath = workspacePath
	c.Config.IsGitRepo = true
	c.Config.RepoPath = workspacePath
	c.Config.HasGit = true

	// Cambiar al directorio y obtener info del repo
	restore, err := enterDir(workspacePath)
	if err != nil {
		return "", fmt.Errorf("error cambiando a directorio: %v", err)
	}
	defer restore()

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
	restore, err := enterDir(workingDir)
	if err != nil {
		return "", err
	}
	defer restore()

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
	restore, err := enterDir(workingDir)
	if err != nil {
		return "", err
	}
	defer restore()

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
	restore, err := enterDir(workingDir)
	if err != nil {
		return "", err
	}
	defer restore()

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
	restore, err := enterDir(workingDir)
	if err != nil {
		return "", err
	}
	defer restore()

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
	// Normalizar ruta (convierte WSL a Windows, limpia separadores)
	path = normalizeWindowsPath(path)

	// Verificar que el directorio existe
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("directorio no existe: %s", path)
		}
		return "", fmt.Errorf("no se puede acceder al directorio '%s': %v", path, err)
	}

	// Verificar que .git existe
	gitPath := filepath.Join(path, ".git")
	if _, err := os.Stat(gitPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no es un repositorio Git válido (falta carpeta .git): %s", path)
		}
		return "", fmt.Errorf("no se puede acceder a .git en '%s': %v", path, err)
	}

	// Verificar Git disponible
	if _, err := c.executor.LookPath("git"); err != nil {
		return "", fmt.Errorf("git no está disponible en el sistema: %v", err)
	}

	// Cambiar al directorio y obtener info del repo
	restore, err := enterDir(path)
	if err != nil {
		return "", fmt.Errorf("error cambiando a directorio: %v", err)
	}
	defer restore()

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
	restore, err := enterDir(workingDir)
	if err != nil {
		return "", err
	}
	defer restore()

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

