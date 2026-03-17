package hybrid

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jotajotape/github-go-server-mcp/pkg/interfaces"
)

// stat is a variable that can be replaced in tests for mocking file existence
var stat = os.Stat

// readFile is a variable that can be replaced in tests for mocking file reading
var readFile = os.ReadFile

// SmartCreateFile: PRIORIZA Git local, fallback a GitHub API solo si es necesario
func SmartCreateFile(gitOps interfaces.GitOperations, githubOps interfaces.GitHubOperations, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'path' requerido")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'content' requerido")
	}

	// 1. SIEMPRE intentar Git local primero (OPTIMIZACIÓN DE TOKENS)
	if gitOps.HasGit() && gitOps.IsGitRepo() {
		result, err := gitOps.CreateFile(path, content)
		if err == nil {
			message := fmt.Sprintf("Add %s", path)
			if m, ok := args["message"].(string); ok {
				message = m
			}

			return fmt.Sprintf("✅ ARCHIVO CREADO CON GIT LOCAL (0 tokens API)\n%s\n\n🔧 Siguiente paso: git_add('%s') -> git_commit('%s')",
				result, path, message), nil
		}
		// Si falla Git local, continuar con API
		return fmt.Sprintf("⚠️ Git local falló: %v\n⤵️ Intentando GitHub API...", err), fmt.Errorf("git_local_failed")
	}

	// 2. Solo si NO hay Git local, usar GitHub API
	return createFileWithAPI(githubOps, args)
}

// SmartUpdateFile: PRIORIZA Git local, fallback a GitHub API solo si es necesario
func SmartUpdateFile(gitOps interfaces.GitOperations, githubOps interfaces.GitHubOperations, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'path' requerido")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'content' requerido")
	}

	// 1. SIEMPRE intentar Git local primero (OPTIMIZACIÓN DE TOKENS)
	if gitOps.HasGit() && gitOps.IsGitRepo() {
		// Verificar si el archivo existe localmente
		fullPath := filepath.Join(gitOps.GetRepoPath(), path)
		if _, err := stat(fullPath); err == nil {
			// El SHA no es necesario para la operación local, pero lo mantenemos en la firma por consistencia
			result, err := gitOps.UpdateFile(path, content, "")
			if err == nil {
				message := fmt.Sprintf("Update %s", path)
				if m, ok := args["message"].(string); ok {
					message = m
				}

				return fmt.Sprintf("✅ ARCHIVO ACTUALIZADO CON GIT LOCAL (0 tokens API)\n%s\n\n🔧 Siguiente paso: git_add('%s') -> git_commit('%s')",
					result, path, message), nil
			}
			// Si UpdateFile falla, también consideramos que es un fallo local
		}
		// Si Stat falla (el archivo no existe) o si UpdateFile falla, devolvemos el error de fallback
		return "⚠️ Archivo no existe localmente o Git local falló\n⤵️ Intentando GitHub API...", fmt.Errorf("git_local_failed")
	}

	// 2. Solo si NO hay Git local, usar GitHub API
	return updateFileWithAPI(githubOps, args)
}

// AutoDetectContext: Detecta automáticamente si usar Git local o GitHub API
func AutoDetectContext(gitOps interfaces.GitOperations) string {
	if gitOps.HasGit() && gitOps.IsGitRepo() {
		return fmt.Sprintf(`🔧 MODO GIT LOCAL DETECTADO (OPTIMIZACIÓN DE TOKENS)
📁 Repo: %s
🌿 Rama: %s
🔗 Remote: %s

✅ RECOMENDACIÓN: Usar commands git_* para operaciones sin costo de tokens
- create_file/update_file: 0 tokens (Git local)
- git_add + git_commit: 0 tokens
- git_push: Solo si necesario sincronizar

❌ EVITAR: github_* APIs a menos que sea estrictamente necesario`,
			gitOps.GetRepoPath(), gitOps.GetCurrentBranch(), gitOps.GetRemoteURL())
	}

	return `⚠️ MODO GITHUB API (COSTO TOKENS)
❌ No se detectó Git local o repositorio Git
📡 Usando GitHub API (consume tokens)

💡 OPTIMIZACIÓN: Clona el repo localmente para reducir costos`
}

// createFileWithAPI: Función auxiliar para GitHub API
func createFileWithAPI(githubOps interfaces.GitHubOperations, args map[string]interface{}) (string, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'owner' requerido para GitHub API")
	}

	repo, ok := args["repo"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'repo' requerido para GitHub API")
	}

	path, _ := args["path"].(string)
	content, _ := args["content"].(string)
	message, ok := args["message"].(string)
	if !ok {
		message = fmt.Sprintf("Add %s", path)
	}

	branch := "main"
	if b, ok := args["branch"].(string); ok {
		branch = b
	}

	result, err := githubOps.CreateFile(context.Background(), owner, repo, path, content, message, branch)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("📡 ARCHIVO CREADO CON GITHUB API (tokens consumidos)\nCommit: %s", result.Commit.GetSHA()), nil
}

