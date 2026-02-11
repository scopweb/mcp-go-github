package git

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

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

	restore, err := c.enterWorkingDir()
	if err != nil {
		result["message"] = err.Error()
		output, _ := json.MarshalIndent(result, "", "  ")
		return string(output), nil
	}
	defer restore()

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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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
				return "", fmt.Errorf("conflicts detectados durante pull. Divergencia: %d commits locales, %d commits remotos. Resuelve manualmente o usa PullWithStrategy con 'rebase'", divergenceInfo["aheadCount"].(int), divergenceInfo["behindCount"].(int))
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
		if _, err := fmt.Sscanf(parts[0], "%d", &aheadCount); err != nil {
			// Parsing failed, keep default values
			_ = err
		}
		if _, err := fmt.Sscanf(parts[1], "%d", &behindCount); err != nil {
			// Parsing failed, keep default values
			_ = err
		}

		result["aheadCount"] = aheadCount
		result["behindCount"] = behindCount
		result["isDiverged"] = aheadCount > 0 && behindCount > 0
		result["canFastForward"] = aheadCount == 0 && behindCount > 0
	}

	return result, nil
}

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

