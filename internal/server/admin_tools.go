package server

import "github.com/jotajotape/github-go-server-mcp/pkg/types"

// ListAdminTools returns the list of administrative tools (v3.0)
func ListAdminTools() []types.Tool {
	return []types.Tool{
		// ========================================================================
		// Repository Settings (4 tools)
		// ========================================================================
		{
			Name:        "github_get_repo_settings",
			Title:       "Get Repository Settings",
			Description: "📋 View repository configuration and settings",
			Annotations: ReadOnlyAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner": {Type: "string", Description: "Repository owner (username or organization)"},
					"repo":  {Type: "string", Description: "Repository name"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_update_repo_settings",
			Title:       "Update Repository Settings",
			Description: "⚙️ Modify repository configuration (name, description, visibility, features)",
			Annotations: ModifyingAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":       {Type: "string", Description: "Repository owner"},
					"repo":        {Type: "string", Description: "Repository name"},
					"name":        {Type: "string", Description: "New repository name (optional)"},
					"description": {Type: "string", Description: "Repository description (optional)"},
					"homepage":    {Type: "string", Description: "Repository homepage URL (optional)"},
					"private":     {Type: "boolean", Description: "Set repository visibility (optional)"},
					"has_issues":  {Type: "boolean", Description: "Enable/disable issues (optional)"},
					"has_wiki":    {Type: "boolean", Description: "Enable/disable wiki (optional)"},
					"has_projects": {Type: "boolean", Description: "Enable/disable projects (optional)"},
					"default_branch": {Type: "string", Description: "Default branch name (optional)"},
					"allow_squash_merge": {Type: "boolean", Description: "Allow squash merging (optional)"},
					"allow_merge_commit": {Type: "boolean", Description: "Allow merge commits (optional)"},
					"allow_rebase_merge": {Type: "boolean", Description: "Allow rebase merging (optional)"},
					"delete_branch_on_merge": {Type: "boolean", Description: "Auto-delete branches after merge (optional)"},
					"dry_run": {Type: "boolean", Description: "Preview changes without applying (default: true)"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_archive_repository",
			Title:       "Archive Repository",
			Description: "📦 Archive repository (makes it read-only) - CRITICAL operation",
			Annotations: DestructiveAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
					"confirmation_token": {Type: "string", Description: "Confirmation token from previous call"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_delete_repository",
			Title:       "Delete Repository",
			Description: "💣 Delete repository PERMANENTLY - CRITICAL operation",
			Annotations: DestructiveAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
					"confirmation_token": {Type: "string", Description: "Confirmation token required"},
				},
				Required: []string{"owner", "repo"},
			},
		},

		// ========================================================================
		// Branch Protection (3 tools)
		// ========================================================================
		{
			Name:        "github_get_branch_protection",
			Title:       "Get Branch Protection",
			Description: "🛡️ View branch protection rules",
			Annotations: ReadOnlyAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":  {Type: "string", Description: "Repository owner"},
					"repo":   {Type: "string", Description: "Repository name"},
					"branch": {Type: "string", Description: "Branch name (e.g., 'main')"},
				},
				Required: []string{"owner", "repo", "branch"},
			},
		},
		{
			Name:        "github_update_branch_protection",
			Title:       "Update Branch Protection",
			Description: "🔒 Configure branch protection rules (requires reviews, status checks, etc.)",
			Annotations: ModifyingAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":  {Type: "string", Description: "Repository owner"},
					"repo":   {Type: "string", Description: "Repository name"},
					"branch": {Type: "string", Description: "Branch name"},
					"required_approving_review_count": {Type: "number", Description: "Number of required approvals (1-6)"},
					"dismiss_stale_reviews": {Type: "boolean", Description: "Dismiss approvals when new commits are pushed"},
					"require_code_owner_reviews": {Type: "boolean", Description: "Require review from code owners"},
					"enforce_admins": {Type: "boolean", Description: "Enforce rules for administrators"},
					"required_status_checks": {Type: "array", Description: "List of required status check names"},
					"strict_status_checks": {Type: "boolean", Description: "Require branches to be up to date"},
					"dry_run": {Type: "boolean", Description: "Preview changes (default: true)"},
					"confirmation_token": {Type: "string", Description: "Confirmation token for high-risk changes"},
				},
				Required: []string{"owner", "repo", "branch"},
			},
		},
		{
			Name:        "github_delete_branch_protection",
			Title:       "Delete Branch Protection",
			Description: "⚠️ Remove branch protection - CRITICAL operation",
			Annotations: DestructiveAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":  {Type: "string", Description: "Repository owner"},
					"repo":   {Type: "string", Description: "Repository name"},
					"branch": {Type: "string", Description: "Branch name"},
					"confirmation_token": {Type: "string", Description: "Confirmation token required"},
				},
				Required: []string{"owner", "repo", "branch"},
			},
		},

		// ========================================================================
		// Webhooks (5 tools)
		// ========================================================================
		{
			Name:        "github_list_webhooks",
			Title:       "List Webhooks",
			Description: "📡 List all repository webhooks",
			Annotations: ReadOnlyAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_create_webhook",
			Title:       "Create Webhook",
			Description: "➕ Create new repository webhook",
			Annotations: CombineAnnotations(ModifyingAnnotation(), OpenWorldAnnotation()),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":        {Type: "string", Description: "Repository owner"},
					"repo":         {Type: "string", Description: "Repository name"},
					"url":          {Type: "string", Description: "Webhook URL"},
					"content_type": {Type: "string", Description: "Content type: json or form (default: json)"},
					"secret":       {Type: "string", Description: "Webhook secret (optional)"},
					"events":       {Type: "array", Description: "Events to trigger: push, pull_request, issues, etc. (default: [push])"},
					"active":       {Type: "boolean", Description: "Activate webhook (default: true)"},
					"dry_run":      {Type: "boolean", Description: "Preview (default: true)"},
				},
				Required: []string{"owner", "repo", "url"},
			},
		},
		{
			Name:        "github_update_webhook",
			Title:       "Update Webhook",
			Description: "✏️ Modify existing webhook configuration",
			Annotations: CombineAnnotations(ModifyingAnnotation(), OpenWorldAnnotation()),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":        {Type: "string", Description: "Repository owner"},
					"repo":         {Type: "string", Description: "Repository name"},
					"hook_id":      {Type: "number", Description: "Webhook ID"},
					"url":          {Type: "string", Description: "New webhook URL (optional)"},
					"content_type": {Type: "string", Description: "Content type (optional)"},
					"secret":       {Type: "string", Description: "New secret (optional)"},
					"events":       {Type: "array", Description: "Events to trigger (optional)"},
					"active":       {Type: "boolean", Description: "Activate/deactivate (optional)"},
					"dry_run":      {Type: "boolean", Description: "Preview (default: true)"},
				},
				Required: []string{"owner", "repo", "hook_id"},
			},
		},
		{
			Name:        "github_delete_webhook",
			Title:       "Delete Webhook",
			Description: "🗑️ Delete webhook (breaks integrations) - HIGH RISK",
			Annotations: DestructiveAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":  {Type: "string", Description: "Repository owner"},
					"repo":   {Type: "string", Description: "Repository name"},
					"hook_id": {Type: "number", Description: "Webhook ID to delete"},
					"confirmation_token": {Type: "string", Description: "Confirmation token required"},
				},
				Required: []string{"owner", "repo", "hook_id"},
			},
		},
		{
			Name:        "github_test_webhook",
			Title:       "Test Webhook",
			Description: "🧪 Trigger webhook test delivery",
			Annotations: CombineAnnotations(ReadOnlyAnnotation(), OpenWorldAnnotation()),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":  {Type: "string", Description: "Repository owner"},
					"repo":   {Type: "string", Description: "Repository name"},
					"hook_id": {Type: "number", Description: "Webhook ID"},
				},
				Required: []string{"owner", "repo", "hook_id"},
			},
		},

		// ========================================================================
		// Collaborators (8 tools)
		// ========================================================================
		{
			Name:        "github_list_collaborators",
			Title:       "List Collaborators",
			Description: "👥 List all repository collaborators",
			Annotations: ReadOnlyAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_check_collaborator",
			Title:       "Check Collaborator",
			Description: "✅ Check if user is a collaborator",
			Annotations: ReadOnlyAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":    {Type: "string", Description: "Repository owner"},
					"repo":     {Type: "string", Description: "Repository name"},
					"username": {Type: "string", Description: "GitHub username to check"},
				},
				Required: []string{"owner", "repo", "username"},
			},
		},
		{
			Name:        "github_add_collaborator",
			Title:       "Add Collaborator",
			Description: "🤝 Invite user as collaborator with specific permissions",
			Annotations: ModifyingAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":      {Type: "string", Description: "Repository owner"},
					"repo":       {Type: "string", Description: "Repository name"},
					"username":   {Type: "string", Description: "GitHub username to invite"},
					"permission": {Type: "string", Description: "Permission level: pull, triage, push, maintain, admin"},
					"dry_run":    {Type: "boolean", Description: "Preview (default: true)"},
				},
				Required: []string{"owner", "repo", "username", "permission"},
			},
		},
		{
			Name:        "github_update_collaborator_permission",
			Title:       "Update Collaborator Permission",
			Description: "🔄 Change collaborator's permission level",
			Annotations: ModifyingAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":      {Type: "string", Description: "Repository owner"},
					"repo":       {Type: "string", Description: "Repository name"},
					"username":   {Type: "string", Description: "GitHub username"},
					"permission": {Type: "string", Description: "New permission: pull, triage, push, maintain, admin"},
					"dry_run":    {Type: "boolean", Description: "Preview (default: true)"},
				},
				Required: []string{"owner", "repo", "username", "permission"},
			},
		},
		{
			Name:        "github_remove_collaborator",
			Title:       "Remove Collaborator",
			Description: "❌ Remove collaborator access - HIGH RISK",
			Annotations: DestructiveAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":    {Type: "string", Description: "Repository owner"},
					"repo":     {Type: "string", Description: "Repository name"},
					"username": {Type: "string", Description: "GitHub username to remove"},
					"confirmation_token": {Type: "string", Description: "Confirmation token required"},
				},
				Required: []string{"owner", "repo", "username"},
			},
		},
		{
			Name:        "github_list_invitations",
			Title:       "List Invitations",
			Description: "📨 List pending repository invitations",
			Annotations: ReadOnlyAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_accept_invitation",
			Title:       "Accept Invitation",
			Description: "✔️ Accept repository invitation",
			Annotations: ModifyingAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"invitation_id": {Type: "number", Description: "Invitation ID"},
				},
				Required: []string{"invitation_id"},
			},
		},
		{
			Name:        "github_cancel_invitation",
			Title:       "Cancel Invitation",
			Description: "✖️ Cancel pending invitation",
			Annotations: ModifyingAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":         {Type: "string", Description: "Repository owner"},
					"repo":          {Type: "string", Description: "Repository name"},
					"invitation_id": {Type: "number", Description: "Invitation ID to cancel"},
					"dry_run":       {Type: "boolean", Description: "Preview (default: true)"},
				},
				Required: []string{"owner", "repo", "invitation_id"},
			},
		},

		// ========================================================================
		// Team Access (2 tools)
		// ========================================================================
		{
			Name:        "github_list_repo_teams",
			Title:       "List Repository Teams",
			Description: "🏢 List teams with access to repository",
			Annotations: ReadOnlyAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_add_repo_team",
			Title:       "Add Repository Team",
			Description: "➕ Grant team access to repository",
			Annotations: ModifyingAnnotation(),
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.Property{
					"owner":      {Type: "string", Description: "Repository owner"},
					"repo":       {Type: "string", Description: "Repository name"},
					"team_id":    {Type: "number", Description: "Team ID"},
					"permission": {Type: "string", Description: "Permission: pull, triage, push, maintain, admin"},
					"dry_run":    {Type: "boolean", Description: "Preview (default: true)"},
				},
				Required: []string{"owner", "repo", "team_id", "permission"},
			},
		},
	}
}
