package git

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (c *Client) LogAnalysis(limit string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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

func (c *Client) Stash(operation, name string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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
			result = "Vista previa - archivos y directories que se eliminarían:"
		} else {
			cmd = c.executor.Command("git", "clean", "-f", "-d")
			result = "Archivos y directories sin seguimiento eliminados:"
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
func (c *Client) ConflictStatus() (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

	// Paso 1: Obtener lista de archivos en conflicto
	conflictedFiles, err := c.getConflictedFiles()
	if err != nil {
		return "", fmt.Errorf("error obteniendo archivos en conflicto: %v", err)
	}

	if len(conflictedFiles) == 0 {
		return "No hay conflicts para resolver", nil
	}

	var cmd cmdWrapper
	var result string

	switch strategy {
	case "theirs":
		// Aceptar versión remota para todos los conflicts
		cmd = c.executor.Command("git", "checkout", "--theirs", ".")
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error aceptando cambios remotos: %v, Output: %s", err, output)
		}

		// Agregar archivos resueltos
		addCmd := c.executor.Command("git", "add", ".")
		if output, err := addCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error agregando archivos resueltos: %v, Output: %s", err, output)
		}

		result = fmt.Sprintf("Conflicts resueltos en %d archivo(s) aceptando versión remota (theirs)", len(conflictedFiles))

	case "ours":
		// Aceptar versión local para todos los conflicts
		cmd = c.executor.Command("git", "checkout", "--ours", ".")
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error aceptando cambios locales: %v, Output: %s", err, output)
		}

		// Agregar archivos resueltos
		addCmd := c.executor.Command("git", "add", ".")
		if output, err := addCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error agregando archivos resueltos: %v, Output: %s", err, output)
		}

		result = fmt.Sprintf("Conflicts resueltos en %d archivo(s) aceptando versión local (ours)", len(conflictedFiles))

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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return false, err
	}
	defer restore()

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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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

// Helper types para manejo de conflicts
type ConflictMarker struct {
	Ours   string `json:"ours"`
	Theirs string `json:"theirs"`
	Base   string `json:"base,omitempty"`
}

type ConflictDetails struct {
	File       string           `json:"file"`
	HasBase    bool             `json:"hasBase"`
	Markers    []ConflictMarker `json:"markers"`
	Content    string           `json:"content,omitempty"`
	Suggestion string           `json:"suggestion"`
}

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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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

// FASE 2: Gestión de Conflicts

// ShowConflict muestra los detalles de un conflicto en un archivo específico
func (c *Client) ShowConflict(filePath string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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
		File:       filePath,
		Content:    content,
		Suggestion: "Resuelve manualmente o usa ResolveFile con estrategia 'ours' o 'theirs'",
	}

	lines := strings.Split(content, "\n")
	var currentMarker ConflictMarker
	inOurs := false
	inBase := false
	inTheirs := false
	hasConflict := false

	for _, line := range lines {
		var markerType string
		switch {
		case strings.HasPrefix(line, "<<<<<<<"):
			markerType = "start"
		case strings.HasPrefix(line, "======="):
			markerType = "separator"
		case strings.HasPrefix(line, "||||||| "):
			markerType = "base"
		case strings.HasPrefix(line, ">>>>>>>"):
			markerType = "end"
		default:
			markerType = "content"
		}

		switch markerType {
		case "start":
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
		case "separator":
			// Separador entre ours y theirs (o base y theirs)
			if inOurs {
				inOurs = false
				inTheirs = true
			} else if inBase {
				inBase = false
				inTheirs = true
			}
		case "base":
			// Separador de base (merge 3-way)
			inOurs = false
			inBase = true
			details.HasBase = true
		case "end":
			// Fin de conflicto
			inTheirs = false
			inOurs = false
			inBase = false
		case "content":
			// Acumular contenido
			switch {
			case inOurs:
				if currentMarker.Ours != "" {
					currentMarker.Ours += "\n"
				}
				currentMarker.Ours += line
			case inBase:
				if currentMarker.Base != "" {
					currentMarker.Base += "\n"
				}
				currentMarker.Base += line
			case inTheirs:
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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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
		if err := os.WriteFile(fullPath, []byte(*customContent), 0600); err != nil {
			return "", fmt.Errorf("error escribiendo contenido personalizado: %v", err)
		}
		result = "Archivo resuelto con contenido personalizado"
	}

	// Ejecutar command si no es manual
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
