package adapter

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewReturnsEchoAdapter(t *testing.T) {
	backend, err := New(KindEcho, Options{})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if !backend.IsReady() {
		t.Fatal("expected echo adapter to be ready")
	}
	if backend.ModelID() != KindEcho {
		t.Fatalf("unexpected model ID: %q", backend.ModelID())
	}
}

func TestNewUsesOptionsKind(t *testing.T) {
	backend, err := New("", Options{
		Kind:       KindOllama,
		HTTPClient: &http.Client{},
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if backend.ModelID() != DefaultOllamaModel {
		t.Fatalf("unexpected default model: %q", backend.ModelID())
	}
}

func TestNewRejectsUnknownAdapter(t *testing.T) {
	_, err := New("unknown", Options{})
	if !errors.Is(err, ErrUnknownAdapter) {
		t.Fatalf("expected ErrUnknownAdapter, got %v", err)
	}
}
