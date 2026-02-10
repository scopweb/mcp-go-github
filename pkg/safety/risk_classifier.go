// Package safety provides security filters and risk classification for administrative operations.
package safety

// RiskLevel represents the danger level of an operation
type RiskLevel int

const (
	// RiskLow - Read-only, informational operations with no side effects
	RiskLow RiskLevel = 1

	// RiskMedium - Modifications that are easily reversible
	RiskMedium RiskLevel = 2

	// RiskHigh - Destructive operations that impact team collaboration
	RiskHigh RiskLevel = 3

	// RiskCritical - Irreversible or high security impact operations
	RiskCritical RiskLevel = 4
)

// String returns the string representation of the risk level
func (r RiskLevel) String() string {
	switch r {
	case RiskLow:
		return "LOW"
	case RiskMedium:
		return "MEDIUM"
	case RiskHigh:
		return "HIGH"
	case RiskCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// OperationRisk defines the risk profile and required safeguards for an operation
type OperationRisk struct {
	Level                RiskLevel
	RequiresDryRun       bool
	RequiresConfirmation bool
	RequiresBackup       bool
	RequiresAudit        bool
	Category             string
	Description          string
}

// operationRiskMap defines the risk classification for all administrative operations
var operationRiskMap = map[string]OperationRisk{
	// Repository Settings - Read operations
	"github_get_repo_settings": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "repository_settings",
		Description:          "View repository configuration",
	},
	"github_list_webhooks": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "webhooks",
		Description:          "List repository webhooks",
	},
	"github_test_webhook": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "webhooks",
		Description:          "Trigger webhook test delivery",
	},
	"github_get_branch_protection": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "branch_protection",
		Description:          "View branch protection rules",
	},

	// Repository Settings - Write operations (MEDIUM)
	"github_update_repo_settings": {
		Level:                RiskMedium,
		RequiresDryRun:       true,
		RequiresConfirmation: false,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "repository_settings",
		Description:          "Modify repository configuration",
	},
	"github_create_webhook": {
		Level:                RiskMedium,
		RequiresDryRun:       true,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "webhooks",
		Description:          "Create repository webhook",
	},
	"github_update_webhook": {
		Level:                RiskMedium,
		RequiresDryRun:       true,
		RequiresConfirmation: false,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "webhooks",
		Description:          "Modify webhook configuration",
	},

	// Repository Settings - Destructive operations (HIGH/CRITICAL)
	"github_delete_webhook": {
		Level:                RiskHigh,
		RequiresDryRun:       true,
		RequiresConfirmation: true,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "webhooks",
		Description:          "Delete webhook (breaks integrations)",
	},
	"github_update_branch_protection": {
		Level:                RiskHigh,
		RequiresDryRun:       true,
		RequiresConfirmation: true,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "branch_protection",
		Description:          "Configure branch protection rules",
	},
	"github_delete_branch_protection": {
		Level:                RiskCritical,
		RequiresDryRun:       true,
		RequiresConfirmation: true,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "branch_protection",
		Description:          "Remove branch protection (dangerous)",
	},
	"github_archive_repository": {
		Level:                RiskCritical,
		RequiresDryRun:       true,
		RequiresConfirmation: true,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "repository_lifecycle",
		Description:          "Archive repository (difficult to reverse)",
	},
	"github_delete_repository": {
		Level:                RiskCritical,
		RequiresDryRun:       true,
		RequiresConfirmation: true,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "repository_lifecycle",
		Description:          "Delete repository PERMANENTLY",
	},

	// Collaborator Management - Read operations (LOW)
	"github_list_collaborators": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "List repository collaborators",
	},
	"github_list_invitations": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "View pending invitations",
	},
	"github_check_collaborator": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "Check collaboration status",
	},
	"github_list_repo_teams": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "teams",
		Description:          "List teams with repo access",
	},

	// Collaborator Management - Write operations (MEDIUM)
	"github_add_collaborator": {
		Level:                RiskMedium,
		RequiresDryRun:       true,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "Invite collaborator with permissions",
	},
	"github_update_collaborator_permission": {
		Level:                RiskMedium,
		RequiresDryRun:       true,
		RequiresConfirmation: false,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "Change collaborator access level",
	},
	"github_accept_invitation": {
		Level:                RiskMedium,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "Accept repository invitation",
	},
	"github_cancel_invitation": {
		Level:                RiskMedium,
		RequiresDryRun:       true,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "Cancel pending invitation",
	},
	"github_add_repo_team": {
		Level:                RiskMedium,
		RequiresDryRun:       true,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "teams",
		Description:          "Grant team access to repository",
	},

	// Collaborator Management - Destructive operations (HIGH)
	"github_remove_collaborator": {
		Level:                RiskHigh,
		RequiresDryRun:       true,
		RequiresConfirmation: true,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "Remove collaborator access (loss of access)",
	},
}

// ClassifyOperation returns the risk profile for a given operation
func ClassifyOperation(operation string) (OperationRisk, bool) {
	risk, exists := operationRiskMap[operation]
	return risk, exists
}

// IsAdminOperation checks if an operation is an administrative operation
func IsAdminOperation(operation string) bool {
	_, exists := operationRiskMap[operation]
	return exists
}

// GetOperationsByRiskLevel returns all operations at a specific risk level
func GetOperationsByRiskLevel(level RiskLevel) []string {
	var operations []string
	for op, risk := range operationRiskMap {
		if risk.Level == level {
			operations = append(operations, op)
		}
	}
	return operations
}

// GetOperationsByCategory returns all operations in a specific category
func GetOperationsByCategory(category string) []string {
	var operations []string
	for op, risk := range operationRiskMap{
		if risk.Category == category {
			operations = append(operations, op)
		}
	}
	return operations
}
