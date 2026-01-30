package server

import (
	"context"
	"fmt"

	"github.com/google/go-github/v81/github"
	"github.com/jotajotape/github-go-server-mcp/pkg/types"
)

// HandleAdminTool routes administrative tool calls through safety middleware
func HandleAdminTool(s *MCPServer, name string, arguments map[string]interface{}) (types.ToolCallResult, error) {
	ctx := context.Background()

	// Check if safety middleware is enabled
	if s.Safety == nil {
		return types.ToolCallResult{}, fmt.Errorf("safety middleware not initialized")
	}

	// Route to specific handler
	switch name {
	// Repository Settings
	case "github_get_repo_settings":
		return handleGetRepoSettings(s, ctx, arguments)
	case "github_update_repo_settings":
		return handleUpdateRepoSettings(s, ctx, arguments)
	case "github_archive_repository":
		return handleArchiveRepository(s, ctx, arguments)
	case "github_delete_repository":
		return handleDeleteRepository(s, ctx, arguments)

	// Branch Protection
	case "github_get_branch_protection":
		return handleGetBranchProtection(s, ctx, arguments)
	case "github_update_branch_protection":
		return handleUpdateBranchProtection(s, ctx, arguments)
	case "github_delete_branch_protection":
		return handleDeleteBranchProtection(s, ctx, arguments)

	// Webhooks
	case "github_list_webhooks":
		return handleListWebhooks(s, ctx, arguments)
	case "github_create_webhook":
		return handleCreateWebhook(s, ctx, arguments)
	case "github_update_webhook":
		return handleUpdateWebhook(s, ctx, arguments)
	case "github_delete_webhook":
		return handleDeleteWebhook(s, ctx, arguments)
	case "github_test_webhook":
		return handleTestWebhook(s, ctx, arguments)

	// Collaborators
	case "github_list_collaborators":
		return handleListCollaborators(s, ctx, arguments)
	case "github_add_collaborator":
		return handleAddCollaborator(s, ctx, arguments)
	case "github_update_collaborator_permission":
		return handleUpdateCollaboratorPermission(s, ctx, arguments)
	case "github_remove_collaborator":
		return handleRemoveCollaborator(s, ctx, arguments)
	case "github_check_collaborator":
		return handleCheckCollaborator(s, ctx, arguments)

	// Invitations
	case "github_list_invitations":
		return handleListInvitations(s, ctx, arguments)
	case "github_accept_invitation":
		return handleAcceptInvitation(s, ctx, arguments)
	case "github_cancel_invitation":
		return handleCancelInvitation(s, ctx, arguments)

	// Team Access
	case "github_list_repo_teams":
		return handleListRepoTeams(s, ctx, arguments)
	case "github_add_repo_team":
		return handleAddRepoTeam(s, ctx, arguments)

	default:
		return types.ToolCallResult{}, fmt.Errorf("unknown administrative tool: %s", name)
	}
}

// ============================================================================
// Repository Settings Handlers
// ============================================================================

func handleGetRepoSettings(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)

	// LOW risk - execute directly without safety checks
	repository, err := s.AdminClient.GetRepositorySettings(ctx, owner, repo)
	if err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to get repository settings: %w", err)
	}

	text := fmt.Sprintf("📋 Repository Settings: %s/%s\n\n", owner, repo)
	text += fmt.Sprintf("Name: %s\n", *repository.Name)
	if repository.Description != nil {
		text += fmt.Sprintf("Description: %s\n", *repository.Description)
	}
	text += fmt.Sprintf("Private: %v\n", *repository.Private)
	text += fmt.Sprintf("Default Branch: %s\n", *repository.DefaultBranch)
	text += fmt.Sprintf("Has Issues: %v\n", *repository.HasIssues)
	text += fmt.Sprintf("Has Wiki: %v\n", *repository.HasWiki)
	text += fmt.Sprintf("Has Projects: %v\n", *repository.HasProjects)
	text += fmt.Sprintf("Allow Squash Merge: %v\n", *repository.AllowSquashMerge)
	text += fmt.Sprintf("Allow Merge Commit: %v\n", *repository.AllowMergeCommit)
	text += fmt.Sprintf("Allow Rebase Merge: %v\n", *repository.AllowRebaseMerge)
	if repository.DeleteBranchOnMerge != nil {
		text += fmt.Sprintf("Delete Branch on Merge: %v\n", *repository.DeleteBranchOnMerge)
	}

	return types.ToolCallResult{
		Content: []types.Content{{Type: "text", Text: text}},
	}, nil
}

