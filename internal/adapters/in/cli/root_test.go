package cli

import (
	"bytes"
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

func TestGlobalOptions_EnvWhenFlagNotSet(t *testing.T) {
	env := cliopts.MapEnv{
		"EBO_PROFILE": "env-profile",
		"EBO_OUTPUT":  "json",
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

	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if got.Options.Profile != "env-profile" {
		t.Fatalf("Profile: got %q", got.Options.Profile)
	}
	if got.Options.Output != cliopts.OutputJSON {
		t.Fatalf("Output: got %q", got.Options.Output)
	}
	if got.Sources["profile"] != "env" {
		t.Fatalf("source(profile): got %q", got.Sources["profile"])
	}
	if got.Sources["output"] != "env" {
		t.Fatalf("source(output): got %q", got.Sources["output"])
	}
}

func TestGlobalOptions_DefaultsWhenUnset(t *testing.T) {
	env := cliopts.MapEnv{}

	var got cliopts.Resolved
	cmd := NewRootCmd(RootDeps{
		Env:    env,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		OnResolved: func(r cliopts.Resolved) {
			got = r
		},
	})

	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if got.Options.Profile != "default" {
		t.Fatalf("Profile: got %q", got.Options.Profile)
	}
	if got.Options.Output != cliopts.OutputTable {
		t.Fatalf("Output: got %q", got.Options.Output)
	}
	if got.Sources["profile"] != "default" {
		t.Fatalf("source(profile): got %q", got.Sources["profile"])
	}
	if got.Sources["output"] != "default" {
		t.Fatalf("source(output): got %q", got.Sources["output"])
	}
}

func TestGlobalOptions_InvalidOutput(t *testing.T) {
	env := cliopts.MapEnv{"EBO_OUTPUT": "nope"}

	cmd := NewRootCmd(RootDeps{Env: env, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error")
	}
}
