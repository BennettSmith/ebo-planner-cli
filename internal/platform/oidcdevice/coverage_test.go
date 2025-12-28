package oidcdevice

import "testing"

func TestIsPendingAndIsSlowDown(t *testing.T) {
	if !(TokenError{Error: "authorization_pending"}).IsPending() {
		t.Fatalf("expected pending")
	}
	if (TokenError{Error: "authorization_pending"}).IsSlowDown() {
		t.Fatalf("expected not slow_down")
	}
	if !(TokenError{Error: "slow_down"}).IsSlowDown() {
		t.Fatalf("expected slow_down")
	}
}
