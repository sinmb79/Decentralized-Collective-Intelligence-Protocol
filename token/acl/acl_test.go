package acl

import "testing"

func TestBalanceReturnsZeroForUnknownAddress(t *testing.T) {
	var ledger Ledger

	if got := ledger.Balance("missing"); got != 0 {
		t.Fatalf("Balance() = %d, want 0", got)
	}
}

func TestTransferSplitsFeeBetweenBurnAndRewardPool(t *testing.T) {
	ledger := &Ledger{
		Balances: map[string]uint64{
			"alice": 1000,
			"bob":   50,
		},
	}

	if err := ledger.Transfer("alice", "bob", 200, 5); err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	if got := ledger.Balance("alice"); got != 795 {
		t.Fatalf("alice balance = %d, want 795", got)
	}
	if got := ledger.Balance("bob"); got != 250 {
		t.Fatalf("bob balance = %d, want 250", got)
	}
	if ledger.Burned != 2 {
		t.Fatalf("Burned = %d, want 2", ledger.Burned)
	}
	if ledger.rewardPool != 3 {
		t.Fatalf("rewardPool = %d, want 3", ledger.rewardPool)
	}
}

func TestTransferRejectsInsufficientBalance(t *testing.T) {
	ledger := &Ledger{
		Balances: map[string]uint64{
			"alice": 100,
		},
	}

	if err := ledger.Transfer("alice", "bob", 100, 1); err != ErrInsufficientBalance {
		t.Fatalf("Transfer() error = %v, want %v", err, ErrInsufficientBalance)
	}
}

func TestMintRewardAppliesHalving(t *testing.T) {
	var ledger Ledger

	first := ledger.MintReward(0)
	second := ledger.MintReward(HalvingInterval)

	if first != InitialReward {
		t.Fatalf("MintReward(0) = %d, want %d", first, InitialReward)
	}
	if second != InitialReward/2 {
		t.Fatalf("MintReward(halving) = %d, want %d", second, InitialReward/2)
	}
	if ledger.Minted != first+second {
		t.Fatalf("Minted = %d, want %d", ledger.Minted, first+second)
	}
	if ledger.rewardPool != first+second {
		t.Fatalf("rewardPool = %d, want %d", ledger.rewardPool, first+second)
	}
}

func TestMintRewardCapsAtTotalSupply(t *testing.T) {
	ledger := &Ledger{
		Minted:     TotalSupply - 10,
		rewardPool: 5,
	}

	reward := ledger.MintReward(0)
	if reward != 10 {
		t.Fatalf("MintReward() = %d, want 10", reward)
	}
	if ledger.Minted != TotalSupply {
		t.Fatalf("Minted = %d, want %d", ledger.Minted, TotalSupply)
	}
	if ledger.rewardPool != 15 {
		t.Fatalf("rewardPool = %d, want 15", ledger.rewardPool)
	}
}

func TestDistributeRewardSplitsProposerAndParticipants(t *testing.T) {
	ledger := &Ledger{
		Balances: map[string]uint64{
			"proposer": 0,
			"p1":       0,
			"p2":       0,
		},
		rewardPool: 100,
	}

	ledger.DistributeReward(100, "proposer", []string{"p1", "p2"})

	if ledger.Balance("proposer") != 60 {
		t.Fatalf("proposer balance = %d, want 60", ledger.Balance("proposer"))
	}
	if ledger.Balance("p1") != 20 || ledger.Balance("p2") != 20 {
		t.Fatalf("participant balances = %d/%d, want 20/20", ledger.Balance("p1"), ledger.Balance("p2"))
	}
	if ledger.rewardPool != 0 {
		t.Fatalf("rewardPool = %d, want 0", ledger.rewardPool)
	}
}

func TestBurnReducesBalanceAndTracksBurned(t *testing.T) {
	ledger := &Ledger{
		Balances: map[string]uint64{
			"alice": 100,
		},
	}

	if err := ledger.Burn("alice", 40); err != nil {
		t.Fatalf("Burn() error = %v", err)
	}

	if ledger.Balance("alice") != 60 {
		t.Fatalf("alice balance = %d, want 60", ledger.Balance("alice"))
	}
	if ledger.Burned != 40 {
		t.Fatalf("Burned = %d, want 40", ledger.Burned)
	}
}
