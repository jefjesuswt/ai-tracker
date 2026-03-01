package openrouter

type ActivityItem struct {
	Date             string  `json:"date"`
	Model            string  `json:"model"`
	ModelPermaslug   string  `json:"model_permaslug"`
	EndpointID       string  `json:"endpoint_id"`
	ProviderName     string  `json:"provider_name"`
	Usage            float64 `json:"usage"`
	ByokUsage        float64 `json:"byok_usage_inference"`
	Requests         int     `json:"requests"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	ReasoningTokens  int     `json:"reasoning_tokens"`
}

type ActivityResponse struct {
	Data []ActivityItem `json:"data"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string   `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatResponse struct {
	Choices []struct{
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

type SystemPrompt string

const (
	PromptDaily SystemPrompt = `Eres un auditor de costos de API. Analiza este JSON de consumo diario de LLMs. Identifica el modelo que generó el mayor gasto financiero (cost_percentage) y compáralo con el modelo de mayor volumen de uso (req_percentage). Si un modelo genera mucho gasto pero poco uso (o viceversa), destácalo. Redacta un solo párrafo analítico y directo de máximo 3 líneas. Usa negritas para los nombres de los modelos y las cifras de dinero en dólares. Sé explícito, no generalices.`

	PromptMonthly SystemPrompt = `Eres un auditor financiero de infraestructura IA. Analiza este JSON acumulado mensual. Tu objetivo es explicar a dónde se fue el dinero. Destaca el modelo que representa el mayor gasto mensual, señala el modelo de mayor volumen (requests), y evalúa si el retorno financiero tiene sentido basado en la distribución. Redacta un análisis explícito en un párrafo de 4 a 5 líneas con datos duros. Usa negritas para resaltar modelos y dólares. Cero saludos, directo a los datos.`
)
