---
name: mcp-github
description: "GitHub operations via mcp-go-github MCP server: 82 tools for Git, GitHub API, dashboard, admin, and safety-controlled operations. Use when: (1) managing repos, branches, PRs, issues via MCP tools, (2) running Git workflows (commit, push, merge, rebase, conflicts), (3) administering repos (settings, webhooks, collaborators, branch protection), (4) viewing dashboard (notifications, assigned issues, PRs to review, security alerts), (5) performing hybrid file operations (Git-first, API fallback). NOT for: direct gh CLI usage (use github skill), non-GitHub platforms (GitLab, Bitbucket), or operations outside MCP protocol."
metadata:
  {
    "openclaw":
      {
        "emoji": "🔧",
        "requires": { "bins": ["mcp-go-github"] },
        "install":
          [
            {
              "id": "go-build",
              "kind": "command",
              "command": "go build -o mcp-go-github ./cmd/github-mcp-server/",
              "bins": ["mcp-go-github"],
              "label": "Build mcp-go-github from source (Go 1.25+)"
            }
          ]
      }
  }
---

# MCP GitHub Skill

Use the **mcp-go-github** MCP server to interact with GitHub repositories, Git workflows, pull requests, issues, dashboards, and administrative operations — all through Claude Desktop via MCP protocol.

This skill provides **86 tools** organized in categories: Git operations, GitHub API, dashboard, hybrid file ops, Git-free file ops, response/repair, and full admin controls with a 4-tier safety system.

## When to Use

**USE this skill when:**

- Managing local Git repositories (status, commit, push, pull, branch, merge, rebase)
- Resolving merge conflicts safely with backup support
- Creating or reviewing pull requests and issues via MCP
- Viewing your GitHub dashboard (notifications, assigned issues, PRs to review)
- Managing repository settings, webhooks, collaborators, or branch protection
- Creating or updating files with hybrid Git-first/API-fallback strategy
- Dismissing security alerts (Dependabot, Code Scanning, Secret Scanning)
- Re-running failed GitHub Actions workflows
- Performing administrative operations with safety controls

## When NOT to Use

**DON'T use this skill when:**

- Using `gh` CLI directly in terminal -> use the `github` skill instead
- Working with non-GitHub platforms (GitLab, Bitbucket) -> different tools required
- Doing local-only file editing without Git context -> use file tools directly
- Complex web UI interactions (GitHub settings pages) -> use browser
- Bulk scripting across many repos -> use `gh api` with scripting

## Setup

### 1. Build the Server

```bash
# Clone and build
git clone https://github.com/scopweb/mcp-go-github.git
cd mcp-go-github
go build -o mcp-go-github ./cmd/github-mcp-server/
```

### 2. Generate GitHub Token