func handleUpdateRepoSettings(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)

	// MEDIUM risk - use safety middleware
	return s.Safety.WrapExecution(ctx, "github_update_repo_settings", args, func() (string, error) {
		// Build settings map from arguments
		settings := make(map[string]interface{})
		for key, value := range args {
			if key != "owner" && key != "repo" && key != "dry_run" && key != "confirmation_token" {
				settings[key] = value
			}
		}

		repository, err := s.AdminClient.UpdateRepositorySettings(ctx, owner, repo, settings)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("✅ Updated repository settings for %s/%s\nNew name: %s", owner, repo, *repository.Name), nil
	})
}

func handleArchiveRepository(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)

	// CRITICAL risk - requires confirmation
	return s.Safety.WrapExecution(ctx, "github_archive_repository", args, func() (string, error) {
		repository, err := s.AdminClient.ArchiveRepository(ctx, owner, repo)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("📦 Archived repository %s/%s\nThe repository is now read-only.\nName: %s", owner, repo, *repository.Name), nil
	})
}

func handleDeleteRepository(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)

	// CRITICAL risk - requires confirmation + backup
	return s.Safety.WrapExecution(ctx, "github_delete_repository", args, func() (string, error) {
		err := s.AdminClient.DeleteRepository(ctx, owner, repo)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("💣 DELETED repository %s/%s\n⚠️  This operation is PERMANENT. All issues, PRs, and history are gone.", owner, repo), nil
	})
}

// ============================================================================
// Branch Protection Handlers
// ============================================================================

func handleGetBranchProtection(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	branch, _ := args["branch"].(string)

	// LOW risk - execute directly
	protection, err := s.AdminClient.GetBranchProtection(ctx, owner, repo, branch)
	if err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to get branch protection: %w", err)
	}

	text := fmt.Sprintf("🛡️ Branch Protection: %s/%s @ %s\n\n", owner, repo, branch)
	if protection.RequiredPullRequestReviews != nil {
		text += fmt.Sprintf("Required Approving Reviews: %d\n", protection.RequiredPullRequestReviews.RequiredApprovingReviewCount)
		text += fmt.Sprintf("Dismiss Stale Reviews: %v\n", protection.RequiredPullRequestReviews.DismissStaleReviews)
	}
	if protection.EnforceAdmins != nil {
		text += fmt.Sprintf("Enforce Admins: %v\n", protection.EnforceAdmins.Enabled)
	}

	return types.ToolCallResult{
		Content: []types.Content{{Type: "text", Text: text}},
	}, nil
}

func handleUpdateBranchProtection(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	branch, _ := args["branch"].(string)

	// HIGH risk - requires confirmation
	return s.Safety.WrapExecution(ctx, "github_update_branch_protection", args, func() (string, error) {
		// Build protection request
		protectionReq := &github.ProtectionRequest{}

		// Required pull request reviews
		if requireReviews, ok := args["require_pull_request_reviews"].(bool); ok && requireReviews {
			reviewCount := 1
			if count, ok := args["required_approving_review_count"].(float64); ok {
				reviewCount = int(count)
			} else if count, ok := args["required_approving_review_count"].(int); ok {
				reviewCount = count
			}

			dismissStale := false
			if dismiss, ok := args["dismiss_stale_reviews"].(bool); ok {
				dismissStale = dismiss
			}

			protectionReq.RequiredPullRequestReviews = &github.PullRequestReviewsEnforcementRequest{
				RequiredApprovingReviewCount: reviewCount,
				DismissStaleReviews:          dismissStale,
			}
		}

		// Required status checks
		if requireChecks, ok := args["require_status_checks"].(bool); ok && requireChecks {
			strict := false
			if s, ok := args["strict_status_checks"].(bool); ok {
				strict = s
			}

			contexts := []string{}
			if ctxs, ok := args["required_status_checks"].([]interface{}); ok {
				for _, ctx := range ctxs {
					if str, ok := ctx.(string); ok {
						contexts = append(contexts, str)
					}
				}
			}

			protectionReq.RequiredStatusChecks = &github.RequiredStatusChecks{
				Strict:   strict,
				Contexts: &contexts,
			}
		}

		// Enforce admins
		if enforce, ok := args["enforce_admins"].(bool); ok {
			protectionReq.EnforceAdmins = enforce
		}

		// Restrictions (who can push)
		if restrict, ok := args["restrictions"].(map[string]interface{}); ok {
			users := []string{}
			teams := []string{}

			if u, ok := restrict["users"].([]interface{}); ok {
				for _, user := range u {
					if str, ok := user.(string); ok {
						users = append(users, str)
					}
				}
			}

			if t, ok := restrict["teams"].([]interface{}); ok {
				for _, team := range t {
					if str, ok := team.(string); ok {
						teams = append(teams, str)
					}
				}
			}

			protectionReq.Restrictions = &github.BranchRestrictionsRequest{
				Users: users,
				Teams: teams,
			}
		}

		// Required linear history
		if linear, ok := args["required_linear_history"].(bool); ok {
			protectionReq.RequireLinearHistory = &linear
		}

		// Allow force pushes
		if allowForce, ok := args["allow_force_pushes"].(bool); ok {
			protectionReq.AllowForcePushes = &allowForce
		}

		// Allow deletions
		if allowDeletions, ok := args["allow_deletions"].(bool); ok {
			protectionReq.AllowDeletions = &allowDeletions
		}

		protection, err := s.AdminClient.UpdateBranchProtection(ctx, owner, repo, branch, protectionReq)
		if err != nil {
			return "", err
		}

		result := fmt.Sprintf("✅ Updated branch protection for %s/%s @ %s\n\n", owner, repo, branch)
		if protection.RequiredPullRequestReviews != nil {
			result += fmt.Sprintf("Required Reviews: %d\n", protection.RequiredPullRequestReviews.RequiredApprovingReviewCount)
		}
		if protection.EnforceAdmins != nil {
			result += fmt.Sprintf("Enforce Admins: %v\n", protection.EnforceAdmins.Enabled)
		}

		return result, nil
	})
}

