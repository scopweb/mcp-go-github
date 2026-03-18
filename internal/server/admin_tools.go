package server

import "github.com/scopweb/mcp-go-github/pkg/types"

// ListAdminTools returns the list of administrative tools (v3.0)
// Consolidated into 4 tools using the operation parameter pattern.
func ListAdminTools() []types.Tool {
	return []types.Tool{
		// ========================================================================
		// Repository Administration (consolidated from 4 tools)
		// ========================================================================
		{
			Name:  "github_admin_repo",
			Title: "Repository Administration",
			Description: "Manage repository settings, archive, or delete. Operations: " +
				"get_settings (view repo configuration), " +
				"update_settings (modify name, description, visibility, features, default_branch, merge options), " +
				"archive (make repo read-only - CRITICAL), " +
				"delete (permanently delete repo - CRITICAL, requires confirmation_token).",
			Annotations: DestructiveAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation":              {Type: "string", Description: "Operation to perform: get_settings, update_settings, archive, delete"},
					"owner":                  {Type: "string", Description: "Repository owner (username or organization)"},
					"repo":                   {Type: "string", Description: "Repository name"},
					"name":                   {Type: "string", Description: "New repository name (update_settings only, optional)"},
					"description":            {Type: "string", Description: "Repository description (update_settings only, optional)"},
					"homepage":               {Type: "string", Description: "Repository homepage URL (update_settings only, optional)"},
					"private":                {Type: "boolean", Description: "Set repository visibility (update_settings only, optional)"},
					"has_issues":             {Type: "boolean", Description: "Enable/disable issues (update_settings only, optional)"},
					"has_wiki":               {Type: "boolean", Description: "Enable/disable wiki (update_settings only, optional)"},
					"has_projects":           {Type: "boolean", Description: "Enable/disable projects (update_settings only, optional)"},
					"default_branch":         {Type: "string", Description: "Default branch name (update_settings only, optional)"},
					"allow_squash_merge":     {Type: "boolean", Description: "Allow squash merging (update_settings only, optional)"},
					"allow_merge_commit":     {Type: "boolean", Description: "Allow merge commits (update_settings only, optional)"},
					"allow_rebase_merge":     {Type: "boolean", Description: "Allow rebase merging (update_settings only, optional)"},
					"delete_branch_on_merge": {Type: "boolean", Description: "Auto-delete branches after merge (update_settings only, optional)"},
					"dry_run":               {Type: "boolean", Description: "Preview changes without applying (default: true)"},
					"confirmation_token":     {Type: "string", Description: "Confirmation token for archive/delete operations"},
				},
				Required: []string{"operation", "owner", "repo"},
			},
		},

		// ========================================================================
		// Branch Protection (consolidated from 3 tools)
		// ========================================================================
		{
			Name:  "github_branch_protection",
			Title: "Branch Protection Management",
			Description: "Manage branch protection rules. Operations: " +
				"get (view branch protection rules), " +
				"update (configure required reviews, status checks, admin enforcement), " +
				"delete (remove all branch protection - CRITICAL, requires confirmation_token).",
			Annotations: DestructiveAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation":                       {Type: "string", Description: "Operation to perform: get, update, delete"},
					"owner":                           {Type: "string", Description: "Repository owner"},
					"repo":                            {Type: "string", Description: "Repository name"},
					"branch":                          {Type: "string", Description: "Branch name (e.g., 'main')"},
					"required_approving_review_count": {Type: "number", Description: "Number of required approvals, 1-6 (update only)"},
					"dismiss_stale_reviews":           {Type: "boolean", Description: "Dismiss approvals when new commits are pushed (update only)"},
					"require_code_owner_reviews":      {Type: "boolean", Description: "Require review from code owners (update only)"},
					"enforce_admins":                  {Type: "boolean", Description: "Enforce rules for administrators (update only)"},
					"required_status_checks":          {Type: "array", Description: "List of required status check names (update only)"},
					"strict_status_checks":            {Type: "boolean", Description: "Require branches to be up to date (update only)"},
					"dry_run":                         {Type: "boolean", Description: "Preview changes (default: true)"},
					"confirmation_token":              {Type: "string", Description: "Confirmation token for delete/high-risk changes"},
				},
				Required: []string{"operation", "owner", "repo", "branch"},
			},
		},

		// ========================================================================
		// Webhooks (consolidated from 5 tools)
		// ========================================================================
		{
			Name:  "github_webhooks",
			Title: "Webhook Management",
			Description: "Manage repository webhooks. Operations: " +
				"list (list all webhooks), " +
				"create (create new webhook with URL, events, content_type), " +
				"update (modify existing webhook by hook_id), " +
				"delete (remove webhook by hook_id - HIGH RISK, requires confirmation_token), " +
				"test (send test delivery to webhook by hook_id).",
			Annotations: CombineAnnotations(ModifyingAnnotation(), OpenWorldAnnotation()),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation":          {Type: "string", Description: "Operation to perform: list, create, update, delete, test"},
					"owner":              {Type: "string", Description: "Repository owner"},
					"repo":               {Type: "string", Description: "Repository name"},
					"hook_id":            {Type: "number", Description: "Webhook ID (required for update, delete, test)"},
					"url":                {Type: "string", Description: "Webhook URL (required for create, optional for update)"},
					"content_type":       {Type: "string", Description: "Content type: json or form (default: json)"},
					"secret":             {Type: "string", Description: "Webhook secret (optional)"},
					"events":             {Type: "array", Description: "Events to trigger: push, pull_request, issues, etc. (default: [push])"},
					"active":             {Type: "boolean", Description: "Activate/deactivate webhook (default: true)"},
					"dry_run":            {Type: "boolean", Description: "Preview changes (default: true)"},
					"confirmation_token": {Type: "string", Description: "Confirmation token for delete operation"},
				},
				Required: []string{"operation", "owner", "repo"},
			},
		},

		// ========================================================================
		// Collaborators and Teams (consolidated from 10 tools)
		// ========================================================================
		{
			Name:  "github_collaborators",
			Title: "Collaborator and Team Management",
			Description: "Manage repository collaborators, invitations, and team access. Operations: " +
				"list (list all collaborators), " +
				"check (check if user is a collaborator, requires username), " +
				"add (invite user with permission: pull/triage/push/maintain/admin), " +
				"update_permission (change collaborator permission level), " +
				"remove (revoke user access - HIGH RISK, requires confirmation_token), " +
				"list_invitations (view pending invitations), " +
				"accept_invitation (accept invitation by invitation_id), " +
				"cancel_invitation (cancel pending invitation by invitation_id), " +
				"list_teams (list teams with repository access), " +
				"add_team (grant team access by team_id with permission level).",
			Annotations: ModifyingAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"operation":          {Type: "string", Description: "Operation to perform: list, check, add, update_permission, remove, list_invitations, accept_invitation, cancel_invitation, list_teams, add_team"},
					"owner":              {Type: "string", Description: "Repository owner"},
					"repo":               {Type: "string", Description: "Repository name"},
					"username":           {Type: "string", Description: "GitHub username (for check, add, update_permission, remove)"},
					"permission":         {Type: "string", Description: "Permission level: pull, triage, push, maintain, admin (for add, update_permission, add_team)"},
					"invitation_id":      {Type: "number", Description: "Invitation ID (for accept_invitation, cancel_invitation)"},
					"team_id":            {Type: "number", Description: "Team ID (for add_team)"},
					"dry_run":            {Type: "boolean", Description: "Preview changes (default: true)"},
					"confirmation_token": {Type: "string", Description: "Confirmation token for remove operation"},
				},
				Required: []string{"operation", "owner", "repo"},
			},
		},
	}
}
