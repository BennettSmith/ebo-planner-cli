package exitcode

import (
	"errors"
	"testing"
)

func TestCode(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, Success},
		{"usage", New(KindUsage, "bad args", nil), Usage},
		{"auth", New(KindAuth, "no token", nil), Auth},
		{"notfound", New(KindNotFound, "missing", nil), NotFound},
		{"conflict", New(KindConflict, "dup", nil), Conflict},
		{"validation", New(KindValidation, "invalid", nil), Validation},
		{"server", New(KindServer, "down", nil), Server},
		{"unexpected", New(KindUnexpected, "boom", nil), Unexpected},
		{"plain error", errors.New("x"), Unexpected},
		{"unknown kind", New("weird", "x", nil), Unexpected},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Code(tc.err); got != tc.want {
				t.Fatalf("got %d want %d", got, tc.want)
			}
		})
	}
}

func TestErrorFormattingAndUnwrap(t *testing.T) {
	inner := errors.New("inner")

	e := New(KindUsage, "bad", inner)
	if e.Error() != "bad: inner" {
		t.Fatalf("error: got %q", e.Error())
	}
	if !errors.Is(e, inner) {
		t.Fatalf("expected unwrap")
	}

	e2 := New(KindUsage, "", inner)
	if e2.Error() != "inner" {
		t.Fatalf("error: got %q", e2.Error())
	}

	e3 := New(KindUsage, "msg", nil)
	if e3.Error() != "msg" {
		t.Fatalf("error: got %q", e3.Error())
	}
}
