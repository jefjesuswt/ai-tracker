package sync

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jefjesuswt/ai-tracker/internal/github"
	"github.com/jefjesuswt/ai-tracker/internal/openrouter"
	"github.com/jefjesuswt/ai-tracker/internal/report"
)

type Syncer struct {
	orClient *openrouter.Client
	gitClient *github.Client
	obsidianBasePath string
}

func NewSyncer(orClient *openrouter.Client, gitClient *github.Client, obsidianBasePath string) *Syncer {
	return &Syncer{
		orClient: orClient,
		gitClient: gitClient,
		obsidianBasePath: obsidianBasePath,
	}
}

func (s *Syncer) Run() error {
	log.Println("📥 [OPENROUTER] Descargando historial de los últimos 30 días...")

	activities, err := s.orClient.FetchActivity("")
	if err != nil {
		return fmt.Errorf("error consultando OpenRouter: %w", err)
	}
	log.Printf("📊 [OPENROUTER] Se recibieron %d registros de actividad.", len(activities))

	dailyData, monthlyData := s.groupActivities(activities)

	s.processDailies(dailyData)
	s.processMonthlies(monthlyData)

	return nil
}

func (s *Syncer) groupActivities(activities []openrouter.ActivityItem) (map[string][]openrouter.ActivityItem, map[string][]openrouter.ActivityItem) {
	dailyData := make(map[string][]openrouter.ActivityItem)
	monthlyData := make(map[string][]openrouter.ActivityItem)

	for _, activity := range activities {
		if len(activity.Date) < 10 {
			log.Printf("⚠️ [WARN] Fecha malformada omitida: '%s'", activity.Date)
			continue
		}

		dateOnly := activity.Date[:10] // "2026-02-28"
		dailyData[dateOnly] = append(dailyData[dateOnly], activity)

		monthStr := dateOnly[:7] // "2026-02"
		monthlyData[monthStr] = append(monthlyData[monthStr], activity)
	}

	return dailyData, monthlyData
}

func (s *Syncer) processDailies(dailyData map[string][]openrouter.ActivityItem) {
	log.Println("--------------------------------------------------")
	log.Println("🔄 [SYNC] Iniciando fase de revisión de Dailies (Backfill)...")

	for dateStr, acts := range dailyData {
		log.Printf("🔍 [DAILY] Evaluando fecha: %s (%d registros)", dateStr, len(acts))

		vaultPath := fmt.Sprintf("%s/Days/%s.md", s.obsidianBasePath, dateStr)

		sha, err := s.gitClient.GetFileSha(vaultPath)

		if err != nil {
			log.Printf("❌ [ERROR] Falló la verificación de SHA para %s: %v", vaultPath, err)
			continue
		}

		if sha != "" && len(sha) == 40 {
			log.Printf("⏭️  [DAILY] Archivo %s.md ya existe en Git (SHA: %s). Saltando...", dateStr, sha[:7])
			continue
		}

		if sha != "" && len(sha) != 40 {
			log.Printf("⚠️ [WARN] La API devolvió un SHA inválido ('%s'). Procediendo a crear el archivo...", sha)
		}

		log.Printf("🛠️  [DAILY] Generando reporte faltante para %s...", dateStr)

		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			log.Printf("❌ [ERROR] Parseo de fecha fallido para %s: %v", dateStr, err)
			continue
		}

		stats := report.NewDailyStats(acts, parsedDate)
		statsJSON, _ := json.Marshal(stats.ModelBreakdown)

		log.Printf("🧠 [LLM] Solicitando análisis a DeepSeek para %s...", dateStr)

		summary, err := s.orClient.GenerateSummmary("deepseek/deepseek-chat", statsJSON, openrouter.PromptDaily)
		if err != nil {
			log.Printf("⚠️ [WARN] LLM falló para %s: %v. Usando fallback.", dateStr, err)
			summary = "Análisis IA no disponible temporalmente."
		}

		stats.LLMSummary = strings.TrimSpace(summary)

		markdownStr, err := report.RenderMarkdown(stats)
		if err != nil {
			log.Printf("❌ [ERROR] Renderizado Markdown falló para %s: %v", dateStr, err)
			continue
		}

		log.Printf("☁️  [GIT] Empujando %s.md a Forgejo/GitHub...", dateStr)

		err = s.gitClient.PushFile(vaultPath, markdownStr, fmt.Sprintf("🤖 Bot: Backfill AI Log for %s", dateStr))
		if err != nil {
			log.Printf("❌ [ERROR] Push fallido para %s: %v", dateStr, err)
		} else {
			log.Printf("✅ [DAILY] %s creado con éxito.", dateStr)
		}
	}
}

func (s *Syncer) processMonthlies(monthlyData map[string][]openrouter.ActivityItem) {
	log.Println("--------------------------------------------------")
	log.Println("🔄 [SYNC] Iniciando fase de actualización Mensual...")

	currentMonth := time.Now().UTC().Format("2006-01")

	for monthStr, acts := range monthlyData {
		vaultPath := fmt.Sprintf("%s/Months/%s.md", s.obsidianBasePath, monthStr)

		if monthStr != currentMonth {
			sha, err := s.gitClient.GetFileSha(vaultPath)
			if err != nil {
				log.Printf("❌ [ERROR] Falló la verificación de SHA para %s: %v", vaultPath, err)
				continue
			}

			if sha != "" && len(sha) == 40 {
				log.Printf("⏭️  [MONTHLY] El mes %s ya está cerrado y guardado. Saltando...", monthStr)
				continue
			}
		}

		log.Printf("📅 [MONTHLY] Procesando mes: %s (%d registros acumulados)", monthStr, len(acts))

		parsedMonth, err := time.Parse("2006-01", monthStr)
		if err != nil {
			log.Printf("❌ [ERROR] Parseo de mes fallido para %s: %v", monthStr, err)
			continue
		}

		stats := report.NewMonthlyStats(acts, parsedMonth)
		statsJSON, _ := json.Marshal(stats.ModelBreakdown)

		log.Printf("🧠 [LLM] Solicitando auditoría mensual a DeepSeek para %s...", monthStr)
		summary, err := s.orClient.GenerateSummmary("deepseek/deepseek-chat", statsJSON, openrouter.PromptMonthly)
		if err != nil {
			log.Printf("⚠️ [WARN] LLM falló para el mes %s: %v. Usando fallback.", monthStr, err)
			summary = "Auditoría IA no disponible temporalmente."
		}
		stats.LLMSummary = strings.TrimSpace(summary)

		markdownStr, err := report.RenderMarkdown(stats)
		if err != nil {
			log.Printf("❌ [ERROR] Renderizado Markdown falló para %s: %v", monthStr, err)
			continue
		}

		log.Printf("☁️  [GIT] Sobreescribiendo %s.md en Forgejo/GitHub...", monthStr)
		err = s.gitClient.PushFile(vaultPath, markdownStr, fmt.Sprintf("🤖 Bot: Update Monthly AI Log for %s", monthStr))
		if err != nil {
			log.Printf("❌ [ERROR] Push fallido para el mes %s: %v", monthStr, err)
		} else {
			log.Printf("✅ [MONTHLY] %s actualizado con éxito.", monthStr)
		}
	}
}
