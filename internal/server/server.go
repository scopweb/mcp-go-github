package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jotajotape/github-go-server-mcp/internal/hybrid"
	"github.com/jotajotape/github-go-server-mcp/pkg/dashboard"
	"github.com/jotajotape/github-go-server-mcp/pkg/interfaces"
	"github.com/jotajotape/github-go-server-mcp/pkg/types"
)

// MCPServer representa el servidor MCP principal
type MCPServer struct {
	GithubClient   interfaces.GitHubOperations
	GitClient      interfaces.GitOperations
	AdminClient    interfaces.AdminOperations  // v3.0: Administrative operations
	Safety         *SafetyMiddleware            // v3.0: Safety filter middleware
	GitAvailable   bool                         // v3.0: Whether git binary is installed
	RawGitHubClient interface{}                 // v3.0: Raw *github.Client for file operations
}

// HandleRequest procesa las peticiones JSON-RPC del protocolo MCP
func HandleRequest(s *MCPServer, req types.JSONRPCRequest) types.JSONRPCResponse {
	id := req.ID
	if id == nil {
		id = 0
	}

	response := types.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
	}

	if req.JSONRPC != "2.0" {
		response.Error = &types.JSONRPCError{
			Code:    -32600,
			Message: "Invalid Request: jsonrpc must be '2.0'",
		}
		return response
	}

	if req.Method == "" {
		response.Error = &types.JSONRPCError{
			Code:    -32600,
			Message: "Invalid Request: method is required",
		}
		return response
	}

	switch req.Method {
	case "initialize":
		// Extract client's requested protocol version and respond with the same
		clientProtocolVersion := "2024-11-05" // Default fallback
		if version, ok := req.Params["protocolVersion"].(string); ok && version != "" {
			clientProtocolVersion = version
		}

		response.Result = map[string]interface{}{
			"protocolVersion": clientProtocolVersion,
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "github-mcp-admin-v3",
				"version": "3.0.1",
			},
		}
	case "initialized":
		response.Result = map[string]interface{}{}
	case "tools/list":
		response.Result = ListTools(s.GitAvailable)
	case "tools/call":
		result, err := CallTool(s, req.Params)
		if err != nil {
			response.Error = &types.JSONRPCError{
				Code:    -32603,
				Message: err.Error(),
			}
		} else {
			response.Result = result
		}
	default:
		response.Error = &types.JSONRPCError{
			Code:    -32601,
			Message: "Method not found",
		}
	}

	return response
}

// isGitTool returns true if the tool requires local Git binary
func isGitTool(name string) bool {
	return strings.HasPrefix(name, "git_")
}

// ListTools retorna la lista de herramientas disponibles
func ListTools(gitAvailable bool) types.ToolsListResult {
	allTools := []types.Tool{}

	// Git tools (information, basic, advanced)
	allTools = append(allTools, ListGitInfoTools()...)
	allTools = append(allTools, ListGitBasicTools()...)
	allTools = append(allTools, ListGitAdvancedTools()...)

	// Hybrid tools (Git-first, API fallback)
	allTools = append(allTools, ListHybridTools()...)

	// GitHub API tools
	allTools = append(allTools, ListGitHubAPITools()...)

	// Dashboard tools
	allTools = append(allTools, ListDashboardTools()...)

	// Response tools
	allTools = append(allTools, ListResponseTools()...)

	// Repair tools
	allTools = append(allTools, ListRepairTools()...)

	// Add administrative tools (v3.0)
	adminTools := ListAdminTools()
	allTools = append(allTools, adminTools...)

	// Add file operation tools (v3.0 - work without Git)
	fileTools := ListFileTools()
	allTools = append(allTools, fileTools...)

	// Filter out Git tools if Git is not available
	if !gitAvailable {
		var filtered []types.Tool
		for _, tool := range allTools {
			if !isGitTool(tool.Name) {
				filtered = append(filtered, tool)
			}
		}
		return types.ToolsListResult{Tools: filtered}
	}

	return types.ToolsListResult{Tools: allTools}
}

