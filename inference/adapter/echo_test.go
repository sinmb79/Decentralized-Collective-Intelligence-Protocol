package adapter

import "testing"

func TestEchoAdapterInfer(t *testing.T) {
	backend := NewEchoAdapter()
	response, err := backend.Infer("hello DCIP")
	if err != nil {
		t.Fatalf("Infer returned error: %v", err)
	}
	if response != "hello DCIP" {
		t.Fatalf("unexpected response: %q", response)
	}
	if backend.ModelID() != KindEcho {
		t.Fatalf("unexpected model ID: %q", backend.ModelID())
	}
	if !backend.IsReady() {
		t.Fatal("expected echo adapter to be ready")
	}
}
