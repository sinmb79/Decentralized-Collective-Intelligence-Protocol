package adapter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// OpenAIAdapter calls the OpenAI Chat Completions API.
type OpenAIAdapter struct {
	baseURL string
	model   string
	apiKey  string
	client  *http.Client
}

type openAIChatRequest struct {
	Model    string              `json:"model"`
	Messages []openAIChatMessage `json:"messages"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message openAIChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NewOpenAIAdapter creates an OpenAI adapter with default endpoint and model values.
func NewOpenAIAdapter(baseURL, model, apiKey string, client *http.Client) *OpenAIAdapter {
	if client == nil {
		client = defaultHTTPClient()
	}

	return &OpenAIAdapter{
		baseURL: normalizeURL(baseURL, DefaultOpenAIURL),
		model:   normalizeModel(model, DefaultOpenAIModel),
		apiKey:  openAIAPIKey(apiKey),
		client:  client,
	}
}

// Infer submits a prompt to the OpenAI Chat Completions API.
func (a *OpenAIAdapter) Infer(query string) (string, error) {
	if !a.IsReady() {
		return "", errors.New("openai adapter is not ready")
	}

	body, err := json.Marshal(openAIChatRequest{
		Model: a.model,
		Messages: []openAIChatMessage{
			{Role: "user", Content: query},
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, a.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if result.Error != nil && result.Error.Message != "" {
			return "", fmt.Errorf("openai API returned status %d: %s", resp.StatusCode, result.Error.Message)
		}
		return "", fmt.Errorf("openai API returned status %d", resp.StatusCode)
	}
	if result.Error != nil && result.Error.Message != "" {
		return "", errors.New(result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", errors.New("openai response contained no choices")
	}

	return result.Choices[0].Message.Content, nil
}

// ModelID returns the configured OpenAI model identifier.
func (a *OpenAIAdapter) ModelID() string {
	return a.model
}

// IsReady reports whether the adapter has the required configuration.
func (a *OpenAIAdapter) IsReady() bool {
	return a != nil && a.client != nil && a.baseURL != "" && a.model != "" && a.apiKey != ""
}
