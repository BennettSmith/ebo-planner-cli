package plannerapi

import (
	"errors"
	"testing"

	gen "github.com/BennettSmith/ebo-planner-cli/internal/gen/plannerapi"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
)

func TestAPIError_Error_NilReceiver(t *testing.T) {
	var e *APIError
	if e.Error() != "" {
		t.Fatalf("expected empty")
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

func TestAPIErrorFromAny_NoVariantUsesStatusOnly(t *testing.T) {
	err := apiErrorFromAny(422, nil)
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Validation {
		t.Fatalf("expected validation exit 6, got %d", exitcode.Code(err))
	}
}

func TestAPIErrorFromErrorResponse_WrapsAPIError(t *testing.T) {
	rid := "req-1"
	er := &gen.ErrorResponse{}
	er.Error.Code = "UNAUTHORIZED"
	er.Error.Message = "nope"
	er.Error.RequestId = &rid

	err := apiErrorFromErrorResponse(401, er)
	var ee *exitcode.Error
	if !errors.As(err, &ee) {
		t.Fatalf("expected exitcode.Error")
	}
	var ae *APIError
	if !errors.As(ee.Err, &ae) {
		t.Fatalf("expected APIError cause, got %T", ee.Err)
	}
	if ae.RequestID != "req-1" {
		t.Fatalf("requestId: got %q", ae.RequestID)
	}
}