func handleDeleteBranchProtection(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	branch, _ := args["branch"].(string)

	// CRITICAL risk - requires confirmation
	return s.Safety.WrapExecution(ctx, "github_delete_branch_protection", args, func() (string, error) {
		err := s.AdminClient.DeleteBranchProtection(ctx, owner, repo, branch)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("⚠️  Removed branch protection from %s/%s @ %s\nThe branch is now unprotected!", owner, repo, branch), nil
	})
}

// ============================================================================
// Webhook Handlers (Placeholders - Full implementation in next iteration)
// ============================================================================

func handleListWebhooks(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)

	hooks, err := s.AdminClient.ListWebhooks(ctx, owner, repo)
	if err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to list webhooks: %w", err)
	}

	text := fmt.Sprintf("📡 Webhooks for %s/%s (%d total)\n\n", owner, repo, len(hooks))
	for _, hook := range hooks {
		text += fmt.Sprintf("ID: %d | Active: %v | Events: %v\n", *hook.ID, *hook.Active, hook.Events)
		if hook.Config != nil && hook.Config.URL != nil {
			text += fmt.Sprintf("  URL: %s\n", *hook.Config.URL)
		}
	}

	return types.ToolCallResult{
		Content: []types.Content{{Type: "text", Text: text}},
	}, nil
}

func handleCreateWebhook(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)

	// MEDIUM risk - use safety middleware
	return s.Safety.WrapExecution(ctx, "github_create_webhook", args, func() (string, error) {
		// Build config map from arguments
		config := make(map[string]interface{})
		if url, ok := args["url"]; ok {
			config["url"] = url
		}
		if contentType, ok := args["content_type"]; ok {
			config["content_type"] = contentType
		}
		if secret, ok := args["secret"]; ok {
			config["secret"] = secret
		}
		if insecureSSL, ok := args["insecure_ssl"]; ok {
			config["insecure_ssl"] = insecureSSL
		}
		if events, ok := args["events"]; ok {
			config["events"] = events
		}
		if active, ok := args["active"]; ok {
			config["active"] = active
		}

		hook, err := s.AdminClient.CreateWebhook(ctx, owner, repo, config)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("✅ Created webhook for %s/%s\nID: %d\nActive: %v", owner, repo, *hook.ID, *hook.Active), nil
	})
}

func handleUpdateWebhook(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)

	// Extract hook_id (may come as float64 from JSON)
	var hookID int64
	switch v := args["hook_id"].(type) {
	case float64:
		hookID = int64(v)
	case int64:
		hookID = v
	case int:
		hookID = int64(v)
	default:
		return types.ToolCallResult{}, fmt.Errorf("invalid hook_id type")
	}

	// MEDIUM risk - use safety middleware
	return s.Safety.WrapExecution(ctx, "github_update_webhook", args, func() (string, error) {
		// Build config map from arguments
		config := make(map[string]interface{})
		if url, ok := args["url"]; ok {
			config["url"] = url
		}
		if contentType, ok := args["content_type"]; ok {
			config["content_type"] = contentType
		}
		if secret, ok := args["secret"]; ok {
			config["secret"] = secret
		}
		if insecureSSL, ok := args["insecure_ssl"]; ok {
			config["insecure_ssl"] = insecureSSL
		}
		if events, ok := args["events"]; ok {
			config["events"] = events
		}
		if active, ok := args["active"]; ok {
			config["active"] = active
		}

		hook, err := s.AdminClient.UpdateWebhook(ctx, owner, repo, hookID, config)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("✅ Updated webhook %d for %s/%s\nActive: %v", hookID, owner, repo, *hook.Active), nil
	})
}

