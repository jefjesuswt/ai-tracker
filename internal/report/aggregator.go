package report

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/jefjesuswt/ai-tracker/internal/openrouter"
)

var spanishMonths = []string{
	"",
	"Enero",
	"Febrero",
	"Marzo",
	"Abril",
	"Mayo",
	"Junio",
	"Julio",
	"Agosto",
	"Septiembre",
	"Octubre",
	"Noviembre",
	"Diciembre",
}

type aggregateResult struct {
	Cost      float64
	Tokens    int
	Requests  int
	Breakdown []ModelDetail
	TopModels []string
}

func aggregateMath(activities []openrouter.ActivityItem) (aggregateResult) {
	var totalCost float64
	var totalTokens int
	var totalRequests int

	modelMap := make(map[string]*ModelDetail)

	for _, activity := range activities {
		totalCost += activity.Usage
		totalTokens += (activity.PromptTokens + activity.CompletionTokens + activity.ReasoningTokens)
		totalRequests += activity.Requests

		if _, exists := modelMap[activity.Model]; !exists {
			modelMap[activity.Model] = &ModelDetail{Name: activity.Model}
		}
		modelMap[activity.Model].Reqs += activity.Requests
		modelMap[activity.Model].ModelCost += activity.Usage
	}

	var breakdown []ModelDetail
	for _, detail := range modelMap {
		// Porcentaje de peticiones
		if totalRequests > 0 {
			detail.ReqPercentage = int(math.Round((float64(detail.Reqs) / float64(totalRequests)) * 100))
		}
		// Porcentaje de impacto financiero
		if totalCost > 0 {
			detail.CostPercentage = int(math.Round((detail.ModelCost / totalCost) * 100))
		}
		breakdown = append(breakdown, *detail)
	}

	sort.Slice(breakdown, func(i, j int) bool {
		return breakdown[i].ModelCost > breakdown[j].ModelCost
	})

	var topModels []string
	for i, model := range breakdown {
		if i >= 3 { break }
		topModels = append(topModels, model.Name)
	}

	return aggregateResult{
		Cost:      totalCost,
		Tokens:    totalTokens,
		Requests:  totalRequests,
		Breakdown: breakdown,
		TopModels: topModels,
	}
}

func NewDailyStats(activities []openrouter.ActivityItem, date time.Time) Stats {
	results := aggregateMath(activities)

	return Stats{
		ObsidianFrontmatter: ObsidianFrontmatter{
			Type:      "Daily-AI-Log",
			Date:      date.Format("2006-01-02"),
			Month:     date.Format("2006-01"),
			Cost:      results.Cost,
			Tokens:    results.Tokens,
			Requests:  results.Requests,
			TopModels: results.TopModels,
		},
		FormattedDate:  fmt.Sprintf("%02d de %s, %d", date.Day(), spanishMonths[date.Month()], date.Year()),
		ModelBreakdown: results.Breakdown,
	}
}

func NewMonthlyStats(activities []openrouter.ActivityItem, month time.Time) Stats {
	results := aggregateMath(activities)

	return Stats{
		ObsidianFrontmatter: ObsidianFrontmatter{
			Type:      "Monthly-AI-Log",
			Month:     month.Format("2006-01"),
			Cost:      results.Cost,
			Tokens:    results.Tokens,
			Requests:  results.Requests,
			TopModels: results.TopModels,
		},
		FormattedDate:  fmt.Sprintf("%s, %d", spanishMonths[month.Month()], month.Year()),
		ModelBreakdown: results.Breakdown,
	}
}
