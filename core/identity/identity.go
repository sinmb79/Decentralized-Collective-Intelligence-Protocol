package identity

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/sha3"
)

const (
	RoleAgent     = "agent"
	RoleValidator = "validator"
	RoleHuman     = "human"
	RoleRelay     = "relay"
)

// Identity stores a node keypair and its derived DCIP address.
type Identity struct {
	PrivKey []byte `json:"priv_key"`
	PubKey  []byte `json:"pub_key"`
	Address string `json:"address"`
}

// NewIdentity creates a new Ed25519 identity.
func NewIdentity() (*Identity, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &Identity{
		PrivKey: cloneBytes(privKey),
		PubKey:  cloneBytes(pubKey),
		Address: AddressFromPubKey(pubKey),
	}, nil
}

// LoadIdentity loads an identity from disk.
func LoadIdentity(path string) (*Identity, error) {
	resolvedPath, err := expandPath(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, err
	}

	var identity Identity
	if err := json.Unmarshal(data, &identity); err != nil {
		return nil, err
	}

	if err := identity.normalize(); err != nil {
		return nil, err
	}

	return &identity, nil
}

// SaveIdentity saves an identity to disk.
func (i *Identity) SaveIdentity(path string) error {
	if err := i.normalize(); err != nil {
		return err
	}

	resolvedPath, err := expandPath(path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(resolvedPath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(resolvedPath, data, 0o600)
}

// AddressFromPubKey derives a DCIP address from an Ed25519 public key.
func AddressFromPubKey(pubKey []byte) string {
	if len(pubKey) != ed25519.PublicKeySize {
		return ""
	}

	sum := sha3.Sum256(pubKey)
	return "DCIP" + base58.CheckEncode(sum[:], 0x00)
}

// Sign signs a message with an Ed25519 private key.
func Sign(privKey, message []byte) []byte {
	if len(privKey) != ed25519.PrivateKeySize {
		return nil
	}

	return cloneBytes(ed25519.Sign(ed25519.PrivateKey(privKey), message))
}

// Verify checks an Ed25519 signature.
func Verify(pubKey, message, sig []byte) bool {
	if len(pubKey) != ed25519.PublicKeySize || len(sig) != ed25519.SignatureSize {
		return false
	}

	return ed25519.Verify(ed25519.PublicKey(pubKey), message, sig)
}

func (i *Identity) normalize() error {
	if len(i.PrivKey) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid Ed25519 private key length: %d", len(i.PrivKey))
	}

	derivedPubKey := ed25519.PrivateKey(i.PrivKey).Public().(ed25519.PublicKey)
	if len(i.PubKey) == 0 {
		i.PubKey = cloneBytes(derivedPubKey)
	} else {
		if len(i.PubKey) != ed25519.PublicKeySize {
			return fmt.Errorf("invalid Ed25519 public key length: %d", len(i.PubKey))
		}
		if !bytes.Equal(i.PubKey, derivedPubKey) {
			return errors.New("identity public key does not match private key")
		}
	}

	derivedAddress := AddressFromPubKey(i.PubKey)
	if derivedAddress == "" {
		return errors.New("failed to derive identity address")
	}
	if i.Address == "" {
		i.Address = derivedAddress
	} else if i.Address != derivedAddress {
		return errors.New("identity address does not match public key")
	}

	return nil
}

func expandPath(path string) (string, error) {
	if path == "" {
		return "", errors.New("identity path is empty")
	}

	if path == "~" {
		return os.UserHomeDir()
	}

	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}

	return path, nil
}

func cloneBytes(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}

	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