func handleDeleteWebhook(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)

	// Extract hook_id (may come as float64 from JSON)
	var hookID int64
	switch v := args["hook_id"].(type) {
	case float64:
		hookID = int64(v)
	case int64:
		hookID = v
	case int:
		hookID = int64(v)
	default:
		return types.ToolCallResult{}, fmt.Errorf("invalid hook_id type")
	}

	// HIGH risk - requires confirmation
	return s.Safety.WrapExecution(ctx, "github_delete_webhook", args, func() (string, error) {
		err := s.AdminClient.DeleteWebhook(ctx, owner, repo, hookID)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("🗑️  Deleted webhook %d from %s/%s\n⚠️  Integrations using this webhook will stop working!", hookID, owner, repo), nil
	})
}

func handleTestWebhook(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)

	// Extract hook_id (may come as float64 from JSON)
	var hookID int64
	switch v := args["hook_id"].(type) {
	case float64:
		hookID = int64(v)
	case int64:
		hookID = v
	case int:
		hookID = int64(v)
	default:
		return types.ToolCallResult{}, fmt.Errorf("invalid hook_id type")
	}

	// LOW risk - execute directly
	err := s.AdminClient.TestWebhook(ctx, owner, repo, hookID)
	if err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to test webhook: %w", err)
	}

	text := fmt.Sprintf("✅ Test delivery sent for webhook %d in %s/%s\nCheck your endpoint to verify the delivery.", hookID, owner, repo)

	return types.ToolCallResult{
		Content: []types.Content{{Type: "text", Text: text}},
	}, nil
}

// ============================================================================
// Collaborator Handlers (Placeholders)
// ============================================================================

func handleListCollaborators(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)

	collaborators, err := s.AdminClient.ListCollaborators(ctx, owner, repo)
	if err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to list collaborators: %w", err)
	}

	text := fmt.Sprintf("👥 Collaborators for %s/%s (%d total)\n\n", owner, repo, len(collaborators))
	for _, collab := range collaborators {
		text += fmt.Sprintf("- @%s\n", *collab.Login)
	}

	return types.ToolCallResult{
		Content: []types.Content{{Type: "text", Text: text}},
	}, nil
}

func handleAddCollaborator(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	username, _ := args["username"].(string)
	permission, _ := args["permission"].(string)

	// Default to push if not specified
	if permission == "" {
		permission = "push"
	}

	// MEDIUM risk - use safety middleware
	return s.Safety.WrapExecution(ctx, "github_add_collaborator", args, func() (string, error) {
		invitation, err := s.AdminClient.AddCollaborator(ctx, owner, repo, username, permission)
		if err != nil {
			return "", err
		}

		var invitationID string
		if invitation != nil && invitation.ID != nil {
			invitationID = fmt.Sprintf("\nInvitation ID: %d", *invitation.ID)
		}

		return fmt.Sprintf("✅ Added @%s to %s/%s with '%s' permission%s\n\n🔄 Rollback:\ngithub_remove_collaborator --owner=%s --repo=%s --username=%s",
			username, owner, repo, permission, invitationID, owner, repo, username), nil
	})
}

func handleUpdateCollaboratorPermission(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	username, _ := args["username"].(string)
	permission, _ := args["permission"].(string)

	// MEDIUM risk - use safety middleware
	return s.Safety.WrapExecution(ctx, "github_update_collaborator_permission", args, func() (string, error) {
		invitation, err := s.AdminClient.UpdateCollaboratorPermission(ctx, owner, repo, username, permission)
		if err != nil {
			return "", err
		}

		var invitationID string
		if invitation != nil && invitation.ID != nil {
			invitationID = fmt.Sprintf("\nInvitation ID: %d", *invitation.ID)
		}

		return fmt.Sprintf("✅ Updated @%s permission to '%s' on %s/%s%s", username, permission, owner, repo, invitationID), nil
	})
}

func handleRemoveCollaborator(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	username, _ := args["username"].(string)

	// HIGH risk - requires confirmation
	return s.Safety.WrapExecution(ctx, "github_remove_collaborator", args, func() (string, error) {
		err := s.AdminClient.RemoveCollaborator(ctx, owner, repo, username)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("⚠️  Removed @%s from %s/%s\n🔒 User lost all access to this repository!", username, owner, repo), nil
	})
}

