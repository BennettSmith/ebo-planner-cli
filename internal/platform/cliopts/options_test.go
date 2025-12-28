package cliopts

import (
	"testing"
	"time"

	"github.com/spf13/pflag"
)

func TestResolveGlobalOptions_Defaults(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	defaults := DefaultGlobalOptions()
	AddGlobalFlags(fs, defaults)
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("parse: %v", err)
	}

	r, err := ResolveGlobalOptions(fs, MapEnv{}, defaults)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if r.Options.Profile != "default" {
		t.Fatalf("profile: got %q", r.Options.Profile)
	}
	if r.Options.Output != OutputTable {
		t.Fatalf("output: got %q", r.Options.Output)
	}
	if r.Options.Timeout != 30*time.Second {
		t.Fatalf("timeout: got %s", r.Options.Timeout)
	}
}

func TestResolveGlobalOptions_EnvTakesEffect(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	defaults := DefaultGlobalOptions()
	AddGlobalFlags(fs, defaults)
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("parse: %v", err)
	}

	env := MapEnv{
		"EBO_OUTPUT":   "json",
		"EBO_TIMEOUT":  "2m",
		"EBO_VERBOSE":  "true",
		"EBO_NO_COLOR": "1",
	}

	r, err := ResolveGlobalOptions(fs, env, defaults)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if r.Options.Output != OutputJSON {
		t.Fatalf("output: got %q", r.Options.Output)
	}
	if r.Options.Timeout != 2*time.Minute {
		t.Fatalf("timeout: got %s", r.Options.Timeout)
	}
	if !r.Options.Verbose {
		t.Fatalf("verbose: got %v", r.Options.Verbose)
	}
	if !r.Options.NoColor {
		t.Fatalf("noColor: got %v", r.Options.NoColor)
	}
}

func TestParseTruthy(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"", false},
		{"  ", false},
		{"1", true},
		{"true", true},
		{"FALSE", false},
	} {
		got, err := parseTruthy(tc.in)
		if err != nil {
			t.Fatalf("parseTruthy(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("parseTruthy(%q): got %v want %v", tc.in, got, tc.want)
		}
	}

	if _, err := parseTruthy("nope"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestResolveGlobalOptions_FlagsOverrideEnv(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	defaults := DefaultGlobalOptions()
	AddGlobalFlags(fs, defaults)
	if err := fs.Parse([]string{"--output", "table", "--timeout", "1s", "--verbose"}); err != nil {
		t.Fatalf("parse: %v", err)
	}

	env := MapEnv{
		"EBO_OUTPUT":  "json",
		"EBO_TIMEOUT": "2m",
		"EBO_VERBOSE": "0",
	}

	r, err := ResolveGlobalOptions(fs, env, defaults)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if r.Options.Output != OutputTable {
		t.Fatalf("output: got %q", r.Options.Output)
	}
	if r.Options.Timeout != 1*time.Second {
		t.Fatalf("timeout: got %s", r.Options.Timeout)
	}
	if !r.Options.Verbose {
		t.Fatalf("verbose: got %v", r.Options.Verbose)
	}
	if r.Sources["output"] != "flag" || r.Sources["timeout"] != "flag" || r.Sources["verbose"] != "flag" {
		t.Fatalf("sources: got %#v", r.Sources)
	}
}

func TestResolveGlobalOptions_InvalidEnvValues(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	defaults := DefaultGlobalOptions()
	AddGlobalFlags(fs, defaults)
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("parse: %v", err)
	}

	for name, env := range map[string]MapEnv{
		"bad output":   {"EBO_OUTPUT": "nope"},
		"bad duration": {"EBO_TIMEOUT": "nope"},
		"bad bool":     {"EBO_VERBOSE": "nope"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ResolveGlobalOptions(fs, env, defaults); err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}

func TestResolveGlobalOptions_Validation(t *testing.T) {
	defaults := DefaultGlobalOptions()

	t.Run("empty profile", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		AddGlobalFlags(fs, defaults)
		if err := fs.Parse([]string{"--profile", ""}); err != nil {
			t.Fatalf("parse: %v", err)
		}
		if _, err := ResolveGlobalOptions(fs, MapEnv{}, defaults); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("negative timeout", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		AddGlobalFlags(fs, defaults)
		if err := fs.Parse([]string{"--timeout", "-1s"}); err != nil {
			t.Fatalf("parse: %v", err)
		}
		if _, err := ResolveGlobalOptions(fs, MapEnv{}, defaults); err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestOSEnv_LookupEnv(t *testing.T) {
	t.Setenv("EBO_TEST_LOOKUP", "yes")
	v, ok := (OSEnv{}).LookupEnv("EBO_TEST_LOOKUP")
	if !ok {
		t.Fatalf("expected ok")
	}
	if v != "yes" {
		t.Fatalf("expected yes, got %q", v)
	}
}
