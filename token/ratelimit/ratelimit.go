package ratelimit

import (
	"sync"
	"time"

	"github.com/dcip/dcip/token/acl"
)

// RateLimiter tracks per-address query windows.
type RateLimiter struct {
	counts  map[string]int
	windows map[string]int64
	mutex   sync.Mutex
	now     func() time.Time
}

// Check records a query attempt and returns the fee to apply for the current window.
func (r *RateLimiter) Check(addr string) (allowed bool, fee uint64) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.ensureState()

	window := r.currentWindow()
	if r.windows[addr] != window {
		r.windows[addr] = window
		r.counts[addr] = 0
	}

	r.counts[addr]++
	if r.counts[addr] > acl.RateLimit {
		return true, acl.RateFee
	}

	return true, 0
}

func (r *RateLimiter) ensureState() {
	if r.counts == nil {
		r.counts = make(map[string]int)
	}
	if r.windows == nil {
		r.windows = make(map[string]int64)
	}
	if r.now == nil {
		r.now = time.Now
	}
}

func (r *RateLimiter) currentWindow() int64 {
	return r.now().UTC().Unix() / int64(time.Hour/time.Second)
}
