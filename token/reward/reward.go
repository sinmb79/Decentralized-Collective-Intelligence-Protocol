package reward

import "github.com/dcip/dcip/token/acl"

// Distribution describes how a reserved reward was split.
type Distribution struct {
	Total            uint64
	Proposer         uint64
	PerParticipant   uint64
	ParticipantCount int
	BurnedRemainder  uint64
}

// Manager coordinates reward minting and distribution on top of the ACL ledger.
type Manager struct {
	Ledger *acl.Ledger
}

// NewManager creates a reward manager bound to a ledger.
func NewManager(ledger *acl.Ledger) *Manager {
	if ledger == nil {
		ledger = &acl.Ledger{}
	}

	return &Manager{Ledger: ledger}
}

// HalvingRound returns the halving epoch for a block height.
func HalvingRound(height uint64) uint64 {
	return height / acl.HalvingInterval
}

// BlockReward returns the nominal block reward before total-supply caps are applied.
func BlockReward(height uint64) uint64 {
	round := HalvingRound(height)
	if round >= 64 {
		return 0
	}

	return acl.InitialReward >> round
}

// InferenceReward returns a simple Phase 1 reward based on the number of signers.
func InferenceReward(signers int) uint64 {
	if signers <= 0 {
		return 0
	}

	reward := (acl.InitialReward / 10) * uint64(signers)
	if reward > acl.InitialReward {
		return acl.InitialReward
	}

	return reward
}

// ReserveBlockReward mints and reserves the block reward in the ledger pool.
func (m *Manager) ReserveBlockReward(height uint64) uint64 {
	if m == nil || m.Ledger == nil {
		return 0
	}

	return m.Ledger.MintReward(height)
}

// DistributeReservedReward splits an already-reserved reward across participants.
func (m *Manager) DistributeReservedReward(reward uint64, proposer string, participants []string) Distribution {
	distribution := Distribution{
		Total:            reward,
		ParticipantCount: len(participants),
	}
	if reward == 0 || m == nil || m.Ledger == nil {
		return distribution
	}

	if len(participants) == 0 {
		distribution.Proposer = reward
		m.Ledger.DistributeReward(reward, proposer, nil)
		return distribution
	}

	distribution.Proposer = reward * 60 / 100
	participantPool := reward * 40 / 100
	distribution.PerParticipant = participantPool / uint64(len(participants))
	distribution.BurnedRemainder = reward - distribution.Proposer - (distribution.PerParticipant * uint64(len(participants)))
	m.Ledger.DistributeReward(reward, proposer, participants)
	return distribution
}

// DistributeBlockReward mints and distributes a reward for the given height.
func (m *Manager) DistributeBlockReward(height uint64, proposer string, participants []string) Distribution {
	reward := m.ReserveBlockReward(height)
	return m.DistributeReservedReward(reward, proposer, participants)
}
