package main

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/jefjesuswt/ai-tracker/internal/github"
	"github.com/jefjesuswt/ai-tracker/internal/openrouter"
	"github.com/jefjesuswt/ai-tracker/internal/sync"
	"github.com/joho/godotenv"
)

// escapeGitPath reemplaza espacios por %20 pero mantiene intactos los slashes (/)
func escapeGitPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func main() {
	// 1. Cargar entorno
	log.Println("⚙️ [INIT] Cargando variables de entorno...")
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("❌ [FATAL] Error cargando .env: %v", err)
	}

	managementKey := os.Getenv("MANAGEMENT_KEY")
	apiKey := os.Getenv("API_KEY")
	gitToken := os.Getenv("GIT_TOKEN")
	gitOwner := os.Getenv("GIT_OWNER")
	gitRepo := os.Getenv("GIT_REPO")
	obsidianBasePath := os.Getenv("OBSIDIAN_BASE_PATH")

	if obsidianBasePath == "" {
		obsidianBasePath = "10 - AI Related/OpenRouter"
		log.Printf("⚠️ [WARN] OBSIDIAN_BASE_PATH vacío. Usando por defecto: '%s'", obsidianBasePath)
	}

	requiredVars := map[string]string{
		"MANAGEMENT_KEY": managementKey,
		"API_KEY":        apiKey,
		"GIT_TOKEN":      gitToken,
		"GIT_OWNER":      gitOwner,
		"GIT_REPO":       gitRepo,
	}

	for key, value := range requiredVars {
		if value == "" {
			log.Fatalf("❌ [FATAL] Falta la variable de entorno: %s", key)
		}
	}

	log.Println("✅ [INIT] Variables cargadas correctamente. Inicializando clientes HTTP...")
	orClient := openrouter.NewClient(managementKey, apiKey)
	gitClient := github.NewClient(gitToken, gitOwner, gitRepo)

	syncer := sync.NewSyncer(orClient, gitClient, obsidianBasePath)

	if err := syncer.Run(); err != nil {
		log.Fatalf("❌ [FATAL] Error ejecutando Syncer: %v", err)
	}

	log.Println("--------------------------------------------------")
	log.Println("🎉 [DONE] Sincronización completada. El Vault está al día.")
}
