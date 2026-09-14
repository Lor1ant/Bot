package web

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	rl := newRateLimiter(3, 50*time.Millisecond)
	for i := 0; i < 3; i++ {
		if !rl.allow("ip1") {
			t.Fatalf("attempt %d should be allowed", i)
		}
	}
	if rl.allow("ip1") {
		t.Fatal("4th attempt must be blocked")
	}
	if !rl.allow("ip2") {
		t.Fatal("a different key must be independent")
	}
	time.Sleep(60 * time.Millisecond)
	if !rl.allow("ip1") {
		t.Fatal("must allow again after the window")
	}
}
