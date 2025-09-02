package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"

	"github.com/scopweb/mcp-go-github/internal/git"
	githubclient "github.com/scopweb/mcp-go-github/internal/github"
	"github.com/scopweb/mcp-go-github/internal/server"
	"github.com/scopweb/mcp-go-github/internal/types"
)

func main() {
	// Configuración de perfiles
	profile := flag.String("profile", "default", "Profile name for this MCP instance")
	flag.Parse()

	log.Printf("🚀 Starting GitHub MCP Server with profile: %s", *profile)

	mcpServer, err := NewMCPServer(*profile)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req types.JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}

		// Usar el handler del paquete server
		resp := server.HandleRequest(mcpServer, req)
		output, err := json.Marshal(resp)
		if err != nil {
			continue
		}

		fmt.Println(string(output))
	}
}

func NewMCPServer(profile string) (*server.MCPServer, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN required for profile: %s", profile)
	}

	// Log del perfil y token (solo primeros 7 caracteres por seguridad)
	tokenPreview := "***"
	if len(token) >= 7 {
		tokenPreview = token[:7] + "***"
	}
	log.Printf("📋 Profile: %s | Token: %s", profile, tokenPreview)

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	ghClient := github.NewClient(tc)

	// Crear cliente Git
	gitClient, err := git.NewClient()
	if err != nil {
		return nil, fmt.Errorf("error creating git client: %w", err)
	}

	// Log del estado de Git
	if gitClient.HasGit() && gitClient.IsGitRepo() {
		log.Printf("🔧 Git local detected for profile: %s", profile)
	} else if gitClient.HasGit() {
		log.Printf("⚠️ Git available but not in repo for profile: %s", profile)
	} else {
		log.Printf("📡 Git not available, API-only mode for profile: %s", profile)
	}

	// Crear cliente GitHub
	githubClient := githubclient.NewClient(ghClient)
	log.Printf("🔧 GitHub client initialized for profile: %s", profile)

	return &server.MCPServer{
		GithubClient: githubClient,
		GitClient:    gitClient,
	}, nil
}
