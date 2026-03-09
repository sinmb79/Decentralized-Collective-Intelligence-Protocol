package state

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/dcip/dcip/core/block"
	"github.com/dcip/dcip/token/acl"
)

func TestApplyTransactionUpdatesBalancesAndNonce(t *testing.T) {
	state := New()
	state.Restore(blocklessSnapshot(100))

	tx := block.Transaction{
		From:   []byte("alice"),
		To:     []byte("bob"),
		Amount: 40,
		Fee:    10,
		Nonce:  1,
	}

	if err := state.ApplyTransaction(tx); err != nil {
		t.Fatalf("ApplyTransaction() error = %v", err)
	}
	if state.Balance("alice") != 50 {
		t.Fatalf("alice balance = %d, want 50", state.Balance("alice"))
	}
	if state.Balance("bob") != 40 {
		t.Fatalf("bob balance = %d, want 40", state.Balance("bob"))
	}
	if state.Nonce("alice") != 1 {
		t.Fatalf("alice nonce = %d, want 1", state.Nonce("alice"))
	}
}

func TestApplyBlockRollsBackOnFailure(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	state := New()
	state.Restore(blocklessSnapshot(20))

	blk := &block.Block{
		Header: block.BlockHeader{
			Version:   block.CurrentBlockVersion,
			Height:    1,
			PrevHash:  make([]byte, block.HashSize),
			Timestamp: 1,
			Proposer:  append([]byte(nil), pubKey...),
			VRFProof:  []byte("proof"),
		},
		Txs: []block.Transaction{
			{From: []byte("alice"), To: []byte("bob"), Amount: 10, Fee: 1, Nonce: 2},
		},
	}
	blk.Header.MerkleRoot = blk.MerkleRoot()
	blk.Sig = ed25519.Sign(privKey, blk.Hash())

	if err := state.ApplyBlock(blk); err == nil {
		t.Fatal("expected ApplyBlock() to fail on invalid nonce")
	}
	if state.Balance("alice") != 20 {
		t.Fatalf("alice balance changed after rollback: %d", state.Balance("alice"))
	}
	if state.Nonce("alice") != 0 {
		t.Fatalf("alice nonce = %d, want 0", state.Nonce("alice"))
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	state := New()
	state.Restore(Snapshot{
		Ledger: blocklessSnapshot(25).Ledger,
		Nonces: map[string]uint64{"alice": 3},
	})

	data, err := state.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	restored := New()
	if err := restored.Decode(data); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if restored.Balance("alice") != 25 {
		t.Fatalf("alice balance = %d, want 25", restored.Balance("alice"))
	}
	if restored.Nonce("alice") != 3 {
		t.Fatalf("alice nonce = %d, want 3", restored.Nonce("alice"))
	}
}

func blocklessSnapshot(balance uint64) Snapshot {
	return Snapshot{
		Ledger: acl.Snapshot{
			Balances: map[string]uint64{"alice": balance},
		},
		Nonces: map[string]uint64{},
	}
}
