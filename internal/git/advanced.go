package git

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jotajotape/github-go-server-mcp/internal/types"
)

// LogAnalysis muestra el historial de commits con análisis
func LogAnalysis(config types.GitConfig, limit string) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(config.RepoPath)

	if limit == "" {
		limit = "20"
	}

	result := map[string]interface{}{}

	// Log gráfico
	cmd := exec.Command("git", "log", "--graph", "--oneline", "--decorate", "--all", "-"+limit)
	if output, err := cmd.Output(); err == nil {
		result["graphLog"] = strings.TrimSpace(string(output))
	}

	// Estadísticas de commits por autor
	cmd = exec.Command("git", "shortlog", "-sn", "--all")
	if output, err := cmd.Output(); err == nil {
		result["authorStats"] = strings.TrimSpace(string(output))
	}

	// Últimos commits con detalles
	cmd = exec.Command("git", "log", "--pretty=format:%h|%an|%ad|%s", "--date=short", "-"+limit)
	if output, err := cmd.Output(); err == nil {
		result["recentCommits"] = strings.TrimSpace(string(output))
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}

// DiffFiles muestra archivos modificados con detalles
func DiffFiles(config types.GitConfig, staged bool) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(config.RepoPath)

	result := map[string]interface{}{}

	var cmd *exec.Cmd
	if staged {
		// Archivos en staging
		cmd = exec.Command("git", "diff", "--name-status", "--cached")
	} else {
		// Archivos modificados
		cmd = exec.Command("git", "diff", "--name-status")
	}

	if output, err := cmd.Output(); err == nil {
		result["files"] = strings.TrimSpace(string(output))
	}

	// Estadísticas de cambios
	if staged {
		cmd = exec.Command("git", "diff", "--stat", "--cached")
	} else {
		cmd = exec.Command("git", "diff", "--stat")
	}

	if output, err := cmd.Output(); err == nil {
		result["stats"] = strings.TrimSpace(string(output))
	}

	// Archivos sin seguimiento (solo si no es staged)
	if !staged {
		cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
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

// BranchList lista todas las ramas con información detallada
func BranchList(config types.GitConfig, remote bool) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(config.RepoPath)

	result := map[string]interface{}{}

	// Ramas locales
	cmd := exec.Command("git", "branch", "-v")
	if output, err := cmd.Output(); err == nil {
		result["localBranches"] = strings.TrimSpace(string(output))
	}

	if remote {
		// Ramas remotas
		cmd = exec.Command("git", "branch", "-r", "-v")
		if output, err := cmd.Output(); err == nil {
			result["remoteBranches"] = strings.TrimSpace(string(output))
		}

		// Todas las ramas
		cmd = exec.Command("git", "branch", "-a", "-v")
		if output, err := cmd.Output(); err == nil {
			result["allBranches"] = strings.TrimSpace(string(output))
		}
	}

	// Rama actual
	cmd = exec.Command("git", "branch", "--show-current")
	if output, err := cmd.Output(); err == nil {
		result["currentBranch"] = strings.TrimSpace(string(output))
	}

	// Último commit de cada rama
	cmd = exec.Command("git", "for-each-ref", "--format=%(refname:short) %(committerdate:short) %(subject)", "refs/heads/")
	if output, err := cmd.Output(); err == nil {
		result["branchCommits"] = strings.TrimSpace(string(output))
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}

// StashOperations maneja operaciones de stash
func StashOperations(config types.GitConfig, operation, name string) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(config.RepoPath)

	var cmd *exec.Cmd
	var result string

	switch operation {
	case "list":
		cmd = exec.Command("git", "stash", "list")
		if output, err := cmd.Output(); err == nil {
			result = fmt.Sprintf("Stash list:\n%s", strings.TrimSpace(string(output)))
		} else {
			result = "No hay stashes guardados"
		}

	case "push":
		if name != "" {
			cmd = exec.Command("git", "stash", "push", "-m", name)
		} else {
			cmd = exec.Command("git", "stash", "push")
		}
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Stash creado:\n%s", strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error creando stash: %v\nOutput: %s", err, output)
		}

	case "pop":
		if name != "" {
			cmd = exec.Command("git", "stash", "pop", name)
		} else {
			cmd = exec.Command("git", "stash", "pop")
		}
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Stash aplicado y eliminado:\n%s", strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error aplicando stash: %v\nOutput: %s", err, output)
		}

	case "apply":
		if name != "" {
			cmd = exec.Command("git", "stash", "apply", name)
		} else {
			cmd = exec.Command("git", "stash", "apply")
		}
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Stash aplicado (mantenido):\n%s", strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error aplicando stash: %v\nOutput: %s", err, output)
		}

	case "drop":
		if name != "" {
			cmd = exec.Command("git", "stash", "drop", name)
		} else {
			cmd = exec.Command("git", "stash", "drop")
		}
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Stash eliminado:\n%s", strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error eliminando stash: %v\nOutput: %s", err, output)
		}

	case "clear":
		cmd = exec.Command("git", "stash", "clear")
		if output, err := cmd.CombinedOutput(); err == nil {
			result = "Todos los stashes han sido eliminados"
		} else {
			return "", fmt.Errorf("error limpiando stashes: %v\nOutput: %s", err, output)
		}

	default:
		return "", fmt.Errorf("operación no válida: %s. Usa: list, push, pop, apply, drop, clear", operation)
	}

	return result, nil
}