// CallTool ejecuta la herramienta solicitada
func CallTool(s *MCPServer, params map[string]interface{}) (types.ToolCallResult, error) {
	name, ok := params["name"].(string)
	if !ok {
		return types.ToolCallResult{}, fmt.Errorf("tool name required")
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	ctx := context.Background()
	var text string
	var err error

	// Check if Git tool is called without Git installed
	if isGitTool(name) && !s.GitAvailable {
		return types.ToolCallResult{
			Content: []types.Content{{Type: "text", Text: fmt.Sprintf("⚠️ Git is not installed on this system.\n\nThe tool '%s' requires a local Git binary.\n\nAlternatives:\n• Use GitHub API tools (github_*) which work without Git\n• Install Git: https://git-scm.com/downloads\n\nAvailable without Git: dashboard, repos, PRs, issues, webhooks, collaborators, branch protection, and all admin tools.", name)}},
		}, nil
	}

	switch name {
	// Herramientas Git básicas
	case "git_status":
		text, err = s.GitClient.Status()
	case "git_set_workspace":
		path, _ := arguments["path"].(string)
		text, err = s.GitClient.SetWorkspace(path)
	case "git_get_file_sha":
		path, _ := arguments["path"].(string)
		text, err = s.GitClient.GetFileSHA(path)
	case "git_get_last_commit":
		text, err = s.GitClient.GetLastCommit()
	case "git_get_file_content":
		path, _ := arguments["path"].(string)
		ref, _ := arguments["ref"].(string)
		text, err = s.GitClient.GetFileContent(path, ref)
	case "git_get_changed_files":
		staged, _ := arguments["staged"].(bool)
		text, err = s.GitClient.GetChangedFiles(staged)
	case "git_validate_repo":
		path, _ := arguments["path"].(string)
		text, err = s.GitClient.ValidateRepo(path)
	case "git_list_files":
		ref, _ := arguments["ref"].(string)
		text, err = s.GitClient.ListFiles(ref)
	case "git_init":
		path, _ := arguments["path"].(string)
		initialBranch, _ := arguments["initial_branch"].(string)
		text, err = s.GitClient.Init(path, initialBranch)
	case "git_add":
		files, _ := arguments["files"].(string)
		text, err = s.GitClient.Add(files)
	case "git_commit":
		message, _ := arguments["message"].(string)
		text, err = s.GitClient.Commit(message)
	case "git_push":
		branch, _ := arguments["branch"].(string)
		text, err = s.GitClient.Push(branch)
	case "git_pull":
		branch, _ := arguments["branch"].(string)
		text, err = s.GitClient.Pull(branch)
	case "git_checkout":
		branch, _ := arguments["branch"].(string)
		create, _ := arguments["create"].(bool)
		text, err = s.GitClient.Checkout(branch, create)

	// Herramientas Git avanzadas
	case "git_log_analysis":
		limit, _ := arguments["limit"].(string)
		text, err = s.GitClient.LogAnalysis(limit)
	case "git_diff_files":
		staged, _ := arguments["staged"].(bool)
		text, err = s.GitClient.DiffFiles(staged)
	case "git_branch_list":
		remote, _ := arguments["remote"].(bool)
		branches, branchErr := s.GitClient.BranchList(remote)
		if branchErr != nil {
			err = branchErr
		} else {
			// Convertir a JSON para una salida más estructurada
			jsonOutput, jsonErr := json.MarshalIndent(branches, "", "  ")
			if jsonErr != nil {
				err = fmt.Errorf("failed to marshal branch list: %w", jsonErr)
			} else {
				text = string(jsonOutput)
			}
		}
	case "git_stash":
		operation, _ := arguments["operation"].(string)
		name, _ := arguments["name"].(string)
		text, err = s.GitClient.Stash(operation, name)
	case "git_remote":
		operation, _ := arguments["operation"].(string)
		name, _ := arguments["name"].(string)
		url, _ := arguments["url"].(string)
		text, err = s.GitClient.Remote(operation, name, url)
	case "git_tag":
		operation, _ := arguments["operation"].(string)
		tagName, _ := arguments["tag_name"].(string)
		message, _ := arguments["message"].(string)
		text, err = s.GitClient.Tag(operation, tagName, message)
	case "git_clean":
		operation, _ := arguments["operation"].(string)
		dryRun, exists := arguments["dry_run"].(bool)
		if !exists {
			dryRun = true // default a true para seguridad
		}
		text, err = s.GitClient.Clean(operation, dryRun)

	case "git_context":
		text = hybrid.AutoDetectContext(s.GitClient)
		err = nil

	// Advanced Git Operations
	case "git_checkout_remote":
		remoteBranch, _ := arguments["remote_branch"].(string)
		localBranch, _ := arguments["local_branch"].(string)
		text, err = s.GitClient.CheckoutRemote(remoteBranch, localBranch)
	case "git_merge":
		sourceBranch, _ := arguments["source_branch"].(string)
		targetBranch, _ := arguments["target_branch"].(string)
		text, err = s.GitClient.Merge(sourceBranch, targetBranch)
	case "git_rebase":
		branch, _ := arguments["branch"].(string)
		text, err = s.GitClient.Rebase(branch)
	case "git_pull_with_strategy":
		branch, _ := arguments["branch"].(string)
		strategy, _ := arguments["strategy"].(string)
		text, err = s.GitClient.PullWithStrategy(branch, strategy)
	case "git_force_push":
		branch, _ := arguments["branch"].(string)
		force, _ := arguments["force"].(bool)
		text, err = s.GitClient.ForcePush(branch, force)
	case "git_push_upstream":
		branch, _ := arguments["branch"].(string)
		text, err = s.GitClient.PushUpstream(branch)
	case "git_sync_with_remote":
		remoteBranch, _ := arguments["remote_branch"].(string)
		text, err = s.GitClient.SyncWithRemote(remoteBranch)
	case "git_safe_merge":
		source, _ := arguments["source"].(string)
		target, _ := arguments["target"].(string)
		text, err = s.GitClient.SafeMerge(source, target)
	case "git_conflict_status":
		text, err = s.GitClient.ConflictStatus()
	case "git_resolve_conflicts":
		strategy, _ := arguments["strategy"].(string)
		text, err = s.GitClient.ResolveConflicts(strategy)
	case "git_validate_clean_state":
		clean, validateErr := s.GitClient.ValidateCleanState()
		if validateErr != nil {
			err = validateErr
		} else {
			if clean {
				text = "✅ Working directory is clean"
			} else {
				text = "⚠️ Working directory has uncommitted changes"
			}
		}
	case "git_detect_conflicts":
		sourceBranch, _ := arguments["source_branch"].(string)
		targetBranch, _ := arguments["target_branch"].(string)
		conflictInfo, detectErr := s.GitClient.DetectPotentialConflicts(sourceBranch, targetBranch)
		if detectErr != nil {
			err = detectErr
		} else {
			if conflictInfo == "" {
				text = "✅ No potential conflicts detected between branches"
			} else {
				text = fmt.Sprintf("⚠️ %s", conflictInfo)
			}
		}
	case "git_create_backup":
		name, _ := arguments["name"].(string)
		text, err = s.GitClient.CreateBackup(name)

	// Herramientas híbridas
	case "create_file":
		text, err = hybrid.SmartCreateFile(s.GitClient, s.GithubClient, arguments)
	case "update_file":
		text, err = hybrid.SmartUpdateFile(s.GitClient, s.GithubClient, arguments)

	// Herramientas API puras
	case "github_list_repos":
		listType, _ := arguments["type"].(string)
		repos, listErr := s.GithubClient.ListRepositories(ctx, listType)
		if listErr != nil {
			err = listErr
		} else {
			var repoNames []string
			for _, repo := range repos {
				repoNames = append(repoNames, repo.GetFullName())
			}
			text = fmt.Sprintf("Repositories:\n%s", strings.Join(repoNames, "\n"))
		}
	case "github_create_repo":
		name, _ := arguments["name"].(string)
		description, _ := arguments["description"].(string)
		private, _ := arguments["private"].(bool)
		repo, createErr := s.GithubClient.CreateRepository(ctx, name, description, private)
		if createErr != nil {
			err = createErr
		} else {
			text = fmt.Sprintf("Successfully created repository: %s", repo.GetFullName())
		}
	case "github_list_prs":
		owner, _ := arguments["owner"].(string)
		repo, _ := arguments["repo"].(string)
		state, _ := arguments["state"].(string)
		prs, listErr := s.GithubClient.ListPullRequests(ctx, owner, repo, state)
		if listErr != nil {
			err = listErr
		} else {
			var prInfo []string
			for _, pr := range prs {
				prInfo = append(prInfo, fmt.Sprintf("#%d: %s", pr.GetNumber(), pr.GetTitle()))
			}
			if len(prInfo) == 0 {
				text = "No pull requests found."
			} else {
				text = fmt.Sprintf("Pull Requests:\n%s", strings.Join(prInfo, "\n"))
			}
		}
	case "github_create_pr":
		owner, _ := arguments["owner"].(string)
		repo, _ := arguments["repo"].(string)
		title, _ := arguments["title"].(string)
		body, _ := arguments["body"].(string)
		head, _ := arguments["head"].(string)
		base, _ := arguments["base"].(string)
		pr, createErr := s.GithubClient.CreatePullRequest(ctx, owner, repo, title, body, head, base)
		if createErr != nil {
			err = createErr
		} else {
			text = fmt.Sprintf("Successfully created pull request #%d: %s", pr.GetNumber(), pr.GetHTMLURL())
		}

	// === HERRAMIENTAS DASHBOARD ===
	case "github_dashboard":
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			err = fmt.Errorf("GITHUB_TOKEN environment variable not set")
		} else {
			dashClient := dashboard.NewDashboardClient(token)
			summary, dashErr := dashClient.GetFullDashboard(ctx, true)
			if dashErr != nil {
				err = dashErr
			} else {
				text = dashboard.FormatDashboardSummary(summary, true)
			}
		}

	case "github_notifications":
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			err = fmt.Errorf("GITHUB_TOKEN environment variable not set")
		} else {
			dashClient := dashboard.NewDashboardClient(token)
			all, _ := arguments["all"].(bool)
			notifications, notifErr := dashClient.GetNotifications(ctx, all)
			if notifErr != nil {
				err = notifErr
			} else {
				if len(notifications) == 0 {
					text = "🔔 No tienes notificaciones pendientes"
				} else {
					var lines []string
					lines = append(lines, fmt.Sprintf("🔔 **%d Notificaciones:**\n", len(notifications)))
					for _, n := range notifications {
						status := "🔵"
						if n.Unread {
							status = "🔴"
						}
						lines = append(lines, fmt.Sprintf("%s [%s] %s - %s", status, n.Reason, n.Subject.Title, n.Repository.FullName))
					}
					text = strings.Join(lines, "\n")
				}
			}
		}

	case "github_assigned_issues":
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			err = fmt.Errorf("GITHUB_TOKEN environment variable not set")
		} else {
			dashClient := dashboard.NewDashboardClient(token)
			issues, issuesErr := dashClient.GetAssignedIssues(ctx)
			if issuesErr != nil {
				err = issuesErr
			} else {
				if len(issues) == 0 {
					text = "📋 No tienes issues asignadas"
				} else {
					var lines []string
					lines = append(lines, fmt.Sprintf("📋 **%d Issues Asignadas:**\n", len(issues)))
					for _, issue := range issues {
						var labels []string
						for _, l := range issue.Labels {
							labels = append(labels, l.Name)
						}
						labelStr := ""
						if len(labels) > 0 {
							labelStr = fmt.Sprintf(" [%s]", strings.Join(labels, ", "))
						}
						lines = append(lines, fmt.Sprintf("• #%d: %s%s", issue.Number, issue.Title, labelStr))
					}
					text = strings.Join(lines, "\n")
				}
			}
		}

	case "github_prs_to_review":
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			err = fmt.Errorf("GITHUB_TOKEN environment variable not set")
		} else {
			dashClient := dashboard.NewDashboardClient(token)
			prs, prsErr := dashClient.GetPRsToReview(ctx)
			if prsErr != nil {
				err = prsErr
			} else {
				if len(prs) == 0 {
					text = "👀 No tienes PRs pendientes de revisión"
				} else {
					var lines []string
					lines = append(lines, fmt.Sprintf("👀 **%d PRs Pendientes de Revisión:**\n", len(prs)))
					for _, pr := range prs {
						lines = append(lines, fmt.Sprintf("• #%d: %s - %s", pr.Number, pr.Title, pr.HTMLURL))
					}
					text = strings.Join(lines, "\n")
				}
			}
		}

	case "github_security_alerts":
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			err = fmt.Errorf("GITHUB_TOKEN environment variable not set")
		} else {
			dashClient := dashboard.NewDashboardClient(token)
			owner, _ := arguments["owner"].(string)
			repo, _ := arguments["repo"].(string)
			alertType, _ := arguments["type"].(string)
			if alertType == "" {
				alertType = "all"
			}

			var lines []string
			lines = append(lines, "🛡️ **Alertas de Seguridad:**\n")

			if alertType == "all" || alertType == "dependabot" {
				depAlerts, _ := dashClient.GetDependabotAlerts(ctx, owner, repo)
				if len(depAlerts) > 0 {
					lines = append(lines, fmt.Sprintf("**Dependabot (%d):**", len(depAlerts)))
					for _, a := range depAlerts {
						lines = append(lines, fmt.Sprintf("  • [%s] %s - %s", a.SecurityAdvisory.Severity, a.SecurityAdvisory.Summary, a.Dependency.Package.Name))
					}
				}
			}

			if alertType == "all" || alertType == "secret" {
				secretAlerts, _ := dashClient.GetSecretScanningAlerts(ctx, owner, repo)
				if len(secretAlerts) > 0 {
					lines = append(lines, fmt.Sprintf("\n**Secret Scanning (%d):**", len(secretAlerts)))
					for _, a := range secretAlerts {
						lines = append(lines, fmt.Sprintf("  • [%s] %s", a.State, a.SecretType))
					}
				}
			}

			if alertType == "all" || alertType == "code" {
				codeAlerts, _ := dashClient.GetCodeScanningAlerts(ctx, owner, repo)
				if len(codeAlerts) > 0 {
					lines = append(lines, fmt.Sprintf("\n**Code Scanning (%d):**", len(codeAlerts)))
					for _, a := range codeAlerts {
						lines = append(lines, fmt.Sprintf("  • [%s] %s - %s", a.Rule.Severity, a.Rule.Description, a.MostRecentInstance.Location.Path))
					}
				}
			}

			if len(lines) == 1 {
				text = "🛡️ No se encontraron alertas de seguridad"
			} else {
				text = strings.Join(lines, "\n")
			}
		}

	case "github_failed_workflows":
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			err = fmt.Errorf("GITHUB_TOKEN environment variable not set")
		} else {
			dashClient := dashboard.NewDashboardClient(token)
			owner, _ := arguments["owner"].(string)
			repo, _ := arguments["repo"].(string)
			workflows, wfErr := dashClient.GetFailedWorkflows(ctx, owner, repo)
			if wfErr != nil {
				err = wfErr
			} else {
				if len(workflows) == 0 {
					text = "✅ No hay workflows fallidos recientemente"
				} else {
					var lines []string
					lines = append(lines, fmt.Sprintf("❌ **%d Workflows Fallidos:**\n", len(workflows)))
					for _, wf := range workflows {
						lines = append(lines, fmt.Sprintf("• %s - Run #%d - %s", wf.Name, wf.RunNumber, wf.HTMLURL))
					}
					text = strings.Join(lines, "\n")
				}
			}
		}

	case "github_mark_notification_read":
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			err = fmt.Errorf("GITHUB_TOKEN environment variable not set")
		} else {
			dashClient := dashboard.NewDashboardClient(token)
			threadID, _ := arguments["thread_id"].(string)
			markErr := dashClient.MarkNotificationAsRead(ctx, threadID)
			if markErr != nil {
				err = markErr
			} else {
				text = fmt.Sprintf("✅ Notificación %s marcada como leída", threadID)
			}
		}

	// === RESPONSE TOOLS ===
	case "github_comment_issue":
		owner, _ := arguments["owner"].(string)
		repo, _ := arguments["repo"].(string)
		number := int(arguments["number"].(float64))
		body, _ := arguments["body"].(string)

		comment, commentErr := s.GithubClient.CreateIssueComment(ctx, owner, repo, number, body)
		if commentErr != nil {
			err = commentErr
		} else {
			text = fmt.Sprintf("✅ Comentario agregado a issue #%d\n🔗 %s", number, comment.GetHTMLURL())
		}

	case "github_comment_pr":
		owner, _ := arguments["owner"].(string)
		repo, _ := arguments["repo"].(string)
		number := int(arguments["number"].(float64))
		body, _ := arguments["body"].(string)

		comment, commentErr := s.GithubClient.CreatePRComment(ctx, owner, repo, number, body)
		if commentErr != nil {
			err = commentErr
		} else {
			text = fmt.Sprintf("✅ Comentario agregado a PR #%d\n🔗 %s", number, comment.GetHTMLURL())
		}

	case "github_review_pr":
		owner, _ := arguments["owner"].(string)
		repo, _ := arguments["repo"].(string)
		number := int(arguments["number"].(float64))
		event, _ := arguments["event"].(string)
		body, _ := arguments["body"].(string)

		review, reviewErr := s.GithubClient.CreatePRReview(ctx, owner, repo, number, event, body)
		if reviewErr != nil {
			err = reviewErr
		} else {
			var eventEmoji string
			switch event {
			case "APPROVE":
				eventEmoji = "✅ Aprobado"
			case "REQUEST_CHANGES":
				eventEmoji = "🔴 Cambios solicitados"
			default:
				eventEmoji = "💬 Comentario"
			}
			text = fmt.Sprintf("%s PR #%d\n🔗 %s", eventEmoji, number, review.GetHTMLURL())
		}

	// === REPAIR TOOLS ===
	case "github_close_issue":
		owner, _ := arguments["owner"].(string)
		repo, _ := arguments["repo"].(string)
		number := int(arguments["number"].(float64))
		comment, _ := arguments["comment"].(string)

		issue, closeErr := s.GithubClient.CloseIssue(ctx, owner, repo, number, comment)
		if closeErr != nil {
			err = closeErr
		} else {
			text = fmt.Sprintf("🔒 Issue #%d cerrado\n🔗 %s", number, issue.GetHTMLURL())
		}

	case "github_merge_pr":
		owner, _ := arguments["owner"].(string)
		repo, _ := arguments["repo"].(string)
		number := int(arguments["number"].(float64))
		commitMessage, _ := arguments["commit_message"].(string)
		mergeMethod, _ := arguments["merge_method"].(string)
		if mergeMethod == "" {
			mergeMethod = "merge"
		}

		result, mergeErr := s.GithubClient.MergePullRequest(ctx, owner, repo, number, commitMessage, mergeMethod)
		if mergeErr != nil {
			err = mergeErr
		} else {
			text = fmt.Sprintf("🔀 PR #%d mergeado exitosamente\n✅ Mergeado: %v\n📝 SHA: %s",
				number, result.GetMerged(), result.GetSHA())
		}

	case "github_rerun_workflow":
		owner, _ := arguments["owner"].(string)
		repo, _ := arguments["repo"].(string)
		runID := int64(arguments["run_id"].(float64))
		failedOnly, _ := arguments["failed_jobs_only"].(bool)

		if failedOnly {
			err = s.GithubClient.RerunFailedJobs(ctx, owner, repo, runID)
			text = fmt.Sprintf("🔄 Re-ejecutando jobs fallidos para el workflow run %d", runID)
		} else {
			err = s.GithubClient.RerunWorkflow(ctx, owner, repo, runID)
			text = fmt.Sprintf("🔄 Re-ejecutando workflow run completo %d", runID)
		}

	case "github_dismiss_dependabot_alert":
		owner, _ := arguments["owner"].(string)
		repo, _ := arguments["repo"].(string)
		number := int(arguments["number"].(float64))
		reason, _ := arguments["reason"].(string)
		comment, _ := arguments["comment"].(string)

		alert, dismissErr := s.GithubClient.DismissDependabotAlert(ctx, owner, repo, number, reason, comment)
		if dismissErr != nil {
			err = dismissErr
		} else {
			text = fmt.Sprintf("🛡️ Alerta Dependabot #%d dismissada (razón: %s)\n🔗 %s",
				number, reason, alert.GetHTMLURL())
		}

	case "github_dismiss_code_alert":
		owner, _ := arguments["owner"].(string)
		repo, _ := arguments["repo"].(string)
		number := int64(arguments["number"].(float64))
		reason, _ := arguments["reason"].(string)
		comment, _ := arguments["comment"].(string)

		alert, dismissErr := s.GithubClient.DismissCodeScanningAlert(ctx, owner, repo, number, reason, comment)
		if dismissErr != nil {
			err = dismissErr
		} else {
			text = fmt.Sprintf("🔍 Alerta de code scanning #%d dismissada (razón: %s)\n🔗 %s",
				number, reason, alert.GetHTMLURL())
		}

	case "github_dismiss_secret_alert":
		owner, _ := arguments["owner"].(string)
		repo, _ := arguments["repo"].(string)
		number := int64(arguments["number"].(float64))
		resolution, _ := arguments["resolution"].(string)

		alert, dismissErr := s.GithubClient.DismissSecretScanningAlert(ctx, owner, repo, number, resolution)
		if dismissErr != nil {
			err = dismissErr
		} else {
			text = fmt.Sprintf("🔑 Alerta de secret scanning #%d resuelta (%s)\n🔗 %s",
				number, resolution, alert.GetHTMLURL())
		}

	default:
		// Check if it's an administrative tool (v3.0)
		if IsAdminOperation(name) {
			return HandleAdminTool(s, name, arguments)
		}
		// Check if it's a file operation tool (v3.0 - no Git required)
		if IsFileOperation(name) {
			return HandleFileTool(s, name, arguments)
		}
		return types.ToolCallResult{
			Content: []types.Content{{Type: "text", Text: "tool not found"}},
			IsError: true,
		}, nil
	}

	if err != nil {
		return types.ToolCallResult{
			Content: []types.Content{{Type: "text", Text: err.Error()}},
			IsError: true,
		}, nil
	}

	return types.ToolCallResult{
		Content: []types.Content{{Type: "text", Text: text}},
	}, nil
}
