package report

import (
	"bytes"
	"fmt"
	"text/template"
)

// Solucionamos un pequeño bug en el YAML: si es Monthly, Date viene vacío, así que usamos un if.
const obsidianTemplate = `---
type: {{.Type}}{{if .Date}}
date: {{.Date}}{{end}}
month: "{{.Month}}"
cost: {{printf "%.4f" .Cost}}
tokens: {{.Tokens}}
requests: {{.Requests}}
top_models: [{{range $i, $m := .TopModels}}{{if $i}}, {{end}}"{{$m}}"{{end}}]
---
> [!abstract] **Breadcrumbs**
> [[{{if eq .Type "Monthly"}}Months{{else}}Days{{end}}|🤖 {{if eq .Type "Monthly"}}Months{{else}}Days{{end}}]]
# 🤖 Reporte: {{.FormattedDate}}

> [!insight] **AI Summary**
> {{.LLMSummary}}

### 📊 Desglose de Modelos
| Modelo | Peticiones | Costo ($) | % Peticiones | % Costo |
| :--- | :--- | :--- | :--- | :--- |
{{- range .ModelBreakdown}}
| **{{.Name}}** | {{.Reqs}} | {{printf "%.4f" .ModelCost}} | {{.ReqPercentage}}% | <progress value="{{.CostPercentage}}" max="100" style="width: 100px; border-radius: 5px;"></progress> **{{.CostPercentage}}%** |
{{- end}}
`

func RenderMarkdown(stats Stats) (string, error) {
	t, err := template.New("obsidian").Parse(obsidianTemplate)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, stats); err != nil {
		return "", fmt.Errorf("error rendering template: %w", err)
	}

	return buf.String(), nil
}