// RemoteOperations maneja operaciones con remotos
func RemoteOperations(config types.GitConfig, operation, name, url string) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(config.RepoPath)

	var cmd *exec.Cmd
	var result string

	switch operation {
	case "list":
		cmd = exec.Command("git", "remote", "-v")
		if output, err := cmd.Output(); err == nil {
			result = fmt.Sprintf("Remotos configurados:\n%s", strings.TrimSpace(string(output)))
		} else {
			result = "No hay remotos configurados"
		}

	case "add":
		if name == "" || url == "" {
			return "", fmt.Errorf("nombre y URL requeridos para agregar remoto")
		}
		cmd = exec.Command("git", "remote", "add", name, url)
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Remoto '%s' agregado: %s", name, url)
		} else {
			return "", fmt.Errorf("error agregando remoto: %v\nOutput: %s", err, output)
		}

	case "remove":
		if name == "" {
			return "", fmt.Errorf("nombre del remoto requerido")
		}
		cmd = exec.Command("git", "remote", "remove", name)
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Remoto '%s' eliminado", name)
		} else {
			return "", fmt.Errorf("error eliminando remoto: %v\nOutput: %s", err, output)
		}

	case "show":
		if name == "" {
			name = "origin"
		}
		cmd = exec.Command("git", "remote", "show", name)
		if output, err := cmd.Output(); err == nil {
			result = fmt.Sprintf("Información del remoto '%s':\n%s", name, strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error mostrando remoto: %v", err)
		}

	case "fetch":
		if name == "" {
			cmd = exec.Command("git", "fetch", "--all")
			result = "Fetching desde todos los remotos"
		} else {
			cmd = exec.Command("git", "fetch", name)
			result = fmt.Sprintf("Fetching desde '%s'", name)
		}
		if output, err := cmd.CombinedOutput(); err == nil {
			result += fmt.Sprintf(":\n%s", strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error en fetch: %v\nOutput: %s", err, output)
		}

	default:
		return "", fmt.Errorf("operación no válida: %s. Usa: list, add, remove, show, fetch", operation)
	}

	return result, nil
}

