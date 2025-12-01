// Package dashboard provides GitHub dashboard operations for notifications, alerts and issues.
package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// DashboardClient provides access to GitHub dashboard APIs
type DashboardClient struct {
	HTTPClient *http.Client
	Token      string
	BaseURL    string
}

// NewDashboardClient creates a new dashboard client
func NewDashboardClient(token string) *DashboardClient {
	return &DashboardClient{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		Token:      token,
		BaseURL:    "https://api.github.com",
	}
}

// Notification represents a GitHub notification
type Notification struct {
	ID         string     `json:"id"`
	Unread     bool       `json:"unread"`
	Reason     string     `json:"reason"`
	UpdatedAt  time.Time  `json:"updated_at"`
	LastReadAt *time.Time `json:"last_read_at"`
	Subject    struct {
		Title            string `json:"title"`
		URL              string `json:"url"`
		LatestCommentURL string `json:"latest_comment_url"`
		Type             string `json:"type"`
	} `json:"subject"`
	Repository struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Private  bool   `json:"private"`
		HTMLURL  string `json:"html_url"`
	} `json:"repository"`
	URL             string `json:"url"`
	SubscriptionURL string `json:"subscription_url"`
}

// DependabotAlert represents a Dependabot security alert
type DependabotAlert struct {
	Number                int                   `json:"number"`
	State                 string                `json:"state"`
	CreatedAt             time.Time             `json:"created_at"`
	UpdatedAt             time.Time             `json:"updated_at"`
	HTMLURL               string                `json:"html_url"`
	DismissedAt           *string               `json:"dismissed_at"`
	DismissedReason       *string               `json:"dismissed_reason"`
	DismissedComment      *string               `json:"dismissed_comment"`
	FixedAt               *string               `json:"fixed_at"`
	Dependency            Dependency            `json:"dependency"`
	SecurityAdvisory      SecurityAdvisory      `json:"security_advisory"`
	SecurityVulnerability SecurityVulnerability `json:"security_vulnerability"`
	Repository            *RepositoryInfo       `json:"repository,omitempty"`
}

// Dependency contains package information
type Dependency struct {
	Package struct {
		Ecosystem string `json:"ecosystem"`
		Name      string `json:"name"`
	} `json:"package"`
	ManifestPath string `json:"manifest_path"`
	Scope        string `json:"scope"`
}

// SecurityAdvisory contains advisory information
type SecurityAdvisory struct {
	GHSAID      string `json:"ghsa_id"`
	CVEID       string `json:"cve_id"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	CVSS        struct {
		VectorString string  `json:"vector_string"`
		Score        float64 `json:"score"`
	} `json:"cvss"`
	PublishedAt time.Time `json:"published_at"`
}

// SecurityVulnerability contains vulnerability details
type SecurityVulnerability struct {
	Package struct {
		Ecosystem string `json:"ecosystem"`
		Name      string `json:"name"`
	} `json:"package"`
	Severity               string `json:"severity"`
	VulnerableVersionRange string `json:"vulnerable_version_range"`
	FirstPatchedVersion    *struct {
		Identifier string `json:"identifier"`
	} `json:"first_patched_version"`
}

// RepositoryInfo contains basic repository information
type RepositoryInfo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	HTMLURL  string `json:"html_url"`
}

// Issue represents a GitHub issue
type Issue struct {
	ID          int64           `json:"id"`
	Number      int             `json:"number"`
	Title       string          `json:"title"`
	State       string          `json:"state"`
	HTMLURL     string          `json:"html_url"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Body        string          `json:"body"`
	Labels      []Label         `json:"labels"`
	Assignee    *User           `json:"assignee"`
	Assignees   []User          `json:"assignees"`
	User        User            `json:"user"`
	Comments    int             `json:"comments"`
	Repository  *RepositoryInfo `json:"repository,omitempty"`
	PullRequest *struct {
		URL string `json:"url"`
	} `json:"pull_request,omitempty"`
}

