package openrouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	ManagementKey string
	APIKey string
 	BaseURL string
	HTTPClient *http.Client
}

func NewClient(managementKey, apiKey string) *Client {
	return &Client{
		ManagementKey: managementKey,
		APIKey: apiKey,
		BaseURL: "https://openrouter.ai/api/v1",
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

func (c *Client) FetchActivity(date string) ([]ActivityItem, error) {
	endpoint := fmt.Sprintf("%s/activity", c.BaseURL)

	if date != "" {
		endpoint = fmt.Sprintf("%s/activity?date=%s", c.BaseURL, date)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.ManagementKey))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error de red conectando a openrouter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openrouter rechazo la petición: %s - Detalles: %s", resp.Status, string(bodyBytes))
	}

	var wrapper ActivityResponse

	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("error decodificando respuesta de openrouter: %w", err)
	}

	return wrapper.Data, nil
}

func (c *Client) GenerateSummmary(model string, statsJSON []byte, systemPrompt SystemPrompt) (string, error) {
	endpoint := fmt.Sprintf("%s/chat/completions", c.BaseURL)

	reqBody := ChatRequest{
		Model: model,
		Messages: []ChatMessage{
			{
				Role: "system",
				Content: string(systemPrompt),
			},
			{
				Role: "user",
				Content: string(statsJSON),
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error serializando el request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("HTTP-Referer", "https://github.com/jefjesuswt/ai-tracker")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error de red conectando a openrouter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openrouter rechazo la petición: %s - Detalles: %s", resp.Status, string(bodyBytes))
	}

	var chatRes ChatResponse

	if err := json.NewDecoder(resp.Body).Decode(&chatRes); err != nil {
		return "", fmt.Errorf("error decodificando respuesta de openrouter: %w", err)
	}

	if len(chatRes.Choices) == 0 {
		return "", fmt.Errorf("no se pudo obtener respuesta de OpenRouter")
	}

	return chatRes.Choices[0].Message.Content, nil
}

func (c *Client) setDatePrompt(isMonthly bool) (string) {
	if (!isMonthly) {
		systemPrompt := `Eres un analista de datos muy directo. Recibirás un JSON con el consumo de IA de mi cuenta de OpenRouter del día de hoy.
Redacta exactamente 2 líneas en español resumiendo el comportamiento, destacando el modelo más usado y el gasto total.
No uses saludos, ni Markdown innecesario. Sé clínico y conciso.`

		return systemPrompt
	}

	systemPrompt := `Eres un analista de datos muy directo. Recibirás un JSON con el consumo de IA de mi cuenta de OpenRouter del mes actual, es decir, 30 días o menos.
Redacta exactamente 2 líneas en español resumiendo el comportamiento, destacando el modelo más usado y el gasto total.
No uses saludos, ni Markdown innecesario. Sé clínico y conciso.`

	return systemPrompt
}