// TagOperations maneja operaciones con tags
func TagOperations(config types.GitConfig, operation, tagName, message string) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(config.RepoPath)

	var cmd *exec.Cmd
	var result string

	switch operation {
	case "list":
		cmd = exec.Command("git", "tag", "-l", "--sort=-version:refname")
		if output, err := cmd.Output(); err == nil {
			result = fmt.Sprintf("Tags disponibles:\n%s", strings.TrimSpace(string(output)))
		} else {
			result = "No hay tags creados"
		}

	case "create":
		if tagName == "" {
			return "", fmt.Errorf("nombre del tag requerido")
		}
		if message != "" {
			cmd = exec.Command("git", "tag", "-a", tagName, "-m", message)
		} else {
			cmd = exec.Command("git", "tag", tagName)
		}
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Tag '%s' creado", tagName)
		} else {
			return "", fmt.Errorf("error creando tag: %v\nOutput: %s", err, output)
		}

	case "delete":
		if tagName == "" {
			return "", fmt.Errorf("nombre del tag requerido")
		}
		cmd = exec.Command("git", "tag", "-d", tagName)
		if output, err := cmd.CombinedOutput(); err == nil {
			result = fmt.Sprintf("Tag '%s' eliminado localmente", tagName)
		} else {
			return "", fmt.Errorf("error eliminando tag: %v\nOutput: %s", err, output)
		}

	case "push":
		if tagName == "" {
			cmd = exec.Command("git", "push", "origin", "--tags")
			result = "Todos los tags enviados al remoto"
		} else {
			cmd = exec.Command("git", "push", "origin", tagName)
			result = fmt.Sprintf("Tag '%s' enviado al remoto", tagName)
		}
		if output, err := cmd.CombinedOutput(); err == nil {
			result += fmt.Sprintf(":\n%s", strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error enviando tags: %v\nOutput: %s", err, output)
		}

	case "show":
		if tagName == "" {
			return "", fmt.Errorf("nombre del tag requerido")
		}
		cmd = exec.Command("git", "show", tagName)
		if output, err := cmd.Output(); err == nil {
			result = fmt.Sprintf("Información del tag '%s':\n%s", tagName, strings.TrimSpace(string(output)))
		} else {
			return "", fmt.Errorf("error mostrando tag: %v", err)
		}

	default:
		return "", fmt.Errorf("operación no válida: %s. Usa: list, create, delete, push, show", operation)
	}

	return result, nil
}

// CleanOperations operaciones de limpieza del repositorio
func CleanOperations(config types.GitConfig, operation string, dryRun bool) (string, error) {
	if !config.HasGit || !config.IsGitRepo {
		return "", fmt.Errorf("Git no disponible o no es un repositorio Git")
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(config.RepoPath)

	var cmd *exec.Cmd
	var result string

	switch operation {
	case "untracked":
		if dryRun {
			cmd = exec.Command("git", "clean", "-n")
			result = "Vista previa - archivos que se eliminarían:"
		} else {
			cmd = exec.Command("git", "clean", "-f")
			result = "Archivos sin seguimiento eliminados:"
		}

	case "untracked_dirs":
		if dryRun {
			cmd = exec.Command("git", "clean", "-n", "-d")
			result = "Vista previa - archivos y directorios que se eliminarían:"
		} else {
			cmd = exec.Command("git", "clean", "-f", "-d")
			result = "Archivos y directorios sin seguimiento eliminados:"
		}

	case "ignored":
		if dryRun {
			cmd = exec.Command("git", "clean", "-n", "-X")
			result = "Vista previa - archivos ignorados que se eliminarían:"
		} else {
			cmd = exec.Command("git", "clean", "-f", "-X")
			result = "Archivos ignorados eliminados:"
		}

	case "all":
		if dryRun {
			cmd = exec.Command("git", "clean", "-n", "-d", "-x")
			result = "Vista previa - todos los archivos sin seguimiento que se eliminarían:"
		} else {
			cmd = exec.Command("git", "clean", "-f", "-d", "-x")
			result = "Todos los archivos sin seguimiento eliminados:"
		}

	default:
		return "", fmt.Errorf("operación no válida: %s. Usa: untracked, untracked_dirs, ignored, all", operation)
	}

	if output, err := cmd.Output(); err == nil {
		cleanResult := strings.TrimSpace(string(output))
		if cleanResult == "" {
			result += "\nNo hay archivos para procesar"
		} else {
			result += fmt.Sprintf("\n%s", cleanResult)
		}
	} else {
		return "", fmt.Errorf("error en limpieza: %v", err)
	}

	return result, nil
}
