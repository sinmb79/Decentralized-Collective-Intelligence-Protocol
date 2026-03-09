package identity

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewIdentity(t *testing.T) {
	identity, err := NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity() error = %v", err)
	}

	if len(identity.PrivKey) != 64 {
		t.Fatalf("PrivKey length = %d, want 64", len(identity.PrivKey))
	}
	if len(identity.PubKey) != 32 {
		t.Fatalf("PubKey length = %d, want 32", len(identity.PubKey))
	}
	if !strings.HasPrefix(identity.Address, "DCIP") {
		t.Fatalf("Address prefix = %q, want DCIP", identity.Address)
	}
	if identity.Address != AddressFromPubKey(identity.PubKey) {
		t.Fatalf("AddressFromPubKey() mismatch")
	}
}

func TestSaveLoadIdentityRoundTrip(t *testing.T) {
	identity, err := NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity() error = %v", err)
	}

	path := filepath.Join(t.TempDir(), "nested", "identity.key")
	if err := identity.SaveIdentity(path); err != nil {
		t.Fatalf("SaveIdentity() error = %v", err)
	}

	loaded, err := LoadIdentity(path)
	if err != nil {
		t.Fatalf("LoadIdentity() error = %v", err)
	}

	if !bytes.Equal(loaded.PrivKey, identity.PrivKey) {
		t.Fatalf("loaded PrivKey mismatch")
	}
	if !bytes.Equal(loaded.PubKey, identity.PubKey) {
		t.Fatalf("loaded PubKey mismatch")
	}
	if loaded.Address != identity.Address {
		t.Fatalf("loaded Address = %q, want %q", loaded.Address, identity.Address)
	}
}

func TestLoadIdentityRejectsMismatchedAddress(t *testing.T) {
	identity, err := NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity() error = %v", err)
	}

	payload, err := json.Marshal(&Identity{
		PrivKey: identity.PrivKey,
		PubKey:  identity.PubKey,
		Address: "DCIPbadaddress",
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	path := filepath.Join(t.TempDir(), "identity.key")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := LoadIdentity(path); err == nil {
		t.Fatalf("LoadIdentity() error = nil, want mismatch error")
	}
}

func TestAddressFromPubKeyRejectsInvalidPubKey(t *testing.T) {
	if got := AddressFromPubKey([]byte("short")); got != "" {
		t.Fatalf("AddressFromPubKey() = %q, want empty string", got)
	}
}

func TestSignAndVerify(t *testing.T) {
	identity, err := NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity() error = %v", err)
	}

	message := []byte("hello identity")
	sig := Sign(identity.PrivKey, message)
	if len(sig) != 64 {
		t.Fatalf("signature length = %d, want 64", len(sig))
	}
	if !Verify(identity.PubKey, message, sig) {
		t.Fatalf("Verify() = false, want true")
	}
	if Verify(identity.PubKey, []byte("changed"), sig) {
		t.Fatalf("Verify() = true for changed message, want false")
	}

	otherIdentity, err := NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity() error = %v", err)
	}
	if Verify(otherIdentity.PubKey, message, sig) {
		t.Fatalf("Verify() = true for other public key, want false")
	}
}

func TestSignRejectsInvalidPrivateKey(t *testing.T) {
	if sig := Sign([]byte("short"), []byte("message")); sig != nil {
		t.Fatalf("Sign() = %v, want nil", sig)
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	resolved, err := expandPath("~/dcip/identity.key")
	if err != nil {
		t.Fatalf("expandPath() error = %v", err)
	}

	want := filepath.Join(home, "dcip", "identity.key")
	if resolved != want {
		t.Fatalf("expandPath() = %q, want %q", resolved, want)
	}
}