func handleCheckCollaborator(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	username, _ := args["username"].(string)

	isCollab, err := s.AdminClient.CheckCollaborator(ctx, owner, repo, username)
	if err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to check collaborator: %w", err)
	}

	text := fmt.Sprintf("✅ @%s is a collaborator on %s/%s: %v", username, owner, repo, isCollab)

	return types.ToolCallResult{
		Content: []types.Content{{Type: "text", Text: text}},
	}, nil
}

func handleListInvitations(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)

	// LOW risk - execute directly
	invitations, err := s.AdminClient.ListInvitations(ctx, owner, repo)
	if err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to list invitations: %w", err)
	}

	text := fmt.Sprintf("📨 Repository Invitations for %s/%s (%d total)\n\n", owner, repo, len(invitations))
	if len(invitations) == 0 {
		text += "No pending invitations."
	} else {
		for _, inv := range invitations {
			text += fmt.Sprintf("ID: %d | Invitee: @%s | Permission: %s\n", *inv.ID, *inv.Invitee.Login, *inv.Permissions)
		}
	}

	return types.ToolCallResult{
		Content: []types.Content{{Type: "text", Text: text}},
	}, nil
}

func handleAcceptInvitation(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	// Extract invitation_id (may come as float64 from JSON)
	var invitationID int64
	switch v := args["invitation_id"].(type) {
	case float64:
		invitationID = int64(v)
	case int64:
		invitationID = v
	case int:
		invitationID = int64(v)
	default:
		return types.ToolCallResult{}, fmt.Errorf("invalid invitation_id type")
	}

	// MEDIUM risk - use safety middleware
	return s.Safety.WrapExecution(ctx, "github_accept_invitation", args, func() (string, error) {
		err := s.AdminClient.AcceptInvitation(ctx, invitationID)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("✅ Accepted repository invitation ID %d\nYou now have access to the repository!", invitationID), nil
	})
}

func handleCancelInvitation(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)

	// Extract invitation_id (may come as float64 from JSON)
	var invitationID int64
	switch v := args["invitation_id"].(type) {
	case float64:
		invitationID = int64(v)
	case int64:
		invitationID = v
	case int:
		invitationID = int64(v)
	default:
		return types.ToolCallResult{}, fmt.Errorf("invalid invitation_id type")
	}

	// MEDIUM risk - use safety middleware
	return s.Safety.WrapExecution(ctx, "github_cancel_invitation", args, func() (string, error) {
		err := s.AdminClient.CancelInvitation(ctx, owner, repo, invitationID)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("✅ Cancelled repository invitation ID %d for %s/%s", invitationID, owner, repo), nil
	})
}

func handleListRepoTeams(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)

	// LOW risk - execute directly
	teams, err := s.AdminClient.ListRepositoryTeams(ctx, owner, repo)
	if err != nil {
		return types.ToolCallResult{}, fmt.Errorf("failed to list teams: %w", err)
	}

	text := fmt.Sprintf("👥 Teams with access to %s/%s (%d total)\n\n", owner, repo, len(teams))
	if len(teams) == 0 {
		text += "No teams have access to this repository."
	} else {
		for _, team := range teams {
			text += fmt.Sprintf("- %s (ID: %d)\n", *team.Name, *team.ID)
			if team.Permission != nil {
				text += fmt.Sprintf("  Permission: %s\n", *team.Permission)
			}
		}
	}

	return types.ToolCallResult{
		Content: []types.Content{{Type: "text", Text: text}},
	}, nil
}

func handleAddRepoTeam(s *MCPServer, ctx context.Context, args map[string]interface{}) (types.ToolCallResult, error) {
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	permission, _ := args["permission"].(string)

	// Extract team_id (may come as float64 from JSON)
	var teamID int64
	switch v := args["team_id"].(type) {
	case float64:
		teamID = int64(v)
	case int64:
		teamID = v
	case int:
		teamID = int64(v)
	default:
		return types.ToolCallResult{}, fmt.Errorf("invalid team_id type")
	}

	// Default to push if not specified
	if permission == "" {
		permission = "push"
	}

	// MEDIUM risk - use safety middleware
	return s.Safety.WrapExecution(ctx, "github_add_repo_team", args, func() (string, error) {
		err := s.AdminClient.AddRepositoryTeam(ctx, owner, repo, teamID, permission)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("✅ Added team %d to %s/%s with '%s' permission", teamID, owner, repo, permission), nil
	})
}
