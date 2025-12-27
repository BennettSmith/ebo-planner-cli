package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/BennettSmith/ebo-planner-cli/internal/platform/cliopts"
)

func TestGlobalOptions_PreferenceFlagsOverEnv(t *testing.T) {
	env := cliopts.MapEnv{
		"EBO_API_URL":  "http://env",
		"EBO_PROFILE":  "env-profile",
		"EBO_OUTPUT":   "json",
		"EBO_NO_COLOR": "1",
		"EBO_TIMEOUT":  "10s",
		"EBO_VERBOSE":  "1",
	}

	var got cliopts.Resolved
	cmd := NewRootCmd(RootDeps{
		Env:    env,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		OnResolved: func(r cliopts.Resolved) {
			got = r
		},
	})

	cmd.SetArgs([]string{
		"--api-url", "http://flag",
		"--profile", "flag-profile",
		"--output", "table",
		"--no-color",
		"--timeout", "2m",
		"--verbose",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if got.Options.APIURL != "http://flag" {
		t.Fatalf("APIURL: got %q", got.Options.APIURL)
	}
	if got.Options.Profile != "flag-profile" {
		t.Fatalf("Profile: got %q", got.Options.Profile)
	}
	if got.Options.Output != cliopts.OutputTable {
		t.Fatalf("Output: got %q", got.Options.Output)
	}
	if got.Options.NoColor != true {
		t.Fatalf("NoColor: got %v", got.Options.NoColor)
	}
	if got.Options.Timeout.String() != "2m0s" {
		t.Fatalf("Timeout: got %s", got.Options.Timeout)
	}
	if got.Options.Verbose != true {
		t.Fatalf("Verbose: got %v", got.Options.Verbose)
	}

	for _, k := range []string{"api-url", "profile", "output", "no-color", "timeout", "verbose"} {
		if got.Sources[k] != "flag" {
			t.Fatalf("source(%s): got %q", k, got.Sources[k])
		}
	}
}

func TestRoot_JSONOutputIsValidJSON(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{Env: cliopts.MapEnv{}, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr, got: %q", stderr.String())
	}

	var got any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, stdout.String())
	}
}
