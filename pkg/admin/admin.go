// Package admin provides GitHub administrative operations for repository and collaborator management.
package admin

import (
	"context"

	"github.com/google/go-github/v81/github"
)

// Client wraps the GitHub client for administrative operations
type Client struct {
	client *github.Client
}

// NewClient creates a new administrative client
func NewClient(githubClient *github.Client) *Client {
	return &Client{
		client: githubClient,
	}
}

// ============================================================================
// Repository Settings
// ============================================================================

// GetRepositorySettings retrieves repository configuration
func (c *Client) GetRepositorySettings(ctx context.Context, owner, repo string) (*github.Repository, error) {
	repository, _, err := c.client.Repositories.Get(ctx, owner, repo)
	return repository, err
}

// UpdateRepositorySettings modifies repository configuration
func (c *Client) UpdateRepositorySettings(ctx context.Context, owner, repo string, settings map[string]interface{}) (*github.Repository, error) {
	// Convert settings map to Repository struct
	repoUpdate := &github.Repository{}

	if name, ok := settings["name"].(string); ok {
		repoUpdate.Name = &name
	}
	if description, ok := settings["description"].(string); ok {
		repoUpdate.Description = &description
	}
	if homepage, ok := settings["homepage"].(string); ok {
		repoUpdate.Homepage = &homepage
	}
	if private, ok := settings["private"].(bool); ok {
		repoUpdate.Private = &private
	}
	if hasIssues, ok := settings["has_issues"].(bool); ok {
		repoUpdate.HasIssues = &hasIssues
	}
	if hasProjects, ok := settings["has_projects"].(bool); ok {
		repoUpdate.HasProjects = &hasProjects
	}
	if hasWiki, ok := settings["has_wiki"].(bool); ok {
		repoUpdate.HasWiki = &hasWiki
	}
	if defaultBranch, ok := settings["default_branch"].(string); ok {
		repoUpdate.DefaultBranch = &defaultBranch
	}
	if allowSquashMerge, ok := settings["allow_squash_merge"].(bool); ok {
		repoUpdate.AllowSquashMerge = &allowSquashMerge
	}
	if allowMergeCommit, ok := settings["allow_merge_commit"].(bool); ok {
		repoUpdate.AllowMergeCommit = &allowMergeCommit
	}
	if allowRebaseMerge, ok := settings["allow_rebase_merge"].(bool); ok {
		repoUpdate.AllowRebaseMerge = &allowRebaseMerge
	}
	if deleteBranchOnMerge, ok := settings["delete_branch_on_merge"].(bool); ok {
		repoUpdate.DeleteBranchOnMerge = &deleteBranchOnMerge
	}

	repository, _, err := c.client.Repositories.Edit(ctx, owner, repo, repoUpdate)
	return repository, err
}

// ArchiveRepository archives a repository (makes it read-only)
func (c *Client) ArchiveRepository(ctx context.Context, owner, repo string) (*github.Repository, error) {
	archived := true
	repoUpdate := &github.Repository{
		Archived: &archived,
	}

	repository, _, err := c.client.Repositories.Edit(ctx, owner, repo, repoUpdate)
	return repository, err
}

// DeleteRepository permanently deletes a repository
func (c *Client) DeleteRepository(ctx context.Context, owner, repo string) error {
	_, err := c.client.Repositories.Delete(ctx, owner, repo)
	return err
}

// ============================================================================
// Branch Protection
// ============================================================================

// GetBranchProtection retrieves branch protection rules
func (c *Client) GetBranchProtection(ctx context.Context, owner, repo, branch string) (*github.Protection, error) {
	protection, _, err := c.client.Repositories.GetBranchProtection(ctx, owner, repo, branch)
	return protection, err
}

// UpdateBranchProtection configures branch protection rules
func (c *Client) UpdateBranchProtection(ctx context.Context, owner, repo, branch string, protection *github.ProtectionRequest) (*github.Protection, error) {
	result, _, err := c.client.Repositories.UpdateBranchProtection(ctx, owner, repo, branch, protection)
	return result, err
}

// DeleteBranchProtection removes branch protection
func (c *Client) DeleteBranchProtection(ctx context.Context, owner, repo, branch string) error {
	_, err := c.client.Repositories.RemoveBranchProtection(ctx, owner, repo, branch)
	return err
}

// ============================================================================
// Webhooks
// ============================================================================

// ListWebhooks lists all repository webhooks
func (c *Client) ListWebhooks(ctx context.Context, owner, repo string) ([]*github.Hook, error) {
	hooks, _, err := c.client.Repositories.ListHooks(ctx, owner, repo, nil)
	return hooks, err
}

// CreateWebhook creates a new repository webhook
func (c *Client) CreateWebhook(ctx context.Context, owner, repo string, config map[string]interface{}) (*github.Hook, error) {
	// Convert config map to HookConfig struct
	hookConfig := &github.HookConfig{}

	if url, ok := config["url"].(string); ok {
		hookConfig.URL = &url
	}

	contentType := "json" // Default
	if ct, ok := config["content_type"].(string); ok {
		contentType = ct
	}
	hookConfig.ContentType = &contentType

	if secret, ok := config["secret"].(string); ok {
		hookConfig.Secret = &secret
	}

	insecureSSL := "0" // Default (secure)
	if ssl, ok := config["insecure_ssl"].(string); ok {
		insecureSSL = ssl
	}
	hookConfig.InsecureSSL = &insecureSSL

	var events []string
	if evts, ok := config["events"].([]string); ok {
		events = evts
	} else {
		events = []string{"push"} // Default event
	}

	active := true
	if act, ok := config["active"].(bool); ok {
		active = act
	}

	hook := &github.Hook{
		Events: events,
		Active: &active,
		Config: hookConfig,
	}

	result, _, err := c.client.Repositories.CreateHook(ctx, owner, repo, hook)
	return result, err
}

