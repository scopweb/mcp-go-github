package git

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jotajotape/github-go-server-mcp/pkg/types"
)

func (c *Client) Checkout(branch string, create bool) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	if branch == "" {
		return "", fmt.Errorf("nombre de rama requerido")
	}

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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
			if _, stashErr := c.Stash("pop", stashName); stashErr != nil {
				// Log error but don't fail the operation
				_ = stashErr
			}
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

func (c *Client) BranchList(remote bool) ([]types.BranchInfo, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return nil, fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	workingDir := c.getEffectiveWorkingDir()
	restore, err := enterDir(workingDir)
	if err != nil {
		return nil, err
	}
	defer restore()

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

func (c *Client) CheckoutRemote(remoteBranch string, localBranch string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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
			return "", fmt.Errorf("conflicts de merge detectados. Usa 'ConflictStatus' para ver detalles y 'ResolveConflicts' para resolverlos: %s", output)
		}
		return "", fmt.Errorf("error en merge: %v, Output: %s", err, output)
	}

	return fmt.Sprintf("Merge exitoso: %s -> %s", sourceBranch, targetBranch), nil
}

func (c *Client) Rebase(branch string) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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
			return "", fmt.Errorf("conflicts de rebase detectados. Usa 'ConflictStatus' para ver detalles: %s", output)
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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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
			return "", fmt.Errorf("conflicts detectados durante pull con estrategia %s: %s", strategy, output)
		}
		return "", fmt.Errorf("error en pull con estrategia %s: %v, Output: %s", strategy, err, output)
	}

	return fmt.Sprintf("Pull con estrategia '%s' exitoso en rama: %s", strategy, branch), nil
}

func (c *Client) ForcePush(branch string, force bool) (string, error) {
	if !c.Config.HasGit || !c.Config.IsGitRepo {
		return "", fmt.Errorf("git no disponible o no es un repositorio Git")
	}

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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
				return "", fmt.Errorf("conflicts detectados durante sincronización: %s", output)
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

	restore, err := c.enterWorkingDir()
	if err != nil {
		return "", err
	}
	defer restore()

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
		return "", fmt.Errorf("error detectando conflicts: %v", err)
	} else if conflicts != "" {
		return "", fmt.Errorf("conflicts potenciales detectados: %s", conflicts)
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
		if _, resetErr := resetCmd.CombinedOutput(); resetErr != nil {
			// Log error but continue with original error
			_ = resetErr
		}

		return "", fmt.Errorf("safe merge falló, rollback realizado: %v, Output: %s", err, output)
	}

	return fmt.Sprintf("Safe merge exitoso: %s -> %s (backup creado: %s)", source, target, backupName), nil
}

// Conflict management

