package exitcode

import (
	"errors"
	"fmt"
)

// Codes are the minimum required exit-code contract from docs/cli-spec.md.
const (
	Success    = 0
	Unexpected = 1
	Usage      = 2
	Auth       = 3
	NotFound   = 4
	Conflict   = 5
	Validation = 6
	Server     = 7
)

type Kind string

const (
	KindUnexpected Kind = "unexpected"
	KindUsage      Kind = "usage"
	KindAuth       Kind = "auth"
	KindNotFound   Kind = "not_found"
	KindConflict   Kind = "conflict"
	KindValidation Kind = "validation"
	KindServer     Kind = "server"
)

type Error struct {
	Kind Kind
	Msg  string
	Err  error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil && e.Msg != "" {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Msg
}

func (e *Error) Unwrap() error { return e.Err }

func New(kind Kind, msg string, err error) *Error {
	return &Error{Kind: kind, Msg: msg, Err: err}
}

func Code(err error) int {
	if err == nil {
		return Success
	}
	var e *Error
	if errors.As(err, &e) {
		switch e.Kind {
		case KindUsage:
			return Usage
		case KindAuth:
			return Auth
		case KindNotFound:
			return NotFound
		case KindConflict:
			return Conflict
		case KindValidation:
			return Validation
		case KindServer:
			return Server
		default:
			return Unexpected
		}
	}
	return Unexpected
}