// UpdateWebhook modifies an existing webhook
func (c *Client) UpdateWebhook(ctx context.Context, owner, repo string, hookID int64, config map[string]interface{}) (*github.Hook, error) {
	// Get existing hook
	existingHook, _, err := c.client.Repositories.GetHook(ctx, owner, repo, hookID)
	if err != nil {
		return nil, err
	}

	// Initialize Config if nil
	if existingHook.Config == nil {
		existingHook.Config = &github.HookConfig{}
	}

	// Update config fields
	if url, ok := config["url"].(string); ok {
		existingHook.Config.URL = &url
	}
	if contentType, ok := config["content_type"].(string); ok {
		existingHook.Config.ContentType = &contentType
	}
	if secret, ok := config["secret"].(string); ok {
		existingHook.Config.Secret = &secret
	}
	if insecureSSL, ok := config["insecure_ssl"].(string); ok {
		existingHook.Config.InsecureSSL = &insecureSSL
	}
	if events, ok := config["events"].([]string); ok {
		existingHook.Events = events
	}
	if active, ok := config["active"].(bool); ok {
		existingHook.Active = &active
	}

	result, _, err := c.client.Repositories.EditHook(ctx, owner, repo, hookID, existingHook)
	return result, err
}

// DeleteWebhook deletes a repository webhook
func (c *Client) DeleteWebhook(ctx context.Context, owner, repo string, hookID int64) error {
	_, err := c.client.Repositories.DeleteHook(ctx, owner, repo, hookID)
	return err
}

// TestWebhook triggers a test delivery for a webhook
func (c *Client) TestWebhook(ctx context.Context, owner, repo string, hookID int64) error {
	_, err := c.client.Repositories.TestHook(ctx, owner, repo, hookID)
	return err
}

// ============================================================================
// Collaborators
// ============================================================================

// ListCollaborators lists all repository collaborators
func (c *Client) ListCollaborators(ctx context.Context, owner, repo string) ([]*github.User, error) {
	collaborators, _, err := c.client.Repositories.ListCollaborators(ctx, owner, repo, nil)
	return collaborators, err
}

// AddCollaborator invites a user as a collaborator
func (c *Client) AddCollaborator(ctx context.Context, owner, repo, username, permission string) (*github.CollaboratorInvitation, error) {
	opts := &github.RepositoryAddCollaboratorOptions{
		Permission: permission,
	}

	invitation, _, err := c.client.Repositories.AddCollaborator(ctx, owner, repo, username, opts)
	return invitation, err
}

// UpdateCollaboratorPermission updates a collaborator's permission level
func (c *Client) UpdateCollaboratorPermission(ctx context.Context, owner, repo, username, permission string) (*github.CollaboratorInvitation, error) {
	// GitHub API uses AddCollaborator for both adding and updating
	return c.AddCollaborator(ctx, owner, repo, username, permission)
}

// RemoveCollaborator removes a user's access to a repository
func (c *Client) RemoveCollaborator(ctx context.Context, owner, repo, username string) error {
	_, err := c.client.Repositories.RemoveCollaborator(ctx, owner, repo, username)
	return err
}

// CheckCollaborator checks if a user is a collaborator
func (c *Client) CheckCollaborator(ctx context.Context, owner, repo, username string) (bool, error) {
	isCollaborator, _, err := c.client.Repositories.IsCollaborator(ctx, owner, repo, username)
	return isCollaborator, err
}

// ============================================================================
// Repository Invitations
// ============================================================================

// ListInvitations lists pending repository invitations
func (c *Client) ListInvitations(ctx context.Context, owner, repo string) ([]*github.RepositoryInvitation, error) {
	invitations, _, err := c.client.Repositories.ListInvitations(ctx, owner, repo, nil)
	return invitations, err
}

// AcceptInvitation accepts a repository invitation
func (c *Client) AcceptInvitation(ctx context.Context, invitationID int64) error {
	_, err := c.client.Users.AcceptInvitation(ctx, invitationID)
	return err
}

// CancelInvitation declines or cancels a repository invitation
func (c *Client) CancelInvitation(ctx context.Context, owner, repo string, invitationID int64) error {
	_, err := c.client.Repositories.DeleteInvitation(ctx, owner, repo, invitationID)
	return err
}

// ============================================================================
// Team Access
// ============================================================================

// ListRepositoryTeams lists teams with access to a repository
func (c *Client) ListRepositoryTeams(ctx context.Context, owner, repo string) ([]*github.Team, error) {
	teams, _, err := c.client.Repositories.ListTeams(ctx, owner, repo, nil)
	return teams, err
}

// AddRepositoryTeam grants a team access to a repository
func (c *Client) AddRepositoryTeam(ctx context.Context, owner, repo string, teamID int64, permission string) error {
	opts := &github.TeamAddTeamRepoOptions{
		Permission: permission,
	}

	// Get organization from owner (assuming owner is org)
	_, err := c.client.Teams.AddTeamRepoByID(ctx, 0, teamID, owner, repo, opts)
	return err
}
