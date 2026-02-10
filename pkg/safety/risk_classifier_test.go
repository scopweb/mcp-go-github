package safety

import (
	"testing"
)

func TestRiskLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level RiskLevel
		want  string
	}{
		{"Low risk", RiskLow, "LOW"},
		{"Medium risk", RiskMedium, "MEDIUM"},
		{"High risk", RiskHigh, "HIGH"},
		{"Critical risk", RiskCritical, "CRITICAL"},
		{"Unknown risk", RiskLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("RiskLevel.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifyOperation(t *testing.T) {
	tests := []struct {
		name           string
		operation      string
		wantExists     bool
		wantLevel      RiskLevel
		wantCategory   string
		wantDryRun     bool
		wantConfirm    bool
	}{
		{
			name:        "Read-only operation",
			operation:   "github_get_repo_settings",
			wantExists:  true,
			wantLevel:   RiskLow,
			wantCategory: "repository_settings",
			wantDryRun:  false,
			wantConfirm: false,
		},
		{
			name:        "Medium risk operation",
			operation:   "github_add_collaborator",
			wantExists:  true,
			wantLevel:   RiskMedium,
			wantCategory: "collaborators",
			wantDryRun:  true,
			wantConfirm: false,
		},
		{
			name:        "High risk operation",
			operation:   "github_delete_webhook",
			wantExists:  true,
			wantLevel:   RiskHigh,
			wantCategory: "webhooks",
			wantDryRun:  true,
			wantConfirm: true,
		},
		{
			name:        "Critical operation",
			operation:   "github_delete_repository",
			wantExists:  true,
			wantLevel:   RiskCritical,
			wantCategory: "repository_lifecycle",
			wantDryRun:  true,
			wantConfirm: true,
		},
		{
			name:       "Unknown operation",
			operation:  "github_unknown_operation",
			wantExists: false,
		},
		{
			name:       "Non-admin operation",
			operation:  "git_status",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			risk, exists := ClassifyOperation(tt.operation)

			if exists != tt.wantExists {
				t.Errorf("ClassifyOperation(%s) exists = %v, want %v", tt.operation, exists, tt.wantExists)
				return
			}

			if !tt.wantExists {
				return // No need to check risk details
			}

			if risk.Level != tt.wantLevel {
				t.Errorf("Risk level = %v, want %v", risk.Level, tt.wantLevel)
			}
			if risk.Category != tt.wantCategory {
				t.Errorf("Category = %v, want %v", risk.Category, tt.wantCategory)
			}
			if risk.RequiresDryRun != tt.wantDryRun {
				t.Errorf("RequiresDryRun = %v, want %v", risk.RequiresDryRun, tt.wantDryRun)
			}
			if risk.RequiresConfirmation != tt.wantConfirm {
				t.Errorf("RequiresConfirmation = %v, want %v", risk.RequiresConfirmation, tt.wantConfirm)
			}
		})
	}
}

func TestIsAdminOperation(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		want      bool
	}{
		{"Admin operation - repo settings", "github_get_repo_settings", true},
		{"Admin operation - collaborator", "github_add_collaborator", true},
		{"Admin operation - webhook", "github_create_webhook", true},
		{"Admin operation - critical", "github_delete_repository", true},
		{"Non-admin operation - git", "git_status", false},
		{"Non-admin operation - github list", "github_list_repos", false},
		{"Unknown operation", "unknown_operation", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAdminOperation(tt.operation); got != tt.want {
				t.Errorf("IsAdminOperation(%s) = %v, want %v", tt.operation, got, tt.want)
			}
		})
	}
}

func TestGetOperationsByRiskLevel(t *testing.T) {
	tests := []struct {
		name      string
		level     RiskLevel
		wantMin   int // Minimum expected operations
		wantExact string // One operation that should exist
	}{
		{
			name:      "Low risk operations",
			level:     RiskLow,
			wantMin:   4,
			wantExact: "github_get_repo_settings",
		},
		{
			name:      "Medium risk operations",
			level:     RiskMedium,
			wantMin:   5,
			wantExact: "github_add_collaborator",
		},
		{
			name:      "High risk operations",
			level:     RiskHigh,
			wantMin:   2,
			wantExact: "github_delete_webhook",
		},
		{
			name:      "Critical risk operations",
			level:     RiskCritical,
			wantMin:   2,
			wantExact: "github_delete_repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operations := GetOperationsByRiskLevel(tt.level)

			if len(operations) < tt.wantMin {
				t.Errorf("GetOperationsByRiskLevel(%s) returned %d operations, want at least %d",
					tt.level, len(operations), tt.wantMin)
			}

			// Check that expected operation exists
			found := false
			for _, op := range operations {
				if op == tt.wantExact {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected operation '%s' not found in %v level operations", tt.wantExact, tt.level)
			}
		})
	}
}

func TestGetOperationsByCategory(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		wantMin   int
		wantExact string
	}{
		{
			name:      "Repository settings",
			category:  "repository_settings",
			wantMin:   2,
			wantExact: "github_get_repo_settings",
		},
		{
			name:      "Collaborators",
			category:  "collaborators",
			wantMin:   5,
			wantExact: "github_add_collaborator",
		},
		{
			name:      "Webhooks",
			category:  "webhooks",
			wantMin:   4,
			wantExact: "github_create_webhook",
		},
		{
			name:      "Branch protection",
			category:  "branch_protection",
			wantMin:   2,
			wantExact: "github_get_branch_protection",
		},
		{
			name:      "Non-existent category",
			category:  "nonexistent",
			wantMin:   0,
			wantExact: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operations := GetOperationsByCategory(tt.category)

			if len(operations) < tt.wantMin {
				t.Errorf("GetOperationsByCategory(%s) returned %d operations, want at least %d",
					tt.category, len(operations), tt.wantMin)
			}

			if tt.wantExact != "" {
				found := false
				for _, op := range operations {
					if op == tt.wantExact {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected operation '%s' not found in '%s' category", tt.wantExact, tt.category)
				}
			}
		})
	}
}

func TestOperationRiskCompleteness(t *testing.T) {
	// Verify all expected admin operations are classified
	expectedOperations := []string{
		// Repository Settings
		"github_get_repo_settings",
		"github_update_repo_settings",
		"github_archive_repository",
		"github_delete_repository",

		// Branch Protection
		"github_get_branch_protection",
		"github_update_branch_protection",
		"github_delete_branch_protection",

		// Webhooks
		"github_list_webhooks",
		"github_create_webhook",
		"github_update_webhook",
		"github_delete_webhook",
		"github_test_webhook",

		// Collaborators
		"github_list_collaborators",
		"github_add_collaborator",
		"github_update_collaborator_permission",
		"github_remove_collaborator",
		"github_check_collaborator",
		"github_list_invitations",
		"github_accept_invitation",
		"github_cancel_invitation",

		// Teams
		"github_list_repo_teams",
		"github_add_repo_team",
	}

	for _, op := range expectedOperations {
		t.Run(op, func(t *testing.T) {
			risk, exists := ClassifyOperation(op)
			if !exists {
				t.Errorf("Operation %s is not classified", op)
				return
			}

			// Verify risk has required fields
			if risk.Category == "" {
				t.Errorf("Operation %s has no category", op)
			}
			if risk.Description == "" {
				t.Errorf("Operation %s has no description", op)
			}
			if risk.Level < RiskLow || risk.Level > RiskCritical {
				t.Errorf("Operation %s has invalid risk level: %v", op, risk.Level)
			}
		})
	}
}
