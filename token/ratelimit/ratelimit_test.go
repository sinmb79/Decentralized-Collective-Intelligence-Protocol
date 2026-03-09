package ratelimit

import (
	"testing"
	"time"

	"github.com/dcip/dcip/token/acl"
)

func TestCheckWithinLimitHasNoFee(t *testing.T) {
	limiter := &RateLimiter{
		now: func() time.Time {
			return time.Unix(1_700_000_000, 0)
		},
	}

	allowed, fee := limiter.Check("alice")
	if !allowed {
		t.Fatalf("allowed = false, want true")
	}
	if fee != 0 {
		t.Fatalf("fee = %d, want 0", fee)
	}
}

func TestCheckChargesFeeAfterLimit(t *testing.T) {
	limiter := &RateLimiter{
		now: func() time.Time {
			return time.Unix(1_700_000_000, 0)
		},
	}

	var fee uint64
	for i := 0; i < acl.RateLimit+1; i++ {
		_, fee = limiter.Check("alice")
	}

	if fee != acl.RateFee {
		t.Fatalf("fee = %d, want %d", fee, acl.RateFee)
	}
}

func TestCheckResetsOnNewWindow(t *testing.T) {
	current := time.Unix(1_700_000_000, 0)
	limiter := &RateLimiter{
		now: func() time.Time {
			return current
		},
	}

	for i := 0; i < acl.RateLimit+1; i++ {
		limiter.Check("alice")
	}

	current = current.Add(time.Hour)
	_, fee := limiter.Check("alice")
	if fee != 0 {
		t.Fatalf("fee = %d, want 0 after new window", fee)
	}
}
