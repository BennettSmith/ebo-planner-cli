package idempotency

import "testing"

func TestNewKey_NonEmptyAndDistinct(t *testing.T) {
	k1 := NewKey()
	k2 := NewKey()
	if k1 == "" || k2 == "" {
		t.Fatalf("expected non-empty keys: %q %q", k1, k2)
	}
	if k1 == k2 {
		t.Fatalf("expected distinct keys, got %q", k1)
	}
}