// Label represents a GitHub label
type Label struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// User represents a GitHub user
type User struct {
	Login     string `json:"login"`
	ID        int64  `json:"id"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

// WorkflowRun represents a GitHub Actions workflow run
type WorkflowRun struct {
	ID         int64           `json:"id"`
	Name       string          `json:"name"`
	HeadBranch string          `json:"head_branch"`
	Status     string          `json:"status"`
	Conclusion string          `json:"conclusion"`
	HTMLURL    string          `json:"html_url"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	RunNumber  int             `json:"run_number"`
	Event      string          `json:"event"`
	Repository *RepositoryInfo `json:"repository,omitempty"`
}

// SecretScanningAlert represents a secret scanning alert
type SecretScanningAlert struct {
	Number                int             `json:"number"`
	CreatedAt             time.Time       `json:"created_at"`
	State                 string          `json:"state"`
	SecretType            string          `json:"secret_type"`
	SecretTypeDisplayName string          `json:"secret_type_display_name"`
	HTMLURL               string          `json:"html_url"`
	Resolution            *string         `json:"resolution"`
	ResolvedAt            *string         `json:"resolved_at"`
	Validity              string          `json:"validity"`
	Repository            *RepositoryInfo `json:"repository,omitempty"`
}

// CodeScanningAlert represents a code scanning alert
type CodeScanningAlert struct {
	Number          int       `json:"number"`
	CreatedAt       time.Time `json:"created_at"`
	State           string    `json:"state"`
	HTMLURL         string    `json:"html_url"`
	DismissedAt     *string   `json:"dismissed_at"`
	DismissedReason *string   `json:"dismissed_reason"`
	Rule            struct {
		ID          string `json:"id"`
		Severity    string `json:"severity"`
		Description string `json:"description"`
		Name        string `json:"name"`
	} `json:"rule"`
	Tool struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"tool"`
	MostRecentInstance struct {
		Ref      string `json:"ref"`
		State    string `json:"state"`
		Location struct {
			Path      string `json:"path"`
			StartLine int    `json:"start_line"`
			EndLine   int    `json:"end_line"`
		} `json:"location"`
	} `json:"most_recent_instance"`
	Repository *RepositoryInfo `json:"repository,omitempty"`
}

// DashboardSummary contains a complete summary of all GitHub items requiring attention
type DashboardSummary struct {
	Timestamp           time.Time `json:"timestamp"`
	TotalItems          int       `json:"total_items"`
	UnreadNotifications int       `json:"unread_notifications"`
	OpenIssuesAssigned  int       `json:"open_issues_assigned"`
	PendingPRReviews    int       `json:"pending_pr_reviews"`
	DependabotAlerts    int       `json:"dependabot_alerts"`
	SecretAlerts        int       `json:"secret_alerts"`
	CodeScanningAlerts  int       `json:"code_scanning_alerts"`
	FailedWorkflows     int       `json:"failed_workflows"`

	Notifications       []Notification        `json:"notifications,omitempty"`
	Issues              []Issue               `json:"issues,omitempty"`
	PRsToReview         []Issue               `json:"prs_to_review,omitempty"`
	DependabotList      []DependabotAlert     `json:"dependabot_list,omitempty"`
	SecretAlertsList    []SecretScanningAlert `json:"secret_alerts_list,omitempty"`
	CodeAlertsList      []CodeScanningAlert   `json:"code_alerts_list,omitempty"`
	FailedWorkflowsList []WorkflowRun         `json:"failed_workflows_list,omitempty"`
}

// doRequest performs an HTTP request to the GitHub API
func (d *DashboardClient) doRequest(ctx context.Context, method, endpoint string) ([]byte, error) {
	url := d.BaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+d.Token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %s (status %d)", endpoint, resp.StatusCode)
	}

	var body []byte
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	return body, nil
}

// GetNotifications retrieves all notifications for the authenticated user
func (d *DashboardClient) GetNotifications(ctx context.Context, all bool) ([]Notification, error) {
	endpoint := "/notifications"
	if all {
		endpoint += "?all=true"
	}

	body, err := d.doRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	var notifications []Notification
	if err := json.Unmarshal(body, &notifications); err != nil {
		return nil, fmt.Errorf("failed to parse notifications: %w", err)
	}

	return notifications, nil
}

// GetAssignedIssues retrieves issues assigned to the authenticated user
func (d *DashboardClient) GetAssignedIssues(ctx context.Context) ([]Issue, error) {
	endpoint := "/issues?filter=assigned&state=open&per_page=100"

	body, err := d.doRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	var issues []Issue
	if err := json.Unmarshal(body, &issues); err != nil {
		return nil, fmt.Errorf("failed to parse issues: %w", err)
	}

	// Filter out pull requests (they come mixed with issues)
	var filteredIssues []Issue
	for _, issue := range issues {
		if issue.PullRequest == nil {
			filteredIssues = append(filteredIssues, issue)
		}
	}

	return filteredIssues, nil
}

// GetPRsToReview retrieves PRs where review is requested from the authenticated user
func (d *DashboardClient) GetPRsToReview(ctx context.Context) ([]Issue, error) {
	// Using search API to find PRs where review is requested
	endpoint := "/search/issues?q=is:pr+is:open+review-requested:@me&per_page=100"

	body, err := d.doRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	var result struct {
		TotalCount int     `json:"total_count"`
		Items      []Issue `json:"items"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse PRs: %w", err)
	}

	return result.Items, nil
}