1. Go to [GitHub Settings > Personal Access Tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Select scopes:
   - `repo` (required) - Full control of private repositories
   - `delete_repo` (optional) - For repository deletion
   - `workflow` (optional) - For re-running GitHub Actions
   - `security_events` (optional) - For security alert dismissal
   - `admin:repo_hook` (optional) - Enhanced webhook management
   - `admin:org` (optional) - Team management in organizations

### 3. Configure Claude Desktop

```json
{
  "mcpServers": {
    "github": {
      "command": "/path/to/mcp-go-github",
      "args": ["--profile", "default"],
      "env": {
        "GITHUB_TOKEN": "ghp_your_token_here"
      }
    }
  }
}
```

**Multi-profile setup** (personal + work accounts):

```json
{
  "mcpServers": {
    "github-personal": {
      "command": "/path/to/mcp-go-github",
      "args": ["--profile", "personal"],
      "env": { "GITHUB_TOKEN": "ghp_personal_token" }
    },
    "github-work": {
      "command": "/path/to/mcp-go-github",
      "args": ["--profile", "work"],
      "env": { "GITHUB_TOKEN": "ghp_work_token" }
    }
  }
}
```

### 4. Safety Configuration (Optional)

Create `safety.json` next to the executable:

```json
{
  "mode": "moderate",
  "enable_audit_log": true,
  "require_confirmation_above": 3,
  "audit_log_path": "./mcp-admin-audit.log"
}
```

Safety modes: `strict` (confirms MEDIUM+), `moderate` (confirms HIGH+, recommended), `permissive` (CRITICAL only), `disabled`.

## Tool Reference

### Information Tools (8)

```
git_status              -> Working directory status
git_list_files          -> List tracked files (optional ref)
git_get_file_content    -> Read file content (optional ref/branch)
git_get_file_sha        -> Get file SHA hash
git_get_last_commit     -> Last commit details
git_get_changed_files   -> Changed files (staged or unstaged)
git_validate_repo       -> Validate Git repository at path
git_context             -> Auto-detect full repo context
```

**Usage pattern — check repo state before working:**
```
1. git_set_workspace {"path": "/path/to/repo"}
2. git_status
3. git_branch_list {"remote": true}
4. git_get_changed_files {"staged": false}
```

### Basic Git Tools (6)

```
git_set_workspace       -> Set working directory for Git operations
git_add                 -> Stage files ("." for all, or specific paths)
git_commit              -> Commit staged changes with message
git_push                -> Push to remote branch
git_pull                -> Pull from remote branch
git_checkout            -> Switch or create branches
```

**Usage pattern — standard commit flow:**
```
1. git_set_workspace {"path": "/repo"}
2. git_add {"files": "."}
3. git_commit {"message": "feat: add new feature"}
4. git_push {"branch": "main"}
```

### Advanced Git Tools (12)

```
git_log_analysis        -> Commit history analysis (configurable limit)
git_diff_files          -> View file diffs (staged/unstaged)
git_branch_list         -> List branches (local/remote)
git_stash               -> Stash operations (save/pop/list/drop)
git_remote              -> Remote management (add/remove/list)
git_tag                 -> Tag operations (create/list/delete)
git_clean               -> Clean untracked files (dry_run default: true)
git_checkout_remote     -> Checkout remote branch locally
git_merge               -> Merge branches
git_rebase              -> Rebase current branch
git_force_push          -> Force push (with --force-with-lease)
git_push_upstream       -> Push and set upstream tracking
git_sync_with_remote    -> Sync local branch with remote
git_pull_with_strategy  -> Pull with merge strategy (merge/rebase/ff-only)
```

**Usage pattern — feature branch workflow:**
```
1. git_checkout {"branch": "feature/new-thing", "create": true}
2. ... make changes ...
3. git_add {"files": "."}
4. git_commit {"message": "feat: implement new thing"}
5. git_push_upstream {"branch": "feature/new-thing"}
```

### Conflict Management Tools (6)

```
git_safe_merge          -> Merge with automatic backup creation
git_conflict_status     -> View current conflict state
git_resolve_conflicts   -> Resolve with strategy (ours/theirs/manual)
git_validate_clean_state -> Check if working directory is clean
git_detect_conflicts    -> Preview conflicts BEFORE merging
git_create_backup       -> Create named backup branch
```

**Usage pattern — safe merge with conflict detection:**
```
1. git_detect_conflicts {"source_branch": "feature/x", "target_branch": "main"}
2. git_create_backup {"name": "before-merge"}
3. git_safe_merge {"source": "feature/x", "target": "main"}
4. git_conflict_status
5. git_resolve_conflicts {"strategy": "theirs"}   # if conflicts exist
```

### Hybrid File Tools (2)

```
create_file             -> Create file (Git-first, GitHub API fallback)
update_file             -> Update file (Git-first, GitHub API fallback)
```

These tools try local Git operations first (zero API tokens consumed). If Git is unavailable, they fall back to the GitHub API automatically.

### Git-Free File Tools (4)

```
github_list_repo_contents -> List files/directories in a repo path (no Git required)
github_download_file      -> Download a single file to local disk (no Git required)
github_download_repo      -> Clone entire repo via API (no Git required)
github_pull_repo          -> Update local directory from repo via API (no Git required)
```

These tools work entirely through the GitHub API, enabling repository access on systems without Git installed.

**Usage pattern — work without Git:**
```
1. github_list_repo_contents {"owner": "myorg", "repo": "myapp", "path": "src"}
2. github_download_file {"owner": "myorg", "repo": "myapp", "path": "src/main.go"}
3. github_download_repo {"owner": "myorg", "repo": "myapp", "branch": "main", "local_dir": "./myapp"}
4. github_pull_repo {"owner": "myorg", "repo": "myapp", "local_dir": "./myapp"}
```

### GitHub API Tools (4)

```
github_list_repos       -> List repositories (type: all/owner/public/private/member)
github_create_repo      -> Create new repository
github_list_prs         -> List pull requests (state: open/closed/all)
github_create_pr        -> Create pull request (owner, repo, title, body, head, base)
```

**Usage pattern — create PR from feature branch:**
```
1. github_create_pr {
     "owner": "myorg",
     "repo": "myproject",
     "title": "feat: add authentication",
     "body": "Implements OAuth2 authentication flow",
     "head": "feature/auth",
     "base": "main"
   }
```

### Dashboard Tools (7)

```
github_dashboard        -> Full dashboard overview (notifications + issues + PRs + alerts)
github_notifications    -> List notifications (all: true for read+unread)
github_assigned_issues  -> Issues assigned to you
github_prs_to_review    -> PRs awaiting your review
github_security_alerts  -> Security alerts (owner, repo, type: dependabot/code/secret/all)
github_failed_workflows -> Failed CI/CD workflow runs
github_mark_notification_read -> Mark notification as read (thread_id)
```

**Usage pattern — morning triage:**
```
1. github_dashboard                    # overview
2. github_prs_to_review               # what needs review
3. github_assigned_issues              # your issues
4. github_security_alerts {"owner": "myorg", "repo": "myapp", "type": "all"}
```

### Response Tools (3)

```
github_comment_issue    -> Comment on issue (owner, repo, issue_number, body)
github_comment_pr       -> Comment on PR (owner, repo, pr_number, body)
github_review_pr        -> Review PR (owner, repo, pr_number, body, event: APPROVE/REQUEST_CHANGES/COMMENT)
```

**Usage pattern — PR review:**
```
1. github_list_prs {"owner": "myorg", "repo": "myapp", "state": "open"}
2. github_review_pr {
     "owner": "myorg",
     "repo": "myapp",
     "pr_number": 42,
     "body": "LGTM! Clean implementation.",
     "event": "APPROVE"
   }
```

### Repair Tools (6)

```
github_close_issue          -> Close issue with optional comment
github_merge_pr             -> Merge PR (method: merge/squash/rebase)
github_rerun_workflow       -> Re-run failed GitHub Actions workflow
github_dismiss_dependabot_alert -> Dismiss Dependabot alert (reason: fix_started/no_bandwidth/not_used/tolerable_risk)
github_dismiss_code_alert       -> Dismiss Code Scanning alert
github_dismiss_secret_alert     -> Dismiss Secret Scanning alert (resolution: false_positive/wont_fix/revoked/used_in_tests)
```

**Usage pattern — fix failed CI and merge:**
```
1. github_failed_workflows
2. github_rerun_workflow {"owner": "myorg", "repo": "myapp", "run_id": 12345}
3. github_merge_pr {"owner": "myorg", "repo": "myapp", "pr_number": 42, "method": "squash"}
```

### Repository Admin Tools (4)

```
github_get_repo_settings     -> View repository configuration
github_update_repo_settings  -> Modify settings (name, description, visibility, features, merge options)
github_archive_repository    -> Archive repository (read-only) [CRITICAL - requires confirmation]
github_delete_repository     -> Delete repository permanently [CRITICAL - requires confirmation]
```

**Note:** Destructive operations (`archive`, `delete`) require a **confirmation token**. First call returns the token, second call with the token executes.

### Branch Protection Tools (3)

```
github_get_branch_protection    -> View protection rules for a branch
github_update_branch_protection -> Configure protection (required reviews, status checks, restrictions)
github_delete_branch_protection -> Remove all protection rules [HIGH - requires confirmation in strict mode]
```

**Usage pattern — protect main branch:**
```
1. github_update_branch_protection {
     "owner": "myorg",
     "repo": "myapp",
     "branch": "main",
     "required_approving_review_count": 2,
     "require_status_checks": true,
     "required_status_checks": ["ci/build", "ci/test"],
     "enforce_admins": true
   }
```

### Webhook Tools (5)

```
github_list_webhooks    -> List all repository webhooks
github_create_webhook   -> Create webhook (url, events, content_type, secret)
github_update_webhook   -> Update existing webhook
github_delete_webhook   -> Remove webhook [HIGH]
github_test_webhook     -> Send test delivery
```

### Collaborator Tools (8)

```
github_list_collaborators           -> View all collaborators
github_check_collaborator           -> Check user's access level
github_add_collaborator             -> Invite user (permission: pull/triage/push/maintain/admin)
github_update_collaborator_permission -> Change permission level
github_remove_collaborator          -> Revoke access [HIGH]
github_list_invitations             -> View pending invitations
github_accept_invitation            -> Accept repository invitation
github_cancel_invitation            -> Cancel pending invitation
```

### Team Tools (2)

```
github_list_repo_teams  -> View teams with repository access
github_add_repo_team    -> Grant team access (permission: pull/triage/push/maintain/admin)
```

## Workflows

### Daily Developer Workflow

```
# 1. Morning triage
github_dashboard

# 2. Start working on assigned issue
git_set_workspace {"path": "/repo"}
git_checkout {"branch": "fix/issue-123", "create": true}

# 3. Make changes, commit, push
git_add {"files": "."}
git_commit {"message": "fix: resolve issue #123"}
git_push_upstream {"branch": "fix/issue-123"}

# 4. Create PR
github_create_pr {
  "owner": "myorg", "repo": "myapp",
  "title": "fix: resolve issue #123",
  "body": "Closes #123",
  "head": "fix/issue-123", "base": "main"
}

# 5. Comment on the issue
github_comment_issue {
  "owner": "myorg", "repo": "myapp",
  "issue_number": 123,
  "body": "Fix submitted in PR #456"
}
```

### Safe Merge Workflow

```
# 1. Check for conflicts before merging
git_detect_conflicts {"source_branch": "feature/x", "target_branch": "main"}

# 2. Create backup
git_create_backup {"name": "pre-merge-backup"}

# 3. Perform safe merge (creates automatic backup)
git_safe_merge {"source": "feature/x", "target": "main"}

# 4. If conflicts: review and resolve
git_conflict_status
git_resolve_conflicts {"strategy": "theirs"}

# 5. Push merged result
git_push {"branch": "main"}
```

### Repository Setup Workflow

```
# 1. Create repository
github_create_repo {"name": "new-project", "description": "My new project", "private": true}

# 2. Protect main branch
github_update_branch_protection {
  "owner": "myorg", "repo": "new-project", "branch": "main",
  "required_approving_review_count": 1,
  "require_status_checks": true
}

# 3. Add collaborators
github_add_collaborator {
  "owner": "myorg", "repo": "new-project",
  "username": "teammate", "permission": "push"
}

# 4. Set up webhook for CI
github_create_webhook {
  "owner": "myorg", "repo": "new-project",
  "url": "https://ci.example.com/webhook",
  "events": ["push", "pull_request"],
  "content_type": "json"
}
```

### Security Triage Workflow

```
# 1. Check all security alerts
github_security_alerts {"owner": "myorg", "repo": "myapp", "type": "all"}

# 2. Dismiss false positives
github_dismiss_dependabot_alert {
  "owner": "myorg", "repo": "myapp",
  "alert_number": 5,
  "reason": "not_used"
}

github_dismiss_secret_alert {
  "owner": "myorg", "repo": "myapp",
  "alert_number": 2,
  "resolution": "false_positive"
}

# 3. Re-run failed security scans
github_rerun_workflow {"owner": "myorg", "repo": "myapp", "run_id": 98765}
```

## Safety System

All administrative tools go through a 4-tier safety system:

| Risk Level | Examples | Behavior |
|---|---|---|
| **LOW** (1) | Read settings, list collaborators | Execute immediately |
| **MEDIUM** (2) | Update settings, create webhook | Dry-run in strict mode |
| **HIGH** (3) | Remove collaborator, delete webhook | Requires confirmation token |
| **CRITICAL** (4) | Delete repository, archive repository | Requires confirmation + backup recommendation |

**Confirmation flow for destructive operations:**
```
# Step 1: Call without token -> get confirmation token
github_delete_repository {"owner": "myorg", "repo": "old-project"}
# Response: "Confirmation required. Token: abc123... (expires in 5 min)"

# Step 2: Call with token -> execute
github_delete_repository {"owner": "myorg", "repo": "old-project", "confirmation_token": "abc123..."}
```

## Key Differences from `gh` CLI Skill

| Feature | `gh` CLI | `mcp-github` |
|---|---|---|
| **Protocol** | Shell commands | MCP JSON-RPC 2.0 |
| **Integration** | Terminal | Claude Desktop native |
| **Git operations** | Separate (`git` binary) | Built-in (45+ tools) |
| **Token cost** | N/A | Optimized (Git-first = 0 tokens) |
| **Safety** | None | 4-tier with audit logging |
| **Admin tools** | Limited | 22 tools with confirmation |
| **Dashboard** | Manual queries | Unified dashboard view |
| **Multi-account** | gh auth switch | `--profile` flag |
| **Conflict mgmt** | Manual | Automated (detect, backup, resolve) |

## Notes

- The server auto-detects Git availability and filters tools accordingly (82 with Git, 48 without)
- Protocol version is auto-detected from the client — universal MCP compatibility
- `git_set_workspace` must be called before any local Git operations
- All GitHub API tools require `GITHUB_TOKEN` in the environment
- Hybrid tools (`create_file`, `update_file`) try Git first, then fall back to API
- Dashboard tools create a separate authenticated client internally
- Audit logs rotate automatically at 10MB with 5 backup files
