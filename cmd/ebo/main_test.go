package main

import (
	"strings"
	"testing"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/adapters/out/plannerapi"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/cliopts"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
)

func TestBuildErrorEnvelope_IncludesRequestIDWhenPresent(t *testing.T) {
	peek := cliopts.GlobalOptions{APIURL: "http://api", Profile: "default"}
	err := exitcode.New(exitcode.KindAuth, "unauthorized", &plannerapi.APIError{RequestID: "req-1"})
	env := buildErrorEnvelope(peek, err)
	if env.Meta.RequestID != "req-1" {
		t.Fatalf("requestId: got %q", env.Meta.RequestID)
	}
}

func TestFormatHumanError_NoColor_DisablesANSI(t *testing.T) {
	err := exitcode.New(exitcode.KindUsage, "bad\nTry:\n  ebo x", nil)
	out := formatHumanError(cliopts.GlobalOptions{NoColor: true}, err)
	if strings.Contains(out, "\x1b") {
		t.Fatalf("unexpected ansi: %q", out)
	}
	if !strings.HasPrefix(out, "ERROR: ") {
		t.Fatalf("expected prefix, got %q", out)
	}
}

func TestFormatHumanError_ColorEnabled_UsesANSI(t *testing.T) {
	err := exitcode.New(exitcode.KindUsage, "bad", nil)
	out := formatHumanError(cliopts.GlobalOptions{NoColor: false}, err)
	if !strings.Contains(out, "\x1b[31m") {
		t.Fatalf("expected ansi red, got %q", out)
	}
}
