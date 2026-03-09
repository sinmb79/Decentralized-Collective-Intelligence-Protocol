package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/dcip/dcip/core/block"
	"github.com/dcip/dcip/token/acl"
)

var (
	ErrNilBlock           = errors.New("block is nil")
	ErrInvalidTransaction = errors.New("transaction contains empty addresses")
	ErrInvalidNonce       = errors.New("invalid transaction nonce")
	ErrBlockVerification  = errors.New("block verification failed")
)

// Snapshot captures chain state for persistence and rollback.
type Snapshot struct {
	Ledger acl.Snapshot      `json:"ledger"`
	Nonces map[string]uint64 `json:"nonces"`
}

// State tracks balances and nonces derived from applied blocks.
type State struct {
	Ledger *acl.Ledger
	Nonces map[string]uint64
	mutex  sync.RWMutex
}

// New creates an empty state.
func New() *State {
	return &State{
		Ledger: &acl.Ledger{},
		Nonces: make(map[string]uint64),
	}
}

// Balance returns the current ACL balance for an address.
func (s *State) Balance(addr string) uint64 {
	if s == nil || s.Ledger == nil {
		return 0
	}

	return s.Ledger.Balance(addr)
}

// Nonce returns the latest observed nonce for an address.
func (s *State) Nonce(addr string) uint64 {
	if s == nil {
		return 0
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.Nonces == nil {
		return 0
	}

	return s.Nonces[addr]
}

// ApplyTransaction mutates balances and nonces according to a transaction.
func (s *State) ApplyTransaction(tx block.Transaction) error {
	if len(tx.From) == 0 || len(tx.To) == 0 {
		return ErrInvalidTransaction
	}

	from := string(tx.From)
	to := string(tx.To)

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.ensureState()

	expectedNonce := s.Nonces[from] + 1
	if tx.Nonce != expectedNonce {
		return fmt.Errorf("%w: got %d want %d", ErrInvalidNonce, tx.Nonce, expectedNonce)
	}

	if err := s.Ledger.Transfer(from, to, tx.Amount, tx.Fee); err != nil {
		return err
	}

	s.Nonces[from] = tx.Nonce
	return nil
}

// ApplyBlock applies all transactions from a block as a single unit.
func (s *State) ApplyBlock(b *block.Block) error {
	if b == nil {
		return ErrNilBlock
	}
	if !b.Verify() {
		return ErrBlockVerification
	}

	snapshot := s.Snapshot()
	for _, tx := range b.Txs {
		if err := s.ApplyTransaction(tx); err != nil {
			s.Restore(snapshot)
			return err
		}
	}

	return nil
}

// Snapshot returns a deep copy of the current state.
func (s *State) Snapshot() Snapshot {
	if s == nil {
		return Snapshot{}
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	nonces := make(map[string]uint64, len(s.Nonces))
	for addr, nonce := range s.Nonces {
		nonces[addr] = nonce
	}

	ledgerSnapshot := acl.Snapshot{}
	if s.Ledger != nil {
		ledgerSnapshot = s.Ledger.Snapshot()
	}

	return Snapshot{
		Ledger: ledgerSnapshot,
		Nonces: nonces,
	}
}

// Restore replaces the current state with a snapshot.
func (s *State) Restore(snapshot Snapshot) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.ensureState()

	s.Ledger.Restore(snapshot.Ledger)
	s.Nonces = make(map[string]uint64, len(snapshot.Nonces))
	for addr, nonce := range snapshot.Nonces {
		s.Nonces[addr] = nonce
	}
}

// Encode serializes the state snapshot.
func (s *State) Encode() ([]byte, error) {
	return json.Marshal(s.Snapshot())
}

// Decode deserializes and restores the state snapshot.
func (s *State) Decode(data []byte) error {
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}

	s.Restore(snapshot)
	return nil
}

func (s *State) ensureState() {
	if s.Ledger == nil {
		s.Ledger = &acl.Ledger{}
	}
	if s.Nonces == nil {
		s.Nonces = make(map[string]uint64)
	}
}
