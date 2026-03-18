package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/scopweb/mcp-go-github/internal/hybrid"
	"github.com/scopweb/mcp-go-github/pkg/dashboard"
	"github.com/scopweb/mcp-go-github/pkg/interfaces"
	"github.com/scopweb/mcp-go-github/pkg/types"
)

// MCPServer representa el servidor MCP principal
type MCPServer struct {
	GithubClient    interfaces.GitHubOperations
	GitClient       interfaces.GitOperations
	AdminClient     interfaces.AdminOperations // v3.0: Administrative operations
	Safety          *SafetyMiddleware          // v3.0: Safety filter middleware
	GitAvailable    bool                       // v3.0: Whether git binary is installed
	RawGitHubClient interface{}                // v3.0: Raw *github.Client for file operations
	Toolsets        []string                   // Active toolsets filter (nil = all)
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
		clientProtocolVersion := "2024-11-05"
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
				"name":    "github-mcp-server-v4",
				"version": "4.1.0",
			},
		}
	case "initialized":
		response.Result = map[string]interface{}{}
	case "tools/list":
		response.Result = ListTools(s.GitAvailable, s.Toolsets)
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

// hasToolset returns true if the given toolset is active (nil = all active)
func hasToolset(toolsets []string, name string) bool {
	if len(toolsets) == 0 {
		return true
	}
	for _, t := range toolsets {
		if strings.TrimSpace(t) == name {
			return true
		}
	}
	return false
}

