package github

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	Token      string
	Owner      string
	Repo       string
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(token, owner, repo string) *Client {
	return &Client{
		Token:   token,
		Owner:   owner,
		Repo:    repo,
		BaseURL: "https://api.github.com",
		HTTPClient: &http.Client{
			Timeout: time.Second * 15,
		},
	}
}

// FIX 2: Añadido omitempty
type fileReq struct {
	Message string `json:"message"`
	Content string `json:"content"`
	Sha     string `json:"sha,omitempty"`
}

type fileInfoRes struct {
	Sha string `json:"sha"`
}

func escapeGitPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func (c *Client) PushFile(path, content, commitMsg string) error {
	escapedPath := escapeGitPath(path)
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.BaseURL, c.Owner, c.Repo, escapedPath)

	sha, err := c.GetFileSha(path)
	if err != nil {
		return fmt.Errorf("error verificando estado del archivo: %w", err)
	}

	encodedContent := base64.StdEncoding.EncodeToString([]byte(content))
	payload := fileReq{
		Message: commitMsg,
		Content: encodedContent,
		Sha:     sha,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error serializando payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error ejecutando request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return fmt.Errorf("status code inesperado de github: %d", res.StatusCode)
	}

	return nil
}

func (c *Client) GetFileSha(path string) (string, error) {
	escapedPath := escapeGitPath(path)
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.BaseURL, c.Owner, c.Repo, escapedPath)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("no se pudo crear el request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return "", nil // Archivo no existe, retornamos vacío
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status inesperado leyendo sha: %d", res.StatusCode)
	}

	var info fileInfoRes
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return "", err
	}

	return info.Sha, nil
}
