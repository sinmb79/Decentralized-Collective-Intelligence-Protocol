package burn

import (
	"testing"
	"time"

	"github.com/dcip/dcip/token/acl"
)

func TestRecordBurnsTokensAndTracksEvents(t *testing.T) {
	ledger := &acl.Ledger{}
	ledger.Restore(acl.Snapshot{
		Balances: map[string]uint64{"alice": 100},
	})

	tracker := NewTracker(ledger)
	tracker.now = func() time.Time { return time.Unix(123, 0) }

	event, err := tracker.Record("alice", 40, "rate limit")
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if event.Amount != 40 || event.Reason != "rate limit" {
		t.Fatalf("unexpected event: %#v", event)
	}
	if ledger.Balance("alice") != 60 {
		t.Fatalf("alice balance = %d, want 60", ledger.Balance("alice"))
	}
	if tracker.TotalBurned() != 40 {
		t.Fatalf("TotalBurned() = %d, want 40", tracker.TotalBurned())
	}
}

func TestEventsSinceFiltersByTimestamp(t *testing.T) {
	tracker := NewTracker(&acl.Ledger{})
	tracker.Events = []Event{
		{Address: "a", Amount: 1, Timestamp: 10},
		{Address: "b", Amount: 2, Timestamp: 20},
	}

	events := tracker.EventsSince(15)
	if len(events) != 1 || events[0].Address != "b" {
		t.Fatalf("unexpected events: %#v", events)
	}
}
