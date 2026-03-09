package chain

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/dcip/dcip/core/block"
	"github.com/dcip/dcip/core/state"
	"github.com/dcip/dcip/token/acl"
)

func TestOpenCreatesGenesisBlock(t *testing.T) {
	path := t.TempDir() + "\\chain.db"
	chain, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer chain.Close()

	head := chain.Head()
	if head == nil {
		t.Fatal("expected genesis head")
	}
	if head.Header.Height != 0 {
		t.Fatalf("genesis height = %d, want 0", head.Header.Height)
	}
}

func TestAddBlockPersistsAcrossReopen(t *testing.T) {
	path := t.TempDir() + "\\chain.db"
	chain, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	stateCopy := chain.State()
	stateCopy.Restore(state.Snapshot{
		Ledger: acl.Snapshot{
			Balances: map[string]uint64{"alice": 100},
		},
		Nonces: map[string]uint64{},
	})
	chain.state = stateCopy

	blk, err := signedBlock(chain.Head(), block.Transaction{
		From:   []byte("alice"),
		To:     []byte("bob"),
		Amount: 30,
		Fee:    2,
		Nonce:  1,
	})
	if err != nil {
		t.Fatalf("signedBlock() error = %v", err)
	}

	if err := chain.AddBlock(blk); err != nil {
		t.Fatalf("AddBlock() error = %v", err)
	}
	if err := chain.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("Open(reopen) error = %v", err)
	}
	defer reopened.Close()

	head := reopened.Head()
	if head.Header.Height != 1 {
		t.Fatalf("head height = %d, want 1", head.Header.Height)
	}
	if !bytes.Equal(head.Header.PrevHash, block.GenesisBlock().Hash()) {
		t.Fatal("expected persisted previous hash to reference genesis")
	}
	if reopened.State().Balance("bob") != 30 {
		t.Fatalf("bob balance = %d, want 30", reopened.State().Balance("bob"))
	}
}

func TestAddBlockRejectsWrongPrevHash(t *testing.T) {
	path := t.TempDir() + "\\chain.db"
	chain, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer chain.Close()

	blk := block.GenesisBlock()
	blk.Header.Height = 1
	blk.Header.PrevHash = []byte("wrong")

	if err := chain.AddBlock(blk); err == nil {
		t.Fatal("expected AddBlock() to reject invalid block")
	}
}

func signedBlock(prev *block.Block, tx block.Transaction) (*block.Block, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	blk := &block.Block{
		Header: block.BlockHeader{
			Version:   block.CurrentBlockVersion,
			Height:    prev.Header.Height + 1,
			PrevHash:  prev.Hash(),
			Timestamp: prev.Header.Timestamp + 1,
			Proposer:  append([]byte(nil), pubKey...),
			VRFProof:  []byte("proof"),
		},
		Txs: []block.Transaction{tx},
	}
	blk.Header.MerkleRoot = blk.MerkleRoot()
	blk.Sig = ed25519.Sign(privKey, blk.Hash())
	return blk, nil
}
