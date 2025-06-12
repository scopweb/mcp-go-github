package hybrid

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-github/v66/github"
	"github.com/jotajotape/github-go-server-mcp/internal/git"
	githubapi "github.com/jotajotape/github-go-server-mcp/internal/github"
	"github.com/jotajotape/github-go-server-mcp/internal/types"
)

// SmartCreateFile: PRIORIZA Git local, fallback a GitHub API solo si es necesario
func SmartCreateFile(gitConfig types.GitConfig, client *github.Client, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'path' requerido")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'content' requerido")
	}

	// 1. SIEMPRE intentar Git local primero (OPTIMIZACIÓN DE TOKENS)
	if gitConfig.HasGit && gitConfig.IsGitRepo {
		result, err := git.CreateFile(gitConfig, path, content)
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
	return createFileWithAPI(client, args)
}

// SmartUpdateFile: PRIORIZA Git local, fallback a GitHub API solo si es necesario  
func SmartUpdateFile(gitConfig types.GitConfig, client *github.Client, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'path' requerido")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'content' requerido")
	}

	// 1. SIEMPRE intentar Git local primero (OPTIMIZACIÓN DE TOKENS)
	if gitConfig.HasGit && gitConfig.IsGitRepo {
		// Verificar si el archivo existe localmente
		fullPath := filepath.Join(gitConfig.RepoPath, path)
		if _, err := os.Stat(fullPath); err == nil {
			result, err := git.UpdateFile(gitConfig, path, content)
			if err == nil {
				message := fmt.Sprintf("Update %s", path)
				if m, ok := args["message"].(string); ok {
					message = m
				}

				return fmt.Sprintf("✅ ARCHIVO ACTUALIZADO CON GIT LOCAL (0 tokens API)\n%s\n\n🔧 Siguiente paso: git_add('%s') -> git_commit('%s')", 
					result, path, message), nil
			}
		}
		return fmt.Sprintf("⚠️ Archivo no existe localmente o Git local falló\n⤵️ Intentando GitHub API..."), fmt.Errorf("git_local_failed")
	}

	// 2. Solo si NO hay Git local, usar GitHub API
	return updateFileWithAPI(client, args)
}

// AutoDetectContext: Detecta automáticamente si usar Git local o GitHub API
func AutoDetectContext(gitConfig types.GitConfig) string {
	if gitConfig.HasGit && gitConfig.IsGitRepo {
		return fmt.Sprintf(`🔧 MODO GIT LOCAL DETECTADO (OPTIMIZACIÓN DE TOKENS)
📁 Repo: %s
🌿 Rama: %s
🔗 Remote: %s

✅ RECOMENDACIÓN: Usar comandos git_* para operaciones sin costo de tokens
- create_file/update_file: 0 tokens (Git local)
- git_add + git_commit: 0 tokens
- git_push: Solo si necesario sincronizar

❌ EVITAR: github_* APIs a menos que sea estrictamente necesario`, 
			gitConfig.RepoPath, gitConfig.CurrentBranch, gitConfig.RemoteURL)
	}

	return `⚠️ MODO GITHUB API (COSTO TOKENS)
❌ No se detectó Git local o repositorio Git
📡 Usando GitHub API (consume tokens)

💡 OPTIMIZACIÓN: Clona el repo localmente para reducir costos`
}

// createFileWithAPI: Función auxiliar para GitHub API
func createFileWithAPI(client *github.Client, args map[string]interface{}) (string, error) {
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

	result, err := githubapi.CreateFile(client, context.Background(), owner, repo, path, content, message, branch)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("📡 ARCHIVO CREADO CON GITHUB API (tokens consumidos)\n%s", result), nil
}

// updateFileWithAPI: Función auxiliar para GitHub API
func updateFileWithAPI(client *github.Client, args map[string]interface{}) (string, error) {
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

	result, err := githubapi.UpdateFile(client, context.Background(), owner, repo, path, content, message, sha, branch)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("📡 ARCHIVO ACTUALIZADO CON GITHUB API (tokens consumidos)\n%s", result), nil
}

// CreateFile crea un archivo usando Git local si está disponible, sino GitHub API
func CreateFile(gitConfig types.GitConfig, client *github.Client, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'path' requerido")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'content' requerido")
	}

	// Si tenemos Git local, usar workflow Git
	if gitConfig.HasGit && gitConfig.IsGitRepo {
		result, err := git.CreateFile(gitConfig, path, content)
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

	return githubapi.CreateFile(client, context.Background(), owner, repo, path, content, message, branch)
}

// UpdateFile actualiza un archivo usando Git local si está disponible, sino GitHub API
func UpdateFile(gitConfig types.GitConfig, client *github.Client, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'path' requerido")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("parámetro 'content' requerido")
	}

	// Si tenemos Git local, usar workflow Git
	if gitConfig.HasGit && gitConfig.IsGitRepo {
		result, err := git.UpdateFile(gitConfig, path, content)
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

	return githubapi.UpdateFile(client, context.Background(), owner, repo, path, content, message, sha, branch)
}