// GetDependabotAlerts retrieves Dependabot alerts for a repository
func (d *DashboardClient) GetDependabotAlerts(ctx context.Context, owner, repo string) ([]DependabotAlert, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/dependabot/alerts?state=open&per_page=100", owner, repo)

	body, err := d.doRequest(ctx, "GET", endpoint)
	if err != nil {
		// 404 means Dependabot is not enabled
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}

	var alerts []DependabotAlert
	if err := json.Unmarshal(body, &alerts); err != nil {
		return nil, fmt.Errorf("failed to parse dependabot alerts: %w", err)
	}

	return alerts, nil
}

// GetSecretScanningAlerts retrieves secret scanning alerts for a repository
func (d *DashboardClient) GetSecretScanningAlerts(ctx context.Context, owner, repo string) ([]SecretScanningAlert, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/secret-scanning/alerts?state=open&per_page=100", owner, repo)

	body, err := d.doRequest(ctx, "GET", endpoint)
	if err != nil {
		// 404 means secret scanning is not enabled
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}

	var alerts []SecretScanningAlert
	if err := json.Unmarshal(body, &alerts); err != nil {
		return nil, fmt.Errorf("failed to parse secret scanning alerts: %w", err)
	}

	return alerts, nil
}

// GetCodeScanningAlerts retrieves code scanning alerts for a repository
func (d *DashboardClient) GetCodeScanningAlerts(ctx context.Context, owner, repo string) ([]CodeScanningAlert, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/code-scanning/alerts?state=open&per_page=100", owner, repo)

	body, err := d.doRequest(ctx, "GET", endpoint)
	if err != nil {
		// 404/403 means code scanning is not enabled
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "403") {
			return nil, nil
		}
		return nil, err
	}

	var alerts []CodeScanningAlert
	if err := json.Unmarshal(body, &alerts); err != nil {
		return nil, fmt.Errorf("failed to parse code scanning alerts: %w", err)
	}

	return alerts, nil
}

