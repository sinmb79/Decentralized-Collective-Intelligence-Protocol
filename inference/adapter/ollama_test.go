package adapter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOllamaAdapterInfer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/generate" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if payload["model"] != "llama3" {
			t.Fatalf("unexpected model: %#v", payload["model"])
		}
		if payload["prompt"] != "who are you?" {
			t.Fatalf("unexpected prompt: %#v", payload["prompt"])
		}
		if payload["stream"] != false {
			t.Fatalf("expected stream=false, got %#v", payload["stream"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":"I am DCIP.","done":true}`))
	}))
	defer server.Close()

	backend := NewOllamaAdapter(server.URL, "llama3", server.Client())
	response, err := backend.Infer("who are you?")
	if err != nil {
		t.Fatalf("Infer returned error: %v", err)
	}
	if response != "I am DCIP." {
		t.Fatalf("unexpected response: %q", response)
	}
	if backend.ModelID() != "llama3" {
		t.Fatalf("unexpected model ID: %q", backend.ModelID())
	}
	if !backend.IsReady() {
		t.Fatal("expected ollama adapter to be ready")
	}
}

func TestOllamaAdapterReturnsStatusErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model unavailable", http.StatusBadGateway)
	}))
	defer server.Close()

	backend := NewOllamaAdapter(server.URL, "llama3", server.Client())
	_, err := backend.Infer("ping")
	if err == nil {
		t.Fatal("expected Infer to return an error")
	}
	if !strings.Contains(err.Error(), "502") {
		t.Fatalf("expected status code in error, got %v", err)
	}
}