// ListTools retorna la lista de herramientas disponibles
func ListTools(gitAvailable bool, toolsets []string) types.ToolsListResult {
	allTools := []types.Tool{}

	// Git tools (information, basic, advanced)
	if hasToolset(toolsets, "git") {
		allTools = append(allTools, ListGitInfoTools()...)
		allTools = append(allTools, ListGitBasicTools()...)
		allTools = append(allTools, ListGitAdvancedTools()...)
	}

	// Hybrid + file tools
	if hasToolset(toolsets, "files") {
		allTools = append(allTools, ListHybridTools()...)
		allTools = append(allTools, ListFileTools()...)
	}

	// GitHub API tools
	if hasToolset(toolsets, "github") {
		allTools = append(allTools, ListGitHubAPITools()...)
		allTools = append(allTools, ListDashboardTools()...)
		allTools = append(allTools, ListResponseTools()...)
		allTools = append(allTools, ListRepairTools()...)
	}

	// Administrative tools
	if hasToolset(toolsets, "admin") {
		allTools = append(allTools, ListAdminTools()...)
	}

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
			Content: []types.Content{{Type: "text", Text: fmt.Sprintf("Git is not installed on this system.\n\nThe tool '%s' requires a local Git binary.\n\nAlternatives:\n- Use GitHub API tools (github_*) which work without Git\n- Install Git: https://git-scm.com/downloads\n\nAvailable without Git: dashboard, repos, PRs, issues, webhooks, collaborators, branch protection, and all admin tools.", name)}},
		}, nil
	}

	switch name {
	// =================================================================
	// git_info (consolidated: status, file_sha, last_commit, file_content,
	//           changed_files, validate_repo, list_files, context, validate_clean)
	// =================================================================
	case "git_info":
		operation, _ := arguments["operation"].(string)
		switch operation {
		case "status":
			text, err = s.GitClient.Status()
		case "file_sha":
			path, _ := arguments["path"].(string)
			text, err = s.GitClient.GetFileSHA(path)
		case "last_commit":
			text, err = s.GitClient.GetLastCommit()
		case "file_content":
			path, _ := arguments["path"].(string)
			ref, _ := arguments["ref"].(string)
			text, err = s.GitClient.GetFileContent(path, ref)
		case "changed_files":
			staged, _ := arguments["staged"].(bool)
			text, err = s.GitClient.GetChangedFiles(staged)
		case "validate_repo":
			path, _ := arguments["path"].(string)
			text, err = s.GitClient.ValidateRepo(path)
		case "list_files":
			ref, _ := arguments["ref"].(string)
			text, err = s.GitClient.ListFiles(ref)
		case "context":
			text = hybrid.AutoDetectContext(s.GitClient)
		case "validate_clean":
			clean, validateErr := s.GitClient.ValidateCleanState()
			if validateErr != nil {
				err = validateErr
			} else if clean {
				text = "Working directory is clean"
			} else {
				text = "Working directory has uncommitted changes"
			}
		default:
			return types.ToolCallResult{}, fmt.Errorf("unknown operation '%s' for git_info", operation)
		}

	case "git_set_workspace":
		path, _ := arguments["path"].(string)
		text, err = s.GitClient.SetWorkspace(path)

	// =================================================================
	// Individual Git tools (frequent workflow)
	// =================================================================
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

	// =================================================================
	// git_history (consolidated: log, diff)
	// =================================================================
	case "git_history":
		operation, _ := arguments["operation"].(string)
		switch operation {
		case "log":
			limit, _ := arguments["limit"].(string)
			text, err = s.GitClient.LogAnalysis(limit)
		case "diff":
			staged, _ := arguments["staged"].(bool)
			text, err = s.GitClient.DiffFiles(staged)
		default:
			return types.ToolCallResult{}, fmt.Errorf("unknown operation '%s' for git_history", operation)
		}

	// =================================================================
	// git_branch (consolidated: checkout, checkout_remote, list, merge, rebase, backup)
	// =================================================================
	case "git_branch":
		operation, _ := arguments["operation"].(string)
		switch operation {
		case "checkout":
			branch, _ := arguments["branch"].(string)
			create, _ := arguments["create"].(bool)
			text, err = s.GitClient.Checkout(branch, create)
		case "checkout_remote":
			remoteBranch, _ := arguments["remote_branch"].(string)
			localBranch, _ := arguments["local_branch"].(string)
			text, err = s.GitClient.CheckoutRemote(remoteBranch, localBranch)
		case "list":
			remote, _ := arguments["remote"].(bool)
			branches, branchErr := s.GitClient.BranchList(remote)
			if branchErr != nil {
				err = branchErr
			} else {
				jsonOutput, jsonErr := json.MarshalIndent(branches, "", "  ")
				if jsonErr != nil {
					err = fmt.Errorf("failed to marshal branch list: %w", jsonErr)
				} else {
					text = string(jsonOutput)
				}
			}
		case "merge":
			sourceBranch, _ := arguments["source_branch"].(string)
			targetBranch, _ := arguments["target_branch"].(string)
			text, err = s.GitClient.Merge(sourceBranch, targetBranch)
		case "rebase":
			branch, _ := arguments["branch"].(string)
			text, err = s.GitClient.Rebase(branch)
		case "backup":
			backupName, _ := arguments["name"].(string)
			text, err = s.GitClient.CreateBackup(backupName)
		default:
			return types.ToolCallResult{}, fmt.Errorf("unknown operation '%s' for git_branch", operation)
		}

	// =================================================================
	// git_sync (consolidated: push, pull, force_push, push_upstream, sync, pull_strategy)
	// =================================================================
	case "git_sync":
		operation, _ := arguments["operation"].(string)
		switch operation {
		case "push":
			branch, _ := arguments["branch"].(string)
			text, err = s.GitClient.Push(branch)
		case "pull":
			branch, _ := arguments["branch"].(string)
			text, err = s.GitClient.Pull(branch)
		case "force_push":
			branch, _ := arguments["branch"].(string)
			force, _ := arguments["force"].(bool)
			text, err = s.GitClient.ForcePush(branch, force)
		case "push_upstream":
			branch, _ := arguments["branch"].(string)
			text, err = s.GitClient.PushUpstream(branch)
		case "sync":
			remoteBranch, _ := arguments["remote_branch"].(string)
			text, err = s.GitClient.SyncWithRemote(remoteBranch)
		case "pull_strategy":
			branch, _ := arguments["branch"].(string)
			strategy, _ := arguments["strategy"].(string)
			text, err = s.GitClient.PullWithStrategy(branch, strategy)
		default:
			return types.ToolCallResult{}, fmt.Errorf("unknown operation '%s' for git_sync", operation)
		}

	// =================================================================
	// git_conflict (consolidated: status, resolve, detect, safe_merge)
	// =================================================================
	case "git_conflict":
		operation, _ := arguments["operation"].(string)
		switch operation {
		case "status":
			text, err = s.GitClient.ConflictStatus()
		case "resolve":
			strategy, _ := arguments["strategy"].(string)
			text, err = s.GitClient.ResolveConflicts(strategy)
		case "detect":
			sourceBranch, _ := arguments["source_branch"].(string)
			targetBranch, _ := arguments["target_branch"].(string)
			conflictInfo, detectErr := s.GitClient.DetectPotentialConflicts(sourceBranch, targetBranch)
			if detectErr != nil {
				err = detectErr
			} else if conflictInfo == "" {
				text = "No potential conflicts detected between branches"
			} else {
				text = conflictInfo
			}
		case "safe_merge":
			source, _ := arguments["source"].(string)
			target, _ := arguments["target"].(string)
			text, err = s.GitClient.SafeMerge(source, target)
		default:
			return types.ToolCallResult{}, fmt.Errorf("unknown operation '%s' for git_conflict", operation)
		}

	// =================================================================
	// Existing consolidated Git tools (stash, remote, tag, clean, reset)
	// =================================================================
	case "git_stash":
		operation, _ := arguments["operation"].(string)
		stashName, _ := arguments["name"].(string)
		text, err = s.GitClient.Stash(operation, stashName)
	case "git_remote":
		operation, _ := arguments["operation"].(string)
		remoteName, _ := arguments["name"].(string)
		url, _ := arguments["url"].(string)
		text, err = s.GitClient.Remote(operation, remoteName, url)
	case "git_tag":
		operation, _ := arguments["operation"].(string)
		tagName, _ := arguments["tag_name"].(string)
		message, _ := arguments["message"].(string)
		text, err = s.GitClient.Tag(operation, tagName, message)
	case "git_clean":
		operation, _ := arguments["operation"].(string)
		dryRun, exists := arguments["dry_run"].(bool)
		if !exists {
			dryRun = true
		}
		text, err = s.GitClient.Clean(operation, dryRun)
	case "git_reset":
		mode, _ := arguments["mode"].(string)
		target, _ := arguments["target"].(string)
		filesStr, _ := arguments["files"].(string)
		var files []string
		if filesStr != "" {
			for _, f := range strings.Split(filesStr, ",") {
				if f = strings.TrimSpace(f); f != "" {
					files = append(files, f)
				}
			}
		}
		text, err = s.GitClient.Reset(mode, target, files)

	// =================================================================
	// Hybrid tools (Git-first, API fallback)
	// =================================================================
	case "create_file":
		text, err = hybrid.SmartCreateFile(s.GitClient, s.GithubClient, arguments)
	case "update_file":
		text, err = hybrid.SmartUpdateFile(s.GitClient, s.GithubClient, arguments)
	case "push_files":
		text, err = hybrid.PushFiles(s.GitClient, arguments)

	// =================================================================
	// github_repo (consolidated: list_repos, create_repo, list_prs, create_pr)
	// =================================================================
	case "github_repo":
		operation, _ := arguments["operation"].(string)
		switch operation {
		case "list_repos":
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
		case "create_repo":
			repoName, _ := arguments["name"].(string)
			description, _ := arguments["description"].(string)
			private, _ := arguments["private"].(bool)
			repo, createErr := s.GithubClient.CreateRepository(ctx, repoName, description, private)
			if createErr != nil {
				err = createErr
			} else {
				text = fmt.Sprintf("Successfully created repository: %s", repo.GetFullName())
			}
		case "list_prs":
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
		case "create_pr":
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
		default:
			return types.ToolCallResult{}, fmt.Errorf("unknown operation '%s' for github_repo", operation)
		}

	// =================================================================
	// github_dashboard (consolidated: full, notifications, issues,
	//                    prs_review, security, workflows, mark_read)
	// =================================================================
	case "github_dashboard":
		operation, _ := arguments["operation"].(string)
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			err = fmt.Errorf("GITHUB_TOKEN environment variable not set")
		} else {
			dashClient := dashboard.NewDashboardClient(token)
			switch operation {
			case "full":
				summary, dashErr := dashClient.GetFullDashboard(ctx, true)
				if dashErr != nil {
					err = dashErr
				} else {
					text = dashboard.FormatDashboardSummary(summary, true)
				}
			case "notifications":
				all, _ := arguments["all"].(bool)
				notifications, notifErr := dashClient.GetNotifications(ctx, all)
				if notifErr != nil {
					err = notifErr
				} else if len(notifications) == 0 {
					text = "No pending notifications"
				} else {
					var lines []string
					lines = append(lines, fmt.Sprintf("%d Notifications:\n", len(notifications)))
					for _, n := range notifications {
						status := "read"
						if n.Unread {
							status = "unread"
						}
						lines = append(lines, fmt.Sprintf("[%s] [%s] %s - %s", status, n.Reason, n.Subject.Title, n.Repository.FullName))
					}
					text = strings.Join(lines, "\n")
				}
			case "issues":
				issues, issuesErr := dashClient.GetAssignedIssues(ctx)
				if issuesErr != nil {
					err = issuesErr
				} else if len(issues) == 0 {
					text = "No assigned issues"
				} else {
					var lines []string
					lines = append(lines, fmt.Sprintf("%d Assigned Issues:\n", len(issues)))
					for _, issue := range issues {
						var labels []string
						for _, l := range issue.Labels {
							labels = append(labels, l.Name)
						}
						labelStr := ""
						if len(labels) > 0 {
							labelStr = fmt.Sprintf(" [%s]", strings.Join(labels, ", "))
						}
						lines = append(lines, fmt.Sprintf("- #%d: %s%s", issue.Number, issue.Title, labelStr))
					}
					text = strings.Join(lines, "\n")
				}
			case "prs_review":
				prs, prsErr := dashClient.GetPRsToReview(ctx)
				if prsErr != nil {
					err = prsErr
				} else if len(prs) == 0 {
					text = "No PRs pending review"
				} else {
					var lines []string
					lines = append(lines, fmt.Sprintf("%d PRs Pending Review:\n", len(prs)))
					for _, pr := range prs {
						lines = append(lines, fmt.Sprintf("- #%d: %s - %s", pr.Number, pr.Title, pr.HTMLURL))
					}
					text = strings.Join(lines, "\n")
				}
			case "security":
				owner, _ := arguments["owner"].(string)
				repo, _ := arguments["repo"].(string)
				alertType, _ := arguments["type"].(string)
				if alertType == "" {
					alertType = "all"
				}

				var lines []string
				lines = append(lines, "Security Alerts:\n")

				if alertType == "all" || alertType == "dependabot" {
					depAlerts, _ := dashClient.GetDependabotAlerts(ctx, owner, repo)
					if len(depAlerts) > 0 {
						lines = append(lines, fmt.Sprintf("Dependabot (%d):", len(depAlerts)))
						for _, a := range depAlerts {
							lines = append(lines, fmt.Sprintf("  - [%s] %s - %s", a.SecurityAdvisory.Severity, a.SecurityAdvisory.Summary, a.Dependency.Package.Name))
						}
					}
				}
				if alertType == "all" || alertType == "secret" {
					secretAlerts, _ := dashClient.GetSecretScanningAlerts(ctx, owner, repo)
					if len(secretAlerts) > 0 {
						lines = append(lines, fmt.Sprintf("\nSecret Scanning (%d):", len(secretAlerts)))
						for _, a := range secretAlerts {
							lines = append(lines, fmt.Sprintf("  - [%s] %s", a.State, a.SecretType))
						}
					}
				}
				if alertType == "all" || alertType == "code" {
					codeAlerts, _ := dashClient.GetCodeScanningAlerts(ctx, owner, repo)
					if len(codeAlerts) > 0 {
						lines = append(lines, fmt.Sprintf("\nCode Scanning (%d):", len(codeAlerts)))
						for _, a := range codeAlerts {
							lines = append(lines, fmt.Sprintf("  - [%s] %s - %s", a.Rule.Severity, a.Rule.Description, a.MostRecentInstance.Location.Path))
						}
					}
				}

				if len(lines) == 1 {
					text = "No security alerts found"
				} else {
					text = strings.Join(lines, "\n")
				}
			case "workflows":
				owner, _ := arguments["owner"].(string)
				repo, _ := arguments["repo"].(string)
				workflows, wfErr := dashClient.GetFailedWorkflows(ctx, owner, repo)
				if wfErr != nil {
					err = wfErr
				} else if len(workflows) == 0 {
					text = "No failed workflows recently"
				} else {
					var lines []string
					lines = append(lines, fmt.Sprintf("%d Failed Workflows:\n", len(workflows)))
					for _, wf := range workflows {
						lines = append(lines, fmt.Sprintf("- %s - Run #%d - %s", wf.Name, wf.RunNumber, wf.HTMLURL))
					}
					text = strings.Join(lines, "\n")
				}
			case "mark_read":
				threadID, _ := arguments["thread_id"].(string)
				markErr := dashClient.MarkNotificationAsRead(ctx, threadID)
				if markErr != nil {
					err = markErr
				} else {
					text = fmt.Sprintf("Notification %s marked as read", threadID)
				}
			default:
				return types.ToolCallResult{}, fmt.Errorf("unknown operation '%s' for github_dashboard", operation)
			}
		}

	// =================================================================
	// github_respond (consolidated: comment_issue, comment_pr, review_pr)
	// =================================================================
	case "github_respond":
		operation, _ := arguments["operation"].(string)
		switch operation {
		case "comment_issue":
			owner, _ := arguments["owner"].(string)
			repo, _ := arguments["repo"].(string)
			number := int(arguments["number"].(float64))
			body, _ := arguments["body"].(string)
			comment, commentErr := s.GithubClient.CreateIssueComment(ctx, owner, repo, number, body)
			if commentErr != nil {
				err = commentErr
			} else {
				text = fmt.Sprintf("Comment added to issue #%d\n%s", number, comment.GetHTMLURL())
			}
		case "comment_pr":
			owner, _ := arguments["owner"].(string)
			repo, _ := arguments["repo"].(string)
			number := int(arguments["number"].(float64))
			body, _ := arguments["body"].(string)
			comment, commentErr := s.GithubClient.CreatePRComment(ctx, owner, repo, number, body)
			if commentErr != nil {
				err = commentErr
			} else {
				text = fmt.Sprintf("Comment added to PR #%d\n%s", number, comment.GetHTMLURL())
			}
		case "review_pr":
			owner, _ := arguments["owner"].(string)
			repo, _ := arguments["repo"].(string)
			number := int(arguments["number"].(float64))
			event, _ := arguments["event"].(string)
			body, _ := arguments["body"].(string)
			review, reviewErr := s.GithubClient.CreatePRReview(ctx, owner, repo, number, event, body)
			if reviewErr != nil {
				err = reviewErr
			} else {
				var eventLabel string
				switch event {
				case "APPROVE":
					eventLabel = "Approved"
				case "REQUEST_CHANGES":
					eventLabel = "Changes requested"
				default:
					eventLabel = "Comment"
				}
				text = fmt.Sprintf("%s PR #%d\n%s", eventLabel, number, review.GetHTMLURL())
			}
		default:
			return types.ToolCallResult{}, fmt.Errorf("unknown operation '%s' for github_respond", operation)
		}

	// =================================================================
	// github_repair (consolidated: close_issue, merge_pr, rerun_workflow, dismiss_alert)
	// =================================================================
	case "github_repair":
		operation, _ := arguments["operation"].(string)
		switch operation {
		case "close_issue":
			owner, _ := arguments["owner"].(string)
			repo, _ := arguments["repo"].(string)
			number := int(arguments["number"].(float64))
			comment, _ := arguments["comment"].(string)
			issue, closeErr := s.GithubClient.CloseIssue(ctx, owner, repo, number, comment)
			if closeErr != nil {
				err = closeErr
			} else {
				text = fmt.Sprintf("Issue #%d closed\n%s", number, issue.GetHTMLURL())
			}
		case "merge_pr":
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
				text = fmt.Sprintf("PR #%d merged successfully\nMerged: %v\nSHA: %s",
					number, result.GetMerged(), result.GetSHA())
			}
		case "rerun_workflow":
			owner, _ := arguments["owner"].(string)
			repo, _ := arguments["repo"].(string)
			runID := int64(arguments["run_id"].(float64))
			failedOnly, _ := arguments["failed_jobs_only"].(bool)
			if failedOnly {
				err = s.GithubClient.RerunFailedJobs(ctx, owner, repo, runID)
				text = fmt.Sprintf("Re-running failed jobs for workflow run %d", runID)
			} else {
				err = s.GithubClient.RerunWorkflow(ctx, owner, repo, runID)
				text = fmt.Sprintf("Re-running full workflow run %d", runID)
			}
		case "dismiss_alert":
			owner, _ := arguments["owner"].(string)
			repo, _ := arguments["repo"].(string)
			alertType, _ := arguments["alert_type"].(string)
			switch alertType {
			case "dependabot":
				number := int(arguments["number"].(float64))
				reason, _ := arguments["reason"].(string)
				comment, _ := arguments["comment"].(string)
				alert, dismissErr := s.GithubClient.DismissDependabotAlert(ctx, owner, repo, number, reason, comment)
				if dismissErr != nil {
					err = dismissErr
				} else {
					text = fmt.Sprintf("Dependabot alert #%d dismissed (reason: %s)\n%s", number, reason, alert.GetHTMLURL())
				}
			case "code":
				number := int64(arguments["number"].(float64))
				reason, _ := arguments["reason"].(string)
				comment, _ := arguments["comment"].(string)
				alert, dismissErr := s.GithubClient.DismissCodeScanningAlert(ctx, owner, repo, number, reason, comment)
				if dismissErr != nil {
					err = dismissErr
				} else {
					text = fmt.Sprintf("Code scanning alert #%d dismissed (reason: %s)\n%s", number, reason, alert.GetHTMLURL())
				}
			case "secret":
				number := int64(arguments["number"].(float64))
				resolution, _ := arguments["resolution"].(string)
				alert, dismissErr := s.GithubClient.DismissSecretScanningAlert(ctx, owner, repo, number, resolution)
				if dismissErr != nil {
					err = dismissErr
				} else {
					text = fmt.Sprintf("Secret scanning alert #%d resolved (%s)\n%s", number, resolution, alert.GetHTMLURL())
				}
			default:
				return types.ToolCallResult{}, fmt.Errorf("unknown alert_type '%s' for github_repair dismiss_alert (use: dependabot, code, secret)", alertType)
			}
		default:
			return types.ToolCallResult{}, fmt.Errorf("unknown operation '%s' for github_repair", operation)
		}

	// =================================================================
	// Administrative tools (v3.0)
	// =================================================================
	case "github_admin_repo", "github_branch_protection", "github_webhooks", "github_collaborators":
		return HandleAdminTool(s, name, arguments)

	// =================================================================
	// File operations (v3.0 - work without Git)
	// =================================================================
	case "github_files":
		return HandleFileTool(s, name, arguments)

	default:
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
