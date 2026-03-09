package vrf

import (
	"bytes"
	"testing"

	"github.com/dcip/dcip/core/identity"
)

func TestProveAndVerify(t *testing.T) {
	id, err := identity.NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity() error = %v", err)
	}

	engine, err := New(id.PrivKey)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	output, proof, err := engine.Prove([]byte("query-hash:1"))
	if err != nil {
		t.Fatalf("Prove() error = %v", err)
	}
	if len(output) == 0 || len(proof) == 0 {
		t.Fatal("expected non-empty VRF output and proof")
	}
	if !Verify(engine.PubKey, []byte("query-hash:1"), output, proof) {
		t.Fatal("expected Verify() to succeed")
	}
}

func TestVerifyRejectsWrongKey(t *testing.T) {
	first, err := identity.NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity(first) error = %v", err)
	}
	second, err := identity.NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity(second) error = %v", err)
	}

	engine, err := New(first.PrivKey)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	output, proof, err := engine.Prove([]byte("query-hash:2"))
	if err != nil {
		t.Fatalf("Prove() error = %v", err)
	}
	if Verify(second.PubKey, []byte("query-hash:2"), output, proof) {
		t.Fatal("expected Verify() to fail with the wrong public key")
	}
}

func TestSelectNodeIsStable(t *testing.T) {
	selected := SelectNode([]byte{0, 0, 0, 0, 0, 0, 0, 2}, []string{"node-c", "node-a", "node-b"})
	if selected != "node-c" {
		t.Fatalf("SelectNode() = %q, want %q", selected, "node-c")
	}

	other := SelectNode([]byte{0, 0, 0, 0, 0, 0, 0, 2}, []string{"node-b", "node-c", "node-a"})
	if !bytes.Equal([]byte(selected), []byte(other)) {
		t.Fatalf("expected sorted selection to be stable: %q vs %q", selected, other)
	}
}
