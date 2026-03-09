package block

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"

	"golang.org/x/crypto/sha3"
)

const (
	CurrentBlockVersion = 1
	HashSize            = 32
	GenesisInscription  = "Alone we are limited. Together we are intelligence.\nFor the first time, humans and AI stand on the same network."
)

// BlockHeader contains the fields committed by the block hash.
type BlockHeader struct {
	Version    uint32
	Height     uint64
	PrevHash   []byte
	MerkleRoot []byte
	Timestamp  int64
	Proposer   []byte
	VRFProof   []byte
}

// Block is the unit committed to the chain.
type Block struct {
	Header      BlockHeader
	Proofs      []InferenceProof
	Txs         []Transaction
	Sig         []byte
	Inscription string
}

// InferenceProof records a completed inference proof.
type InferenceProof struct {
	QueryHash   []byte
	ResponseCID string
	Signatures  [][]byte
	Timestamp   int64
}

// Transaction records an ACL token transfer.
type Transaction struct {
	From   []byte
	To     []byte
	Amount uint64
	Fee    uint64
	Nonce  uint64
	Sig    []byte
}

// Hash returns the SHA3-256 hash of the block header.
func (b *Block) Hash() []byte {
	if b == nil {
		return nil
	}

	sum := sha3.Sum256(headerBytes(b.Header))
	return sum[:]
}

// Verify validates block structure, Merkle root, and proposer signature.
func (b *Block) Verify() bool {
	if b == nil || b.Header.Version != CurrentBlockVersion {
		return false
	}

	if !bytes.Equal(b.Header.MerkleRoot, b.MerkleRoot()) {
		return false
	}

	if b.Header.Height == 0 {
		return b.Header.Timestamp == 0 &&
			len(b.Header.PrevHash) == 0 &&
			len(b.Header.Proposer) == 0 &&
			len(b.Header.VRFProof) == 0 &&
			len(b.Sig) == 0 &&
			len(b.Proofs) == 0 &&
			len(b.Txs) == 0 &&
			b.Inscription == GenesisInscription
	}

	if len(b.Header.PrevHash) != HashSize {
		return false
	}
	if len(b.Header.Proposer) != ed25519.PublicKeySize || len(b.Sig) != ed25519.SignatureSize {
		return false
	}

	return ed25519.Verify(ed25519.PublicKey(b.Header.Proposer), b.Hash(), b.Sig)
}

// MerkleRoot calculates the transaction Merkle root.
func (b *Block) MerkleRoot() []byte {
	if b == nil || len(b.Txs) == 0 {
		sum := sha3.Sum256(nil)
		return sum[:]
	}

	level := make([][]byte, 0, len(b.Txs))
	for _, tx := range b.Txs {
		level = append(level, hashTransaction(tx))
	}

	for len(level) > 1 {
		nextLevel := make([][]byte, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			left := level[i]
			right := left
			if i+1 < len(level) {
				right = level[i+1]
			}
			sum := sha3.Sum256(append(cloneBytes(left), right...))
			nextLevel = append(nextLevel, sum[:])
		}
		level = nextLevel
	}

	return cloneBytes(level[0])
}

// GenesisBlock returns the deterministic genesis block definition.
func GenesisBlock() *Block {
	block := &Block{
		Header: BlockHeader{
			Version:   CurrentBlockVersion,
			Height:    0,
			Timestamp: 0,
		},
		Inscription: GenesisInscription,
	}
	block.Header.MerkleRoot = block.MerkleRoot()
	return block
}

func hashTransaction(tx Transaction) []byte {
	var buf bytes.Buffer
	writeBytes(&buf, tx.From)
	writeBytes(&buf, tx.To)
	_ = binary.Write(&buf, binary.LittleEndian, tx.Amount)
	_ = binary.Write(&buf, binary.LittleEndian, tx.Fee)
	_ = binary.Write(&buf, binary.LittleEndian, tx.Nonce)
	writeBytes(&buf, tx.Sig)

	sum := sha3.Sum256(buf.Bytes())
	return sum[:]
}

func headerBytes(header BlockHeader) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, header.Version)
	_ = binary.Write(&buf, binary.LittleEndian, header.Height)
	writeBytes(&buf, header.PrevHash)
	writeBytes(&buf, header.MerkleRoot)
	_ = binary.Write(&buf, binary.LittleEndian, header.Timestamp)
	writeBytes(&buf, header.Proposer)
	writeBytes(&buf, header.VRFProof)
	return buf.Bytes()
}

func writeBytes(buf *bytes.Buffer, data []byte) {
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(data)))
	_, _ = buf.Write(data)
}

func cloneBytes(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}

	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