// updateFileWithAPI: Función auxiliar para GitHub API
func updateFileWithAPI(githubOps interfaces.GitHubOperations, args map[string]interface{}) (string, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'owner' requerido para GitHub API")
	}

	repo, ok := args["repo"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'repo' requerido para GitHub API")
	}

	sha, ok := args["sha"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'sha' requerido para GitHub API")
	}

	path, _ := args["path"].(string)
	content, _ := args["content"].(string)
	message, ok := args["message"].(string)
	if !ok {
		message = fmt.Sprintf("Update %s", path)
	}

	branch := "main"
	if b, ok := args["branch"].(string); ok {
		branch = b
	}

	result, err := githubOps.UpdateFile(context.Background(), owner, repo, path, content, message, sha, branch)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("📡 ARCHIVO ACTUALIZADO CON GITHUB API (tokens consumidos)\nCommit: %s", result.Commit.GetSHA()), nil
}

// CreateFile crea un archivo usando Git local si está disponible, sino GitHub API
func CreateFile(gitOps interfaces.GitOperations, githubOps interfaces.GitHubOperations, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'path' requerido")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'content' requerido")
	}

	// Si tenemos Git local, usar workflow Git
	if gitOps.HasGit() && gitOps.IsGitRepo() {
		result, err := gitOps.CreateFile(path, content)
		if err != nil {
			return "", err
		}

		message := fmt.Sprintf("Add %s", path)
		if m, ok := args["message"].(string); ok {
			message = m
		}

		return fmt.Sprintf("%s\nSugerencia: git_add('%s') -> git_commit('%s')", result, path, message), nil
	}

	// Fallback a GitHub API
	owner, ok := args["owner"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'owner' requerido para API")
	}

	repo, ok := args["repo"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'repo' requerido para API")
	}

	message, ok := args["message"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'message' requerido para API")
	}

	branch := "main"
	if b, ok := args["branch"].(string); ok {
		branch = b
	}

	_, err := githubOps.CreateFile(context.Background(), owner, repo, path, content, message, branch)
	if err != nil {
		return "", err
	}
	return "File created via GitHub API", nil
}

// UpdateFile actualiza un archivo usando Git local si está disponible, sino GitHub API
func UpdateFile(gitOps interfaces.GitOperations, githubOps interfaces.GitHubOperations, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'path' requerido")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'content' requerido")
	}

	// Si tenemos Git local, usar workflow Git
	if gitOps.HasGit() && gitOps.IsGitRepo() {
		// El SHA no es necesario para la operación local
		result, err := gitOps.UpdateFile(path, content, "")
		if err != nil {
			return "", err
		}

		message := fmt.Sprintf("Update %s", path)
		if m, ok := args["message"].(string); ok {
			message = m
		}

		return fmt.Sprintf("%s\nSugerencia: git_add('%s') -> git_commit('%s')", result, path, message), nil
	}

	// Fallback a GitHub API
	owner, ok := args["owner"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'owner' requerido para API")
	}

	repo, ok := args["repo"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'repo' requerido para API")
	}

	message, ok := args["message"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'message' requerido para API")
	}

	sha, ok := args["sha"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'sha' requerido para API")
	}

	branch := "main"
	if b, ok := args["branch"].(string); ok {
		branch = b
	}

	_, err := githubOps.UpdateFile(context.Background(), owner, repo, path, content, message, sha, branch)
	if err != nil {
		return "", err
	}
	return "File updated via GitHub API", nil
}

