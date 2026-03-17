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
		name         string
		operation    string
		wantExists   bool
		wantLevel    RiskLevel
		wantCategory string
		wantDryRun   bool
		wantConfirm  bool
	}{
		{
			name:         "Read-only operation",
			operation:    "github_admin_repo:get_settings",
			wantExists:   true,
			wantLevel:    RiskLow,
			wantCategory: "repository_settings",
			wantDryRun:   false,
			wantConfirm:  false,
		},
		{
			name:         "Medium risk operation",
			operation:    "github_collaborators:add",
			wantExists:   true,
			wantLevel:    RiskMedium,
			wantCategory: "collaborators",
			wantDryRun:   true,
			wantConfirm:  false,
		},
		{
			name:         "High risk operation",
			operation:    "github_webhooks:delete",
			wantExists:   true,
			wantLevel:    RiskHigh,
			wantCategory: "webhooks",
			wantDryRun:   true,
			wantConfirm:  true,
		},
		{
			name:         "Critical operation",
			operation:    "github_admin_repo:delete",
			wantExists:   true,
			wantLevel:    RiskCritical,
			wantCategory: "repository_lifecycle",
			wantDryRun:   true,
			wantConfirm:  true,
		},
		{
			name:       "Unknown operation",
			operation:  "github_unknown:operation",
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
				return
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
		{"Admin tool name", "github_admin_repo", true},
		{"Admin composite key", "github_collaborators:add", true},
		{"Admin composite key - webhook", "github_webhooks:create", true},
		{"Admin composite key - branch protection", "github_branch_protection:delete", true},
		{"Non-admin operation - git", "git_status", false},
		{"Non-admin operation - github repo", "github_repo", false},
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
		wantMin   int
		wantExact string
	}{
		{
			name:      "Low risk operations",
			level:     RiskLow,
			wantMin:   4,
			wantExact: "github_admin_repo:get_settings",
		},
		{
			name:      "Medium risk operations",
			level:     RiskMedium,
			wantMin:   5,
			wantExact: "github_collaborators:add",
		},
		{
			name:      "High risk operations",
			level:     RiskHigh,
			wantMin:   2,
			wantExact: "github_webhooks:delete",
		},
		{
			name:      "Critical risk operations",
			level:     RiskCritical,
			wantMin:   2,
			wantExact: "github_admin_repo:delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operations := GetOperationsByRiskLevel(tt.level)

			if len(operations) < tt.wantMin {
				t.Errorf("GetOperationsByRiskLevel(%s) returned %d operations, want at least %d",
					tt.level, len(operations), tt.wantMin)
			}

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
			wantExact: "github_admin_repo:get_settings",
		},
		{
			name:      "Collaborators",
			category:  "collaborators",
			wantMin:   5,
			wantExact: "github_collaborators:add",
		},
		{
			name:      "Webhooks",
			category:  "webhooks",
			wantMin:   4,
			wantExact: "github_webhooks:create",
		},
		{
			name:      "Branch protection",
			category:  "branch_protection",
			wantMin:   2,
			wantExact: "github_branch_protection:get",
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
	// Verify all expected admin operations are classified (composite keys)
	expectedOperations := []string{
		// Repository Settings
		"github_admin_repo:get_settings",
		"github_admin_repo:update_settings",
		"github_admin_repo:archive",
		"github_admin_repo:delete",

		// Branch Protection
		"github_branch_protection:get",
		"github_branch_protection:update",
		"github_branch_protection:delete",

		// Webhooks
		"github_webhooks:list",
		"github_webhooks:create",
		"github_webhooks:update",
		"github_webhooks:delete",
		"github_webhooks:test",

		// Collaborators
		"github_collaborators:list",
		"github_collaborators:add",
		"github_collaborators:update_permission",
		"github_collaborators:remove",
		"github_collaborators:check",
		"github_collaborators:list_invitations",
		"github_collaborators:accept_invitation",
		"github_collaborators:cancel_invitation",

		// Teams
		"github_collaborators:list_teams",
		"github_collaborators:add_team",
	}

	for _, op := range expectedOperations {
		t.Run(op, func(t *testing.T) {
			risk, exists := ClassifyOperation(op)
			if !exists {
				t.Errorf("Operation %s is not classified", op)
				return
			}

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
