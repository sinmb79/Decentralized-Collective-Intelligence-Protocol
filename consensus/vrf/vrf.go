package vrf

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"sort"

	"github.com/dcip/dcip/core/identity"
	"golang.org/x/crypto/sha3"
)

var (
	ErrInvalidPrivateKey = errors.New("invalid VRF private key")
	ErrInvalidPublicKey  = errors.New("invalid VRF public key")
)

// VRF provides a deterministic, verifiable selector over Ed25519 keys.
type VRF struct {
	PrivKey []byte
	PubKey  []byte
}

// New creates a VRF from an Ed25519 private key.
func New(privKey []byte) (*VRF, error) {
	if len(privKey) != ed25519.PrivateKeySize {
		return nil, ErrInvalidPrivateKey
	}

	pubKey := ed25519.PrivateKey(privKey).Public().(ed25519.PublicKey)
	return &VRF{
		PrivKey: cloneBytes(privKey),
		PubKey:  cloneBytes(pubKey),
	}, nil
}

// FromIdentity creates a VRF from a DCIP identity.
func FromIdentity(id *identity.Identity) (*VRF, error) {
	if id == nil {
		return nil, ErrInvalidPrivateKey
	}

	return New(id.PrivKey)
}

// Prove returns a deterministic output and proof for a given input.
// NOTE: Phase 1 approximation using Ed25519 signatures as a VRF substitute.
// Replace with a RFC 9381-compliant ECVRF implementation in Phase 3.
func (v *VRF) Prove(input []byte) ([]byte, []byte, error) {
	if v == nil || len(v.PrivKey) != ed25519.PrivateKeySize || len(v.PubKey) != ed25519.PublicKeySize {
		return nil, nil, ErrInvalidPrivateKey
	}

	proof := ed25519.Sign(ed25519.PrivateKey(v.PrivKey), input)
	sum := sha3.Sum256(append(append(cloneBytes(v.PubKey), input...), proof...))
	return sum[:], cloneBytes(proof), nil
}

// Verify checks that the proof was produced by the given public key and input.
func Verify(pubKey, input, output, proof []byte) bool {
	if len(pubKey) != ed25519.PublicKeySize || len(proof) != ed25519.SignatureSize {
		return false
	}
	if !ed25519.Verify(ed25519.PublicKey(pubKey), input, proof) {
		return false
	}

	sum := sha3.Sum256(append(append(cloneBytes(pubKey), input...), proof...))
	return bytes.Equal(output, sum[:])
}

// SelectNode selects a stable node from the provided set.
func SelectNode(output []byte, nodes []string) string {
	if len(nodes) == 0 {
		return ""
	}

	candidates := append([]string(nil), nodes...)
	sort.Strings(candidates)

	buf := make([]byte, 8)
	if len(output) >= 8 {
		copy(buf, output[:8])
	} else {
		copy(buf[8-len(output):], output)
	}

	index := binary.BigEndian.Uint64(buf) % uint64(len(candidates))
	return candidates[int(index)]
}

func cloneBytes(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}

	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