// GetFailedWorkflows retrieves failed workflow runs for a repository
func (d *DashboardClient) GetFailedWorkflows(ctx context.Context, owner, repo string) ([]WorkflowRun, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/actions/runs?status=failure&per_page=20", owner, repo)

	body, err := d.doRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	var result struct {
		TotalCount   int           `json:"total_count"`
		WorkflowRuns []WorkflowRun `json:"workflow_runs"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse workflow runs: %w", err)
	}

	return result.WorkflowRuns, nil
}

// GetUserRepos retrieves all repositories for the authenticated user
func (d *DashboardClient) GetUserRepos(ctx context.Context) ([]RepositoryInfo, error) {
	endpoint := "/user/repos?per_page=100&sort=updated"

	body, err := d.doRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	var repos []RepositoryInfo
	if err := json.Unmarshal(body, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse repositories: %w", err)
	}

	return repos, nil
}

// GetFullDashboard retrieves a complete summary of all GitHub items requiring attention
func (d *DashboardClient) GetFullDashboard(ctx context.Context, includeDetails bool) (*DashboardSummary, error) {
	summary := &DashboardSummary{
		Timestamp: time.Now(),
	}

	// Get notifications
	notifications, err := d.GetNotifications(ctx, false)
	if err == nil {
		summary.UnreadNotifications = len(notifications)
		if includeDetails {
			summary.Notifications = notifications
		}
	}

	// Get assigned issues
	issues, err := d.GetAssignedIssues(ctx)
	if err == nil {
		summary.OpenIssuesAssigned = len(issues)
		if includeDetails {
			summary.Issues = issues
		}
	}

	// Get PRs to review
	prs, err := d.GetPRsToReview(ctx)
	if err == nil {
		summary.PendingPRReviews = len(prs)
		if includeDetails {
			summary.PRsToReview = prs
		}
	}

	// Get user repos for security scanning
	repos, err := d.GetUserRepos(ctx)
	if err == nil {
		// Scan first 20 recently updated repos for alerts
		maxRepos := 20
		if len(repos) < maxRepos {
			maxRepos = len(repos)
		}

		for i := 0; i < maxRepos; i++ {
			repo := repos[i]
			parts := strings.Split(repo.FullName, "/")
			if len(parts) != 2 {
				continue
			}
			owner, repoName := parts[0], parts[1]

			// Dependabot alerts
			depAlerts, _ := d.GetDependabotAlerts(ctx, owner, repoName)
			summary.DependabotAlerts += len(depAlerts)
			if includeDetails && len(depAlerts) > 0 {
				for j := range depAlerts {
					depAlerts[j].Repository = &repo
				}
				summary.DependabotList = append(summary.DependabotList, depAlerts...)
			}

			// Secret scanning alerts
			secretAlerts, _ := d.GetSecretScanningAlerts(ctx, owner, repoName)
			summary.SecretAlerts += len(secretAlerts)
			if includeDetails && len(secretAlerts) > 0 {
				for j := range secretAlerts {
					secretAlerts[j].Repository = &repo
				}
				summary.SecretAlertsList = append(summary.SecretAlertsList, secretAlerts...)
			}

			// Code scanning alerts
			codeAlerts, _ := d.GetCodeScanningAlerts(ctx, owner, repoName)
			summary.CodeScanningAlerts += len(codeAlerts)
			if includeDetails && len(codeAlerts) > 0 {
				for j := range codeAlerts {
					codeAlerts[j].Repository = &repo
				}
				summary.CodeAlertsList = append(summary.CodeAlertsList, codeAlerts...)
			}

			// Failed workflows
			failedRuns, _ := d.GetFailedWorkflows(ctx, owner, repoName)
			summary.FailedWorkflows += len(failedRuns)
			if includeDetails && len(failedRuns) > 0 {
				for j := range failedRuns {
					failedRuns[j].Repository = &repo
				}
				summary.FailedWorkflowsList = append(summary.FailedWorkflowsList, failedRuns...)
			}
		}
	}

	summary.TotalItems = summary.UnreadNotifications + summary.OpenIssuesAssigned +
		summary.PendingPRReviews + summary.DependabotAlerts + summary.SecretAlerts +
		summary.CodeScanningAlerts + summary.FailedWorkflows

	return summary, nil
}

// FormatDashboardSummary formats the dashboard summary as a readable string
func FormatDashboardSummary(summary *DashboardSummary, detailed bool) string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════\n")
	sb.WriteString("           🎯 GITHUB DASHBOARD SUMMARY\n")
	sb.WriteString("═══════════════════════════════════════════════════\n\n")

	sb.WriteString(fmt.Sprintf("📅 Generated: %s\n\n", summary.Timestamp.Format("2006-01-02 15:04:05")))

	// Summary counts
	sb.WriteString("📊 OVERVIEW\n")
	sb.WriteString("───────────────────────────────────────────────────\n")

	if summary.TotalItems == 0 {
		sb.WriteString("✅ ¡Todo al día! No hay elementos que requieran atención.\n")
	} else {
		sb.WriteString(fmt.Sprintf("📬 Notificaciones sin leer:    %d\n", summary.UnreadNotifications))
		sb.WriteString(fmt.Sprintf("📋 Issues asignados (abiertos): %d\n", summary.OpenIssuesAssigned))
		sb.WriteString(fmt.Sprintf("👀 PRs pendientes de review:    %d\n", summary.PendingPRReviews))
		sb.WriteString(fmt.Sprintf("🔐 Alertas Dependabot:          %d\n", summary.DependabotAlerts))
		sb.WriteString(fmt.Sprintf("🔑 Alertas de secretos:         %d\n", summary.SecretAlerts))
		sb.WriteString(fmt.Sprintf("🔍 Alertas de código:           %d\n", summary.CodeScanningAlerts))
		sb.WriteString(fmt.Sprintf("❌ Workflows fallidos:          %d\n", summary.FailedWorkflows))
		sb.WriteString("───────────────────────────────────────────────────\n")
		sb.WriteString(fmt.Sprintf("📌 TOTAL ITEMS PENDIENTES:      %d\n", summary.TotalItems))
	}

	if detailed {
		// Notifications
		if len(summary.Notifications) > 0 {
			sb.WriteString("\n\n📬 NOTIFICACIONES\n")
			sb.WriteString("───────────────────────────────────────────────────\n")
			for _, n := range summary.Notifications {
				icon := getNotificationIcon(n.Subject.Type)
				sb.WriteString(fmt.Sprintf("%s [%s] %s\n", icon, n.Repository.FullName, n.Subject.Title))
				sb.WriteString(fmt.Sprintf("   Razón: %s | Actualizado: %s\n", n.Reason, n.UpdatedAt.Format("2006-01-02 15:04")))
			}
		}

		// Issues
		if len(summary.Issues) > 0 {
			sb.WriteString("\n\n📋 ISSUES ASIGNADOS\n")
			sb.WriteString("───────────────────────────────────────────────────\n")
			for _, issue := range summary.Issues {
				labels := formatLabels(issue.Labels)
				sb.WriteString(fmt.Sprintf("#%d: %s %s\n", issue.Number, issue.Title, labels))
				sb.WriteString(fmt.Sprintf("   🔗 %s\n", issue.HTMLURL))
			}
		}

		// PRs to review
		if len(summary.PRsToReview) > 0 {
			sb.WriteString("\n\n👀 PRs PENDIENTES DE REVIEW\n")
			sb.WriteString("───────────────────────────────────────────────────\n")
			for _, pr := range summary.PRsToReview {
				sb.WriteString(fmt.Sprintf("🔀 #%d: %s\n", pr.Number, pr.Title))
				sb.WriteString(fmt.Sprintf("   Por: @%s | 🔗 %s\n", pr.User.Login, pr.HTMLURL))
			}
		}

		// Dependabot alerts
		if len(summary.DependabotList) > 0 {
			sb.WriteString("\n\n🔐 ALERTAS DEPENDABOT\n")
			sb.WriteString("───────────────────────────────────────────────────\n")
			for _, alert := range summary.DependabotList {
				severity := getSeverityIcon(alert.SecurityVulnerability.Severity)
				repoName := ""
				if alert.Repository != nil {
					repoName = alert.Repository.FullName
				}
				sb.WriteString(fmt.Sprintf("%s [%s] %s@%s\n", severity, repoName,
					alert.Dependency.Package.Name, alert.SecurityVulnerability.VulnerableVersionRange))
				sb.WriteString(fmt.Sprintf("   %s (%s)\n", alert.SecurityAdvisory.Summary, alert.SecurityAdvisory.GHSAID))
				if alert.SecurityVulnerability.FirstPatchedVersion != nil {
					sb.WriteString(fmt.Sprintf("   ✅ Fix: actualizar a %s\n", alert.SecurityVulnerability.FirstPatchedVersion.Identifier))
				}
			}
		}

		// Secret scanning alerts
		if len(summary.SecretAlertsList) > 0 {
			sb.WriteString("\n\n🔑 ALERTAS DE SECRETOS EXPUESTOS\n")
			sb.WriteString("───────────────────────────────────────────────────\n")
			for _, alert := range summary.SecretAlertsList {
				repoName := ""
				if alert.Repository != nil {
					repoName = alert.Repository.FullName
				}
				sb.WriteString(fmt.Sprintf("⚠️ [%s] %s detectado\n", repoName, alert.SecretTypeDisplayName))
				sb.WriteString(fmt.Sprintf("   Estado: %s | Validez: %s\n", alert.State, alert.Validity))
				sb.WriteString(fmt.Sprintf("   🔗 %s\n", alert.HTMLURL))
			}
		}

		// Code scanning alerts
		if len(summary.CodeAlertsList) > 0 {
			sb.WriteString("\n\n🔍 ALERTAS DE ANÁLISIS DE CÓDIGO\n")
			sb.WriteString("───────────────────────────────────────────────────\n")
			for _, alert := range summary.CodeAlertsList {
				severity := getSeverityIcon(alert.Rule.Severity)
				repoName := ""
				if alert.Repository != nil {
					repoName = alert.Repository.FullName
				}
				sb.WriteString(fmt.Sprintf("%s [%s] %s\n", severity, repoName, alert.Rule.Description))
				sb.WriteString(fmt.Sprintf("   Herramienta: %s | Archivo: %s:%d\n",
					alert.Tool.Name, alert.MostRecentInstance.Location.Path, alert.MostRecentInstance.Location.StartLine))
			}
		}

		// Failed workflows
		if len(summary.FailedWorkflowsList) > 0 {
			sb.WriteString("\n\n❌ WORKFLOWS FALLIDOS\n")
			sb.WriteString("───────────────────────────────────────────────────\n")
			for _, run := range summary.FailedWorkflowsList {
				repoName := ""
				if run.Repository != nil {
					repoName = run.Repository.FullName
				}
				sb.WriteString(fmt.Sprintf("❌ [%s] %s #%d\n", repoName, run.Name, run.RunNumber))
				sb.WriteString(fmt.Sprintf("   Rama: %s | Evento: %s | %s\n", run.HeadBranch, run.Event, run.UpdatedAt.Format("2006-01-02 15:04")))
				sb.WriteString(fmt.Sprintf("   🔗 %s\n", run.HTMLURL))
			}
		}
	}

	sb.WriteString("\n═══════════════════════════════════════════════════\n")

	return sb.String()
}

func getNotificationIcon(notificationType string) string {
	switch notificationType {
	case "Issue":
		return "📋"
	case "PullRequest":
		return "🔀"
	case "Commit":
		return "📝"
	case "Release":
		return "🚀"
	case "Discussion":
		return "💬"
	case "SecurityAdvisory":
		return "🔐"
	default:
		return "📌"
	}
}

func getSeverityIcon(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	case "error":
		return "🔴"
	case "warning":
		return "🟡"
	default:
		return "⚪"
	}
}

func formatLabels(labels []Label) string {
	if len(labels) == 0 {
		return ""
	}
	var names []string
	for _, l := range labels {
		names = append(names, l.Name)
	}
	return "[" + strings.Join(names, ", ") + "]"
}

// MarkNotificationAsRead marks a notification thread as read
func (d *DashboardClient) MarkNotificationAsRead(ctx context.Context, threadID string) error {
	endpoint := fmt.Sprintf("/notifications/threads/%s", threadID)
	_, err := d.doRequest(ctx, "PATCH", endpoint)
	return err
}

// MarkAllNotificationsAsRead marks all notifications as read
func (d *DashboardClient) MarkAllNotificationsAsRead(ctx context.Context) error {
	_, err := d.doRequest(ctx, "PUT", "/notifications")
	return err
}
