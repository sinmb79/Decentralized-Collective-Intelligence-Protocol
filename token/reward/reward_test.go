package reward

import (
	"testing"

	"github.com/dcip/dcip/token/acl"
)

func TestBlockRewardHalvesOnSchedule(t *testing.T) {
	if reward := BlockReward(0); reward != acl.InitialReward {
		t.Fatalf("BlockReward(0) = %d, want %d", reward, acl.InitialReward)
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
