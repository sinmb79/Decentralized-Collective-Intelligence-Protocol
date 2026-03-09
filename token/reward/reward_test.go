package reward

import (
	"testing"

	"github.com/dcip/dcip/token/acl"
)

func TestBlockRewardHalvesOnSchedule(t *testing.T) {
	if acl.InitialReward != 84*100_000_000 {
		t.Fatalf("InitialReward = %d, want %d", acl.InitialReward, uint64(84*100_000_000))
	}
	if reward := BlockReward(0); reward != 84*100_000_000 {
		t.Fatalf("BlockReward(0) = %d, want %d", reward, uint64(84*100_000_000))
	}
	if reward := BlockReward(acl.HalvingInterval); reward != acl.InitialReward/2 {
		t.Fatalf("BlockReward(halving) = %d, want %d", reward, acl.InitialReward/2)
	}
}

func TestDistributeBlockRewardUsesLedger(t *testing.T) {
	ledger := &acl.Ledger{}
	manager := NewManager(ledger)

	distribution := manager.DistributeBlockReward(0, "proposer", []string{"p1", "p2"})
	if distribution.Total != acl.InitialReward {
		t.Fatalf("distribution.Total = %d, want %d", distribution.Total, acl.InitialReward)
	}
	if ledger.Balance("proposer") == 0 {
		t.Fatal("expected proposer to receive a reward")
	}
	if ledger.Balance("p1") == 0 || ledger.Balance("p2") == 0 {
		t.Fatal("expected participants to receive a reward")
	}
}

func TestInferenceRewardCapsAtInitialReward(t *testing.T) {
	if reward := InferenceReward(100); reward != acl.InitialReward {
		t.Fatalf("InferenceReward() = %d, want %d", reward, acl.InitialReward)
	}
}

func TestDistributeReservedRewardBurnsRemainder(t *testing.T) {
	ledger := &acl.Ledger{}
	ledger.Restore(acl.Snapshot{
		RewardPool: 101,
		Balances: map[string]uint64{
			"proposer": 0,
			"p1":       0,
			"p2":       0,
			"p3":       0,
		},
	})

	manager := NewManager(ledger)
	distribution := manager.DistributeReservedReward(101, "proposer", []string{"p1", "p2", "p3"})

	if distribution.Proposer != 60 {
		t.Fatalf("distribution.Proposer = %d, want 60", distribution.Proposer)
	}
	if distribution.PerParticipant != 13 {
		t.Fatalf("distribution.PerParticipant = %d, want 13", distribution.PerParticipant)
	}
	if distribution.BurnedRemainder != 2 {
		t.Fatalf("distribution.BurnedRemainder = %d, want 2", distribution.BurnedRemainder)
	}
	if ledger.Snapshot().Burned != 2 {
		t.Fatalf("ledger.Burned = %d, want 2", ledger.Snapshot().Burned)
	}
}
