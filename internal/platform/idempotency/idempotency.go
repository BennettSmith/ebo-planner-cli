package idempotency

import "github.com/google/uuid"

// NewKey returns a new idempotency key suitable for Idempotency-Key headers.
func NewKey() string {
	return uuid.NewString()
}
