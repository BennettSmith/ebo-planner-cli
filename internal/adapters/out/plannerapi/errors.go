package plannerapi

import (
	"fmt"

	gen "github.com/BennettSmith/ebo-planner-cli/internal/gen/plannerapi"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
)

type APIError struct {
	StatusCode int
	ErrorCode  string
	Message    string
	RequestID  string
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.ErrorCode != "" {
		return fmt.Sprintf("%s: %s", e.ErrorCode, e.Message)
	}
	return e.Message
}

func exitKindForStatus(status int) exitcode.Kind {
	switch status {
	case 401:
		return exitcode.KindAuth
	case 404:
		return exitcode.KindNotFound
	case 409:
		return exitcode.KindConflict
	case 422:
		return exitcode.KindValidation
	default:
		if status >= 500 {
			return exitcode.KindServer
		}
		return exitcode.KindUnexpected
	}
}

// apiErrorFromAny converts an oapi-codegen error variant into a single error type.
//
// In our generated client, error variants like Unauthorized/NotFound/etc are type
// aliases of ErrorResponse, so we only need to handle *ErrorResponse here.
func apiErrorFromAny(status int, variants ...any) error {
	for _, v := range variants {
		if v == nil {
			continue
		}
		if er, ok := v.(*gen.ErrorResponse); ok {
			return apiErrorFromErrorResponse(status, er)
		}
	}
	return exitcode.New(exitKindForStatus(status), fmt.Sprintf("http %d", status), nil)
}

func apiErrorFromErrorResponse(status int, er *gen.ErrorResponse) *exitcode.Error {
	if er == nil {
		return exitcode.New(exitKindForStatus(status), fmt.Sprintf("http %d", status), nil)
	}
	requestID := ""
	if er.Error.RequestId != nil {
		requestID = *er.Error.RequestId
	}
	ae := &APIError{StatusCode: status, ErrorCode: er.Error.Code, Message: er.Error.Message, RequestID: requestID}
	return exitcode.New(exitKindForStatus(status), ae.Error(), ae)
}
