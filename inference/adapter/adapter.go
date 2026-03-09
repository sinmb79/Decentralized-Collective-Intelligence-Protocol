package adapter

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	KindEcho           = "echo"
	KindOllama         = "ollama"
	KindOpenAI         = "openai"
	DefaultOllamaURL   = "http://localhost:11434"
	DefaultOllamaModel = "llama3"
	DefaultOpenAIURL   = "https://api.openai.com"
	DefaultOpenAIModel = "gpt-4o-mini"
	DefaultHTTPTimeout = 30 * time.Second
)

var ErrUnknownAdapter = errors.New("unknown inference adapter")

// Adapter abstracts an inference backend.
type Adapter interface {
	Infer(query string) (response string, err error)
	ModelID() string
	IsReady() bool
}

// Options carries adapter configuration values that can be loaded from YAML.
type Options struct {
	Kind         string       `yaml:"adapter"`
	OllamaURL    string       `yaml:"ollama_url"`
	OllamaModel  string       `yaml:"ollama_model"`
	OpenAIURL    string       `yaml:"openai_url"`
	OpenAIModel  string       `yaml:"openai_model"`
	OpenAIAPIKey string       `yaml:"-"`
	HTTPClient   *http.Client `yaml:"-"`
}

// New constructs an adapter for the requested backend.
func New(kind string, options Options) (Adapter, error) {
	selected := strings.ToLower(strings.TrimSpace(kind))
	if selected == "" {
		selected = strings.ToLower(strings.TrimSpace(options.Kind))
	}

	switch selected {
	case KindEcho:
		return NewEchoAdapter(), nil
	case KindOllama:
		return NewOllamaAdapter(options.OllamaURL, options.OllamaModel, options.HTTPClient), nil
	case KindOpenAI:
		return NewOpenAIAdapter(options.OpenAIURL, options.OpenAIModel, options.OpenAIAPIKey, options.HTTPClient), nil
	default:
		return nil, ErrUnknownAdapter
	}
}

func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: DefaultHTTPTimeout}
}

func normalizeURL(raw string, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = fallback
	}

	return strings.TrimRight(value, "/")
}

func normalizeModel(raw string, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback
	}

	return value
}

func openAIAPIKey(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed != "" {
		return trimmed
	}

	return strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
}
