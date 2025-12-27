package plannerapi

import (
	"errors"
	"testing"

	gen "github.com/BennettSmith/ebo-planner-cli/internal/gen/plannerapi"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
)

func TestExitKindForStatus(t *testing.T) {
	cases := []struct {
		status int
		want   exitcode.Kind
	}{
		{401, exitcode.KindAuth},
		{404, exitcode.KindNotFound},
		{409, exitcode.KindConflict},
		{422, exitcode.KindValidation},
		{500, exitcode.KindServer},
		{418, exitcode.KindUnexpected},
	}
	for _, tc := range cases {
		if got := exitKindForStatus(tc.status); got != tc.want {
			t.Fatalf("status %d: got %q want %q", tc.status, got, tc.want)
		}
	}
}

func TestAPIError_ErrorFormatting(t *testing.T) {
	e := &APIError{ErrorCode: "X", Message: "m"}
	if e.Error() != "X: m" {
		t.Fatalf("got %q", e.Error())
	}
	e2 := &APIError{Message: "m"}
	if e2.Error() != "m" {
		t.Fatalf("got %q", e2.Error())
	}
}

func TestApiErrorFromAny_Fallback(t *testing.T) {
	err := apiErrorFromAny(404, nil)
	if exitcode.Code(err) != exitcode.NotFound {
		t.Fatalf("got %d", exitcode.Code(err))
	}
}

func TestApiErrorFromErrorResponse_WrapsAPIError(t *testing.T) {
	rid := "req-1"
	er := &gen.ErrorResponse{}
	er.Error.Code = "C"
	er.Error.Message = "msg"
	er.Error.RequestId = &rid

	err := apiErrorFromErrorResponse(401, er)
	var ee *exitcode.Error
	if !errors.As(err, &ee) {
		t.Fatalf("expected exitcode.Error")
	}
	var ae *APIError
	if !errors.As(err, &ae) {
		t.Fatalf("expected APIError unwrap")
	}
	if ae.RequestID != "req-1" {
		t.Fatalf("requestId: got %q", ae.RequestID)
	}
}
