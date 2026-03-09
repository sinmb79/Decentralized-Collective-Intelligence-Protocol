package block

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func TestGenesisBlock(t *testing.T) {
	genesis := GenesisBlock()

	if genesis.Header.Version != CurrentBlockVersion {
		t.Fatalf("Version = %d, want %d", genesis.Header.Version, CurrentBlockVersion)
	}
	if genesis.Header.Height != 0 {
		t.Fatalf("Height = %d, want 0", genesis.Header.Height)
	}
	if genesis.Header.Timestamp != 0 {
		t.Fatalf("Timestamp = %d, want 0", genesis.Header.Timestamp)
	}
	if genesis.Inscription != GenesisInscription {
		t.Fatalf("Inscription mismatch")
	}
	if len(genesis.Txs) != 0 {
		t.Fatalf("Genesis transactions = %d, want 0", len(genesis.Txs))
	}
	if !genesis.Verify() {
		t.Fatalf("Verify() = false for genesis block")
	}
}

func TestMerkleRootDeterministic(t *testing.T) {
	block := &Block{
		Txs: []Transaction{
			{From: []byte("alice"), To: []byte("bob"), Amount: 10, Fee: 1, Nonce: 1, Sig: []byte("sig1")},
			{From: []byte("carol"), To: []byte("dave"), Amount: 20, Fee: 2, Nonce: 2, Sig: []byte("sig2")},
		},
	}

	rootA := block.MerkleRoot()
	rootB := block.MerkleRoot()
	if !bytes.Equal(rootA, rootB) {
		t.Fatalf("MerkleRoot() mismatch across repeated calls")
	}
	if len(rootA) != 32 {
		t.Fatalf("MerkleRoot length = %d, want 32", len(rootA))
	}
}

func TestBlockHashChangesWithHeaderMutation(t *testing.T) {
	blockA := &Block{
		Header: BlockHeader{
			Version:    CurrentBlockVersion,
			Height:     1,
			PrevHash:   bytes.Repeat([]byte{0x01}, 32),
			MerkleRoot: bytes.Repeat([]byte{0x02}, 32),
			Timestamp:  123,
			Proposer:   bytes.Repeat([]byte{0x03}, 32),
			VRFProof:   []byte("proof"),
		},
	}
	blockB := &Block{
		Header: blockA.Header,
	}
	blockB.Header.Height = 2

	if bytes.Equal(blockA.Hash(), blockB.Hash()) {
		t.Fatalf("Hash() did not change after header mutation")
	}
}

func TestVerifySignedBlock(t *testing.T) {
	pubKey, privKey := newTestKeypair(t)

	block := &Block{
		Header: BlockHeader{
			Version:   CurrentBlockVersion,
			Height:    1,
			PrevHash:  bytes.Repeat([]byte{0x01}, 32),
			Timestamp: 12345,
			Proposer:  pubKey,
			VRFProof:  []byte("vrf"),
		},
		Txs: []Transaction{
			{From: []byte("alice"), To: []byte("bob"), Amount: 42, Fee: 1, Nonce: 7, Sig: []byte("txsig")},
		},
	}
	block.Header.MerkleRoot = block.MerkleRoot()
	block.Sig = ed25519.Sign(privKey, block.Hash())

	if !block.Verify() {
		t.Fatalf("Verify() = false, want true")
	}
}

func TestVerifyRejectsTamperedTransactions(t *testing.T) {
	pubKey, privKey := newTestKeypair(t)

	block := &Block{
		Header: BlockHeader{
			Version:   CurrentBlockVersion,
			Height:    2,
			PrevHash:  bytes.Repeat([]byte{0x09}, 32),
			Timestamp: 99,
			Proposer:  pubKey,
		},
		Txs: []Transaction{
			{From: []byte("alice"), To: []byte("bob"), Amount: 1, Fee: 1, Nonce: 1, Sig: []byte("sig")},
		},
	}
	block.Header.MerkleRoot = block.MerkleRoot()
	block.Sig = ed25519.Sign(privKey, block.Hash())

	block.Txs[0].Amount = 100
	if block.Verify() {
		t.Fatalf("Verify() = true after transaction mutation, want false")
	}
}

func TestVerifyRejectsBadSignature(t *testing.T) {
	pubKey, privKey := newTestKeypair(t)

	block := &Block{
		Header: BlockHeader{
			Version:   CurrentBlockVersion,
			Height:    3,
			PrevHash:  bytes.Repeat([]byte{0x07}, 32),
			Timestamp: 77,
			Proposer:  pubKey,
		},
		Txs: []Transaction{
			{From: []byte("alice"), To: []byte("bob"), Amount: 5, Fee: 1, Nonce: 1, Sig: []byte("sig")},
		},
	}
	block.Header.MerkleRoot = block.MerkleRoot()
	block.Sig = ed25519.Sign(privKey, block.Hash())
	block.Sig[0] ^= 0xff

	if block.Verify() {
		t.Fatalf("Verify() = true with a bad signature, want false")
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
