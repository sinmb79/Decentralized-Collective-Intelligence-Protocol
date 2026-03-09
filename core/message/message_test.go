package message

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"reflect"
	"testing"
)

func TestMessageSignAndVerify(t *testing.T) {
	pubKey, privKey := newTestKeypair(t)

	msg := &Message{
		Version: 1,
		Type:    MsgQuery,
		Payload: []byte("hello"),
		TS:      123456789,
	}

	if err := msg.Sign(privKey); err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if !bytes.Equal(msg.From, pubKey) {
		t.Fatalf("Sign() did not populate From with the signer public key")
	}

	if !msg.Verify() {
		t.Fatalf("Verify() = false, want true")
	}
}

func TestMessageVerifyFailsAfterMutation(t *testing.T) {
	_, privKey := newTestKeypair(t)

	base := &Message{
		Version: 1,
		Type:    MsgResponse,
		Payload: []byte("payload"),
		TS:      55,
	}
	if err := base.Sign(privKey); err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	payloadTampered := *base
	payloadTampered.Payload = []byte("changed")
	if payloadTampered.Verify() {
		t.Fatalf("Verify() = true after payload mutation, want false")
	}

	tsTampered := *base
	tsTampered.TS++
	if tsTampered.Verify() {
		t.Fatalf("Verify() = true after timestamp mutation, want false")
	}
}

func TestMessageSignRejectsInvalidPrivateKeyLength(t *testing.T) {
	msg := &Message{}

	if err := msg.Sign([]byte("short")); err == nil {
		t.Fatalf("Sign() error = nil, want invalid private key length error")
	}
}

func TestMessageSignRejectsMismatchedFrom(t *testing.T) {
	_, privKey := newTestKeypair(t)
	otherPubKey, _ := newTestKeypair(t)

	msg := &Message{
		Version: 1,
		Type:    MsgPing,
		From:    otherPubKey,
		TS:      1,
	}

	if err := msg.Sign(privKey); err == nil {
		t.Fatalf("Sign() error = nil, want mismatch error")
	}
}

func TestMessageEncodeDecodeRoundTrip(t *testing.T) {
	pubKey, privKey := newTestKeypair(t)
	targetPubKey, _ := newTestKeypair(t)

	original := &Message{
		Version: 1,
		Type:    MsgProof,
		From:    pubKey,
		To:      targetPubKey,
		Payload: []byte("proof"),
		TS:      999,
	}
	if err := original.Sign(privKey); err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	encoded, err := original.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	var decoded Message
	if err := decoded.Decode(encoded); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if !reflect.DeepEqual(decoded, *original) {
		t.Fatalf("decoded message mismatch\n got: %#v\nwant: %#v", decoded, *original)
	}

	if !decoded.Verify() {
		t.Fatalf("decoded Verify() = false, want true")
	}
}

func TestMessageBroadcastRoundTripKeepsNilTo(t *testing.T) {
	_, privKey := newTestKeypair(t)

	original := &Message{
		Version: 1,
		Type:    MsgBlock,
		Payload: []byte("broadcast"),
		TS:      100,
	}
	if err := original.Sign(privKey); err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	encoded, err := original.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	var decoded Message
	if err := decoded.Decode(encoded); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if decoded.To != nil {
		t.Fatalf("decoded To = %v, want nil", decoded.To)
	}
}

func TestMessageHashIgnoresSignatureAndTracksUnsignedChanges(t *testing.T) {
	msg := &Message{
		Version: 1,
		Type:    MsgTx,
		From:    []byte("from"),
		To:      []byte("to"),
		Payload: []byte("payload"),
		TS:      22,
	}

	hashWithoutSig := msg.Hash()
	msg.Sig = []byte("different-signature")
	hashWithSig := msg.Hash()
	if !bytes.Equal(hashWithoutSig, hashWithSig) {
		t.Fatalf("Hash() changed when only the signature changed")
	}

	msg.Payload = []byte("changed")
	hashWithPayloadChange := msg.Hash()
	if bytes.Equal(hashWithSig, hashWithPayloadChange) {
		t.Fatalf("Hash() did not change after payload mutation")
	}
}

func newTestKeypair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	return pubKey, privKey
}
