package adapter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// OllamaAdapter calls the Ollama generate API over HTTP.
type OllamaAdapter struct {
	baseURL string
	model   string
	client  *http.Client
}

type ollamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaGenerateResponse struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

// NewOllamaAdapter creates an Ollama adapter with sane defaults.
func NewOllamaAdapter(baseURL, model string, client *http.Client) *OllamaAdapter {
	if client == nil {
		client = defaultHTTPClient()
	}

	return &OllamaAdapter{
		baseURL: normalizeURL(baseURL, DefaultOllamaURL),
		model:   normalizeModel(model, DefaultOllamaModel),
		client:  client,
	}
}

// Infer submits a prompt to the Ollama generate API.
func (a *OllamaAdapter) Infer(query string) (string, error) {
	if !a.IsReady() {
		return "", errors.New("ollama adapter is not ready")
	}

	body, err := json.Marshal(ollamaGenerateRequest{
		Model:  a.model,
		Prompt: query,
		Stream: false,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, a.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("ollama API returned status %d: %s", resp.StatusCode, readErrorBody(resp.Body))
	}

	var result ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", errors.New(result.Error)
	}

	return result.Response, nil
}

// ModelID returns the configured Ollama model identifier.
func (a *OllamaAdapter) ModelID() string {
	return a.model
}

// IsReady reports whether the adapter is configured for requests.
func (a *OllamaAdapter) IsReady() bool {
	return a != nil && a.client != nil && a.baseURL != "" && a.model != ""
}

func readErrorBody(body io.Reader) string {
	data, err := io.ReadAll(io.LimitReader(body, 4<<10))
	if err != nil || len(data) == 0 {
		return http.StatusText(http.StatusInternalServerError)
	}

	return string(bytes.TrimSpace(data))
}
