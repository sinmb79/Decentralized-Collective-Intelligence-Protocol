package acl

import (
	"errors"
	"math"
	"sync"
)

const (
	TotalSupply     uint64  = 2_100_000_000 * 100_000_000
	InitialReward   uint64  = 84 * 100_000_000
	HalvingInterval uint64  = 12_614_400
	FeesBurnRatio   float64 = 0.5
	RateLimit       int     = 100
	RateFee         uint64  = 10_000
)

var (
	ErrEmptyAddress        = errors.New("address is empty")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrBalanceOverflow     = errors.New("balance overflow")
)

// Ledger tracks ACL balances, minting, and burn accounting.
type Ledger struct {
	Balances   map[string]uint64
	Burned     uint64
	Minted     uint64
	rewardPool uint64
	mutex      sync.RWMutex
}

// Snapshot captures ledger state for persistence or rollback.
type Snapshot struct {
	Balances   map[string]uint64
	Burned     uint64
	Minted     uint64
	RewardPool uint64
}

// Balance returns the current balance for an address.
func (l *Ledger) Balance(addr string) uint64 {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if l.Balances == nil {
		return 0
	}

	return l.Balances[addr]
}

// RewardPool returns the currently reserved reward pool.
func (l *Ledger) RewardPool() uint64 {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.rewardPool
}

// Transfer moves ACL from one address to another and splits the fee.
func (l *Ledger) Transfer(from, to string, amount, fee uint64) error {
	if from == "" || to == "" {
		return ErrEmptyAddress
	}
	if amount > math.MaxUint64-fee {
		return ErrBalanceOverflow
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.ensureState()

	totalDebit := amount + fee
	if l.Balances[from] < totalDebit {
		return ErrInsufficientBalance
	}
	if l.Balances[to] > math.MaxUint64-amount {
		return ErrBalanceOverflow
	}

	burned := fee / 2
	rewardShare := fee - burned
	if l.Burned > math.MaxUint64-burned || l.rewardPool > math.MaxUint64-rewardShare {
		return ErrBalanceOverflow
	}

	l.Balances[from] -= totalDebit
	l.Balances[to] += amount
	l.Burned += burned
	l.rewardPool += rewardShare
	return nil
}

// MintReward calculates and reserves the block reward for distribution.
func (l *Ledger) MintReward(blockHeight uint64) uint64 {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.ensureState()

	halvings := blockHeight / HalvingInterval
	if halvings >= 64 {
		return 0
	}

	reward := InitialReward >> halvings
	if reward == 0 || l.Minted >= TotalSupply {
		return 0
	}

	remaining := TotalSupply - l.Minted
	if reward > remaining {
		reward = remaining
	}

	l.Minted += reward
	l.rewardPool += reward
	return reward
}

// DistributeReward splits a reserved reward between proposer and participants.
func (l *Ledger) DistributeReward(reward uint64, proposer string, participants []string) {
	if reward == 0 {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.ensureState()

	if reward > l.rewardPool {
		reward = l.rewardPool
	}
	if reward == 0 {
		return
	}

	if len(participants) == 0 {
		l.Balances[proposer] += reward
		l.rewardPool -= reward
		return
	}

	proposerShare := reward * 60 / 100
	participantPool := reward * 40 / 100
	perParticipant := participantPool / uint64(len(participants))
	remainder := reward - proposerShare - (perParticipant * uint64(len(participants)))

	l.Balances[proposer] += proposerShare
	for _, participant := range participants {
		l.Balances[participant] += perParticipant
	}

	l.Burned += remainder
	l.rewardPool -= reward
}

// Burn destroys ACL held by an address.
func (l *Ledger) Burn(addr string, amount uint64) error {
	if addr == "" {
		return ErrEmptyAddress
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.ensureState()

	if l.Balances[addr] < amount {
		return ErrInsufficientBalance
	}
	if l.Burned > math.MaxUint64-amount {
		return ErrBalanceOverflow
	}

	l.Balances[addr] -= amount
	l.Burned += amount
	return nil
}

// Snapshot returns a deep copy of the current ledger state.
func (l *Ledger) Snapshot() Snapshot {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	balances := make(map[string]uint64, len(l.Balances))
	for addr, balance := range l.Balances {
		balances[addr] = balance
	}

	return Snapshot{
		Balances:   balances,
		Burned:     l.Burned,
		Minted:     l.Minted,
		RewardPool: l.rewardPool,
	}
}

// Restore replaces the ledger state with a persisted snapshot.
func (l *Ledger) Restore(snapshot Snapshot) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.Balances = make(map[string]uint64, len(snapshot.Balances))
	for addr, balance := range snapshot.Balances {
		l.Balances[addr] = balance
	}
	l.Burned = snapshot.Burned
	l.Minted = snapshot.Minted
	l.rewardPool = snapshot.RewardPool
}

func (l *Ledger) ensureState() {
	if l.Balances == nil {
		l.Balances = make(map[string]uint64)
	}
}
