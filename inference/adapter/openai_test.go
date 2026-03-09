package adapter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIAdapterInfer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization header: %q", got)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if payload["model"] != DefaultOpenAIModel {
			t.Fatalf("unexpected model: %#v", payload["model"])
		}

		messages, ok := payload["messages"].([]any)
		if !ok || len(messages) != 1 {
			t.Fatalf("unexpected messages payload: %#v", payload["messages"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"OpenAI says hi"}}]}`))
	}))
	defer server.Close()

	backend := NewOpenAIAdapter(server.URL, "", "test-key", server.Client())
	response, err := backend.Infer("hello")
	if err != nil {
		t.Fatalf("Infer returned error: %v", err)
	}
	if response != "OpenAI says hi" {
		t.Fatalf("unexpected response: %q", response)
	}
	if backend.ModelID() != DefaultOpenAIModel {
		t.Fatalf("unexpected model ID: %q", backend.ModelID())
	}
}

func TestOpenAIAdapterUsesEnvironmentAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "env-key")

	backend := NewOpenAIAdapter("https://api.openai.com", "gpt-4o-mini", "", &http.Client{})
	if !backend.IsReady() {
		t.Fatal("expected adapter to use OPENAI_API_KEY from the environment")
	}
}

func TestOpenAIAdapterRequiresAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	backend := NewOpenAIAdapter("https://api.openai.com", "gpt-4o-mini", "", &http.Client{})
	if backend.IsReady() {
		t.Fatal("expected adapter without API key to be not ready")
	}
}