// PushFiles escribe múltiples archivos y realiza git add/commit/push en una sola llamada.
// Soporta 3 modos:
//   - files con content: contenido inline (modo original)
//   - files con source_path: copia desde ruta local sin enviar contenido por el AI
//   - paths: archivos que ya existen en el workspace, solo git add/commit/push
func PushFiles(gitOps interfaces.GitOperations, args map[string]interface{}) (string, error) {
	if !gitOps.HasGit() || !gitOps.IsGitRepo() {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git. Usa 'git_set_workspace' para configurarlo")
	}

	commitMessage, ok := args["message"].(string)
	if !ok || strings.TrimSpace(commitMessage) == "" {
		return "", fmt.Errorf("parámetro 'message' requerido para commit")
	}
	commitMessage = strings.TrimSpace(commitMessage)

	branch := strings.TrimSpace(gitOps.GetCurrentBranch())
	if b, ok := args["branch"].(string); ok && strings.TrimSpace(b) != "" {
		branch = strings.TrimSpace(b)
	}
	if branch == "" {
		branch = "main"
	}

	repoPath := gitOps.GetRepoPath()
	if repoPath == "" {
		return "", fmt.Errorf("ruta del repositorio no configurada; ejecuta 'git_set_workspace' antes de push_files")
	}

	rawFiles, _ := args["files"].([]interface{})
	rawPaths, _ := args["paths"].([]interface{})

	if len(rawFiles) == 0 && len(rawPaths) == 0 {
		return "", fmt.Errorf("se requiere 'files' o 'paths' (o ambos)")
	}

	var created []string
	var updated []string
	var staged []string
	totalProcessed := 0

	// Process files array (content or source_path mode)
	for idx, file := range rawFiles {
		fileMap, ok := file.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("files[%d] debe ser un objeto con 'path' y ('content' o 'source_path')", idx)
		}

		path, ok := fileMap["path"].(string)
		if !ok || strings.TrimSpace(path) == "" {
			return "", fmt.Errorf("files[%d].path requerido", idx)
		}

		// Determine content: from inline content or source_path
		var content string
		if c, ok := fileMap["content"].(string); ok {
			content = c
		} else if srcPath, ok := fileMap["source_path"].(string); ok && strings.TrimSpace(srcPath) != "" {
			// Read content from source_path on disk
			data, err := readFile(srcPath)
			if err != nil {
				return "", fmt.Errorf("files[%d]: error leyendo source_path '%s': %w", idx, srcPath, err)
			}
			content = string(data)
		} else {
			return "", fmt.Errorf("files[%d] requiere 'content' o 'source_path'", idx)
		}

		fullPath := filepath.Join(repoPath, path)
		if _, err := stat(fullPath); err == nil {
			if _, err := gitOps.UpdateFile(path, content, ""); err != nil {
				return "", fmt.Errorf("error actualizando %s: %w", path, err)
			}
			updated = append(updated, path)
		} else if errors.Is(err, os.ErrNotExist) {
			if _, err := gitOps.CreateFile(path, content); err != nil {
				return "", fmt.Errorf("error creando %s: %w", path, err)
			}
			created = append(created, path)
		} else {
			return "", fmt.Errorf("error accediendo a %s: %w", path, err)
		}
		totalProcessed++
	}

	// Process paths array (files already in workspace, just stage them)
	for idx, p := range rawPaths {
		path, ok := p.(string)
		if !ok || strings.TrimSpace(path) == "" {
			return "", fmt.Errorf("paths[%d] debe ser un string no vacío", idx)
		}
		fullPath := filepath.Join(repoPath, path)
		if _, err := stat(fullPath); err != nil {
			return "", fmt.Errorf("paths[%d]: archivo '%s' no existe en el workspace", idx, path)
		}
		staged = append(staged, path)
		totalProcessed++
	}

	// Git add: use specific paths if only paths mode, otherwise -A
	var addResult string
	var err error
	if len(rawFiles) == 0 && len(staged) > 0 {
		// Only paths mode: stage specific files
		addResult, err = gitOps.Add(strings.Join(staged, " "))
		if err != nil {
			return "", fmt.Errorf("git add falló: %w", err)
		}
	} else {
		addResult, err = gitOps.Add("-A")
		if err != nil {
			return "", fmt.Errorf("git add falló: %w", err)
		}
	}

	commitResult, err := gitOps.Commit(commitMessage)
	if err != nil {
		return "", fmt.Errorf("git commit falló: %w", err)
	}

	pushResult, err := gitOps.Push(branch)
	if err != nil {
		return "", fmt.Errorf("git push a '%s' falló: %w", branch, err)
	}

	var summary []string
	summary = append(summary, fmt.Sprintf("✅ %d archivo(s) procesados", totalProcessed))
	if len(created) > 0 {
		summary = append(summary, fmt.Sprintf("🆕 Creados: %s", strings.Join(created, ", ")))
	}
	if len(updated) > 0 {
		summary = append(summary, fmt.Sprintf("✏️ Actualizados: %s", strings.Join(updated, ", ")))
	}
	if len(staged) > 0 {
		summary = append(summary, fmt.Sprintf("📂 Staged: %s", strings.Join(staged, ", ")))
	}

	summary = append(summary,
		fmt.Sprintf("📌 git add: %s", addResult),
		fmt.Sprintf("📌 git commit: %s", commitResult),
		fmt.Sprintf("📌 git push (%s): %s", branch, pushResult),
	)

	return strings.Join(summary, "\n"), nil
}
