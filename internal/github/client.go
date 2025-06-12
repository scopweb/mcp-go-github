package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v66/github"
)

// ListRepositories lista repositorios del usuario
func ListRepositories(client *github.Client, ctx context.Context, listType string) (string, error) {
	if listType == "" {
		listType = "all"
	}

	opt := &github.RepositoryListOptions{Type: listType}
	repos, _, err := client.Repositories.List(ctx, "", opt)
	if err != nil {
		return "", err
	}

	var result []map[string]interface{}
	for _, repo := range repos {
		result = append(result, map[string]interface{}{
			"name":        repo.GetName(),
			"description": repo.GetDescription(),
			"private":     repo.GetPrivate(),
			"url":         repo.GetHTMLURL(),
			"language":    repo.GetLanguage(),
			"stars":       repo.GetStargazersCount(),
		})
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}

// CreateRepository crea un nuevo repositorio
func CreateRepository(client *github.Client, ctx context.Context, name, description string, private bool) (string, error) {
	repo := &github.Repository{Name: github.String(name)}

	if description != "" {
		repo.Description = github.String(description)
	}

	repo.Private = github.Bool(private)

	createdRepo, _, err := client.Repositories.Create(ctx, "", repo)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Repository '%s' created successfully: %s", createdRepo.GetName(), createdRepo.GetHTMLURL()), nil
}

// ListPullRequests lista pull requests de un repositorio
func ListPullRequests(client *github.Client, ctx context.Context, owner, repoName, state string) (string, error) {
	if state == "" {
		state = "open"
	}

	opt := &github.PullRequestListOptions{State: state}
	prs, _, err := client.PullRequests.List(ctx, owner, repoName, opt)
	if err != nil {
		return "", err
	}

	var result []map[string]interface{}
	for _, pr := range prs {
		result = append(result, map[string]interface{}{
			"number": pr.GetNumber(),
			"title":  pr.GetTitle(),
			"state":  pr.GetState(),
			"url":    pr.GetHTMLURL(),
			"user":   pr.GetUser().GetLogin(),
			"head":   pr.GetHead().GetRef(),
			"base":   pr.GetBase().GetRef(),
		})
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}

// CreatePullRequest crea un nuevo pull request
func CreatePullRequest(client *github.Client, ctx context.Context, owner, repoName, title, body, head, base string) (string, error) {
	pr := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head),
		Base:  github.String(base),
	}

	if body != "" {
		pr.Body = github.String(body)
	}

	createdPR, _, err := client.PullRequests.Create(ctx, owner, repoName, pr)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Pull Request #%d created: %s", createdPR.GetNumber(), createdPR.GetHTMLURL()), nil
}

// CreateFile crea un archivo usando la GitHub API
func CreateFile(client *github.Client, ctx context.Context, owner, repo, path, content, message, branch string) (string, error) {
	if branch == "" {
		branch = "main"
	}

	fileOptions := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: []byte(content),
		Branch:  github.String(branch),
	}

	result, _, err := client.Repositories.CreateFile(ctx, owner, repo, path, fileOptions)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("File '%s' created successfully via API. Commit SHA: %s", path, result.Commit.GetSHA()), nil
}

// UpdateFile actualiza un archivo usando la GitHub API
func UpdateFile(client *github.Client, ctx context.Context, owner, repo, path, content, message, sha, branch string) (string, error) {
	if branch == "" {
		branch = "main"
	}

	fileOptions := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: []byte(content),
		SHA:     github.String(sha),
		Branch:  github.String(branch),
	}

	_, _, err := client.Repositories.UpdateFile(ctx, owner, repo, path, fileOptions)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("File '%s' updated successfully via API", path), nil
}
