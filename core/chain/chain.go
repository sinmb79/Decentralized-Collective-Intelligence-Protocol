package chain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/dcip/dcip/core/block"
	"github.com/dcip/dcip/core/state"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	keyHeadHash      = "chain/head"
	keyStateSnapshot = "chain/state"
)

var (
	ErrNilBlock      = errors.New("block is nil")
	ErrInvalidHeight = errors.New("block height is not sequential")
	ErrPrevHash      = errors.New("previous hash does not match chain head")
	ErrBlockNotFound = errors.New("block not found")
)

// Chain persists blocks and derived state in LevelDB.
type Chain struct {
	db    *leveldb.DB
	path  string
	head  *block.Block
	state *state.State
	mutex sync.RWMutex
}

// Open opens or creates a chain database at the provided path.
func Open(path string) (*Chain, error) {
	if path == "" {
		return nil, errors.New("chain path is empty")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}

	chain := &Chain{
		db:    db,
		path:  path,
		state: state.New(),
	}

	if err := chain.load(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return chain, nil
}

// Close closes the underlying database.
func (c *Chain) Close() error {
	if c == nil || c.db == nil {
		return nil
	}

	return c.db.Close()
}

// Head returns the current chain head.
func (c *Chain) Head() *block.Block {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return cloneBlock(c.head)
}

// Height returns the height of the current chain head.
func (c *Chain) Height() uint64 {
	head := c.Head()
	if head == nil {
		return 0
	}

	return head.Header.Height
}

// State returns the current derived state.
func (c *Chain) State() *state.State {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	cloned := state.New()
	cloned.Restore(c.state.Snapshot())
	return cloned
}

// BlockByHeight returns the block stored at a given height.
func (c *Chain) BlockByHeight(height uint64) (*block.Block, error) {
	hash, err := c.db.Get(heightKey(height), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}

	return c.BlockByHash(hash)
}

// BlockByHash returns the block stored under a given hash.
func (c *Chain) BlockByHash(hash []byte) (*block.Block, error) {
	data, err := c.db.Get(blockKey(hash), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}

	var stored block.Block
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, err
	}

	return &stored, nil
}

// AddBlock validates, persists, and applies a new block.
func (c *Chain) AddBlock(b *block.Block) error {
	if b == nil {
		return ErrNilBlock
	}
	if !b.Verify() {
		return state.ErrBlockVerification
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	headHash := c.head.Hash()
	if b.Header.Height != c.head.Header.Height+1 {
		return fmt.Errorf("%w: got %d want %d", ErrInvalidHeight, b.Header.Height, c.head.Header.Height+1)
	}
	if !bytes.Equal(b.Header.PrevHash, headHash) {
		return ErrPrevHash
	}

	nextState := state.New()
	nextState.Restore(c.state.Snapshot())
	if err := nextState.ApplyBlock(b); err != nil {
		return err
	}

	blockData, err := json.Marshal(b)
	if err != nil {
		return err
	}
	stateData, err := nextState.Encode()
	if err != nil {
		return err
	}

	hash := b.Hash()
	batch := new(leveldb.Batch)
	batch.Put(blockKey(hash), blockData)
	batch.Put(heightKey(b.Header.Height), hash)
	batch.Put([]byte(keyHeadHash), hash)
	batch.Put([]byte(keyStateSnapshot), stateData)
	if err := c.db.Write(batch, nil); err != nil {
		return err
	}

	c.head = cloneBlock(b)
	c.state = nextState
	return nil
}

func (c *Chain) load() error {
	headHash, err := c.db.Get([]byte(keyHeadHash), nil)
	if errors.Is(err, leveldb.ErrNotFound) {
		genesis := block.GenesisBlock()
		data, marshalErr := json.Marshal(genesis)
		if marshalErr != nil {
			return marshalErr
		}

		stateData, encodeErr := c.state.Encode()
		if encodeErr != nil {
			return encodeErr
		}

		hash := genesis.Hash()
		batch := new(leveldb.Batch)
		batch.Put(blockKey(hash), data)
		batch.Put(heightKey(0), hash)
		batch.Put([]byte(keyHeadHash), hash)
		batch.Put([]byte(keyStateSnapshot), stateData)
		if writeErr := c.db.Write(batch, nil); writeErr != nil {
			return writeErr
		}

		c.head = genesis
		return nil
	}
	if err != nil {
		return err
	}

	head, err := c.BlockByHash(headHash)
	if err != nil {
		return err
	}

	stateData, err := c.db.Get([]byte(keyStateSnapshot), nil)
	if err == nil {
		if decodeErr := c.state.Decode(stateData); decodeErr != nil {
			return decodeErr
		}
	}

	c.head = head
	return nil
}

func blockKey(hash []byte) []byte {
	return []byte("block/" + fmt.Sprintf("%x", hash))
}

func heightKey(height uint64) []byte {
	return []byte(fmt.Sprintf("height/%020d", height))
}

func cloneBlock(src *block.Block) *block.Block {
	if src == nil {
		return nil
	}

	data, err := json.Marshal(src)
	if err != nil {
		return nil
	}

	var cloned block.Block
	if err := json.Unmarshal(data, &cloned); err != nil {
		return nil
	}

	return &cloned
}
