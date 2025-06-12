package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"

	"github.com/jotajotape/github-go-server-mcp/internal/git"
	"github.com/jotajotape/github-go-server-mcp/internal/server"
	"github.com/jotajotape/github-go-server-mcp/internal/types"
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

func NewMCPServer(profile string) (*types.MCPServer, error) {
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
	githubClient := github.NewClient(tc)

	// Detectar entorno Git
	gitConfig := git.DetectGitEnvironment()
	
	// Agregar perfil al gitConfig para logging
	if gitConfig.HasGit {
		log.Printf("🔧 Git environment detected for profile: %s", profile)
	}

	return &types.MCPServer{
		GithubClient: githubClient,
		GitConfig:    gitConfig,
	}, nil
}
