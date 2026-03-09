package burn

import (
	"sync"
	"time"

	"github.com/dcip/dcip/token/acl"
)

// Event records an explicit token burn.
type Event struct {
	Address   string
	Amount    uint64
	Reason    string
	Timestamp int64
}

// Tracker records burn events on top of the ACL ledger.
type Tracker struct {
	Ledger *acl.Ledger
	Events []Event
	mutex  sync.RWMutex
	now    func() time.Time
}

// NewTracker creates a burn tracker for a ledger.
func NewTracker(ledger *acl.Ledger) *Tracker {
	if ledger == nil {
		ledger = &acl.Ledger{}
	}

	return &Tracker{
		Ledger: ledger,
		now:    time.Now,
	}
}

// Record burns tokens from an address and appends an audit event.
func (t *Tracker) Record(addr string, amount uint64, reason string) (Event, error) {
	if t == nil || t.Ledger == nil {
		return Event{}, acl.ErrEmptyAddress
	}

	if err := t.Ledger.Burn(addr, amount); err != nil {
		return Event{}, err
	}

	event := Event{
		Address:   addr,
		Amount:    amount,
		Reason:    reason,
		Timestamp: t.now().Unix(),
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.Events = append(t.Events, event)
	return event, nil
}

// TotalBurned returns the burned amount tracked by the underlying ledger.
func (t *Tracker) TotalBurned() uint64 {
	if t == nil || t.Ledger == nil {
		return 0
	}

	return t.Ledger.Snapshot().Burned
}

// EventsSince returns burn events at or after the provided Unix timestamp.
func (t *Tracker) EventsSince(ts int64) []Event {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	filtered := make([]Event, 0, len(t.Events))
	for _, event := range t.Events {
		if event.Timestamp >= ts {
			filtered = append(filtered, event)
		}
	}

	return filtered
}
