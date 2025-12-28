package main

import (
	"testing"

	"github.com/BennettSmith/ebo-planner-cli/internal/adapters/out/plannerapi"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/cliopts"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
)

func TestBuildErrorEnvelope_IncludesRequestIDWhenPresent(t *testing.T) {
	peek := cliopts.GlobalOptions{APIURL: "http://api", Profile: "default"}
	err := exitcode.New(exitcode.KindAuth, "unauthorized", &plannerapi.APIError{RequestID: "req-1"})
	env := buildErrorEnvelope(peek, err)
	if env.Meta.RequestID != "req-1" {
		t.Fatalf("requestId: got %q", env.Meta.RequestID)
	}
}
