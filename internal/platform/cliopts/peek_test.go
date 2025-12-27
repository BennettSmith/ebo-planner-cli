package cliopts

import "testing"

func TestPeekGlobalOptions_FlagOverridesEnv(t *testing.T) {
	defaults := DefaultGlobalOptions()
	env := MapEnv{"EBO_OUTPUT": "json", "EBO_PROFILE": "env"}

	opts := PeekGlobalOptions([]string{"--output", "table", "--profile", "flag"}, env, defaults)
	if opts.Output != OutputTable {
		t.Fatalf("output: got %q", opts.Output)
	}
	if opts.Profile != "flag" {
		t.Fatalf("profile: got %q", opts.Profile)
	}
}

func TestPeekGlobalOptions_FlagEqualsForm(t *testing.T) {
	defaults := DefaultGlobalOptions()
	opts := PeekGlobalOptions([]string{"--output=json", "--profile=p"}, MapEnv{}, defaults)
	if opts.Output != OutputJSON {
		t.Fatalf("output: got %q", opts.Output)
	}
	if opts.Profile != "p" {
		t.Fatalf("profile: got %q", opts.Profile)
	}
}

func TestPeekGlobalOptions_EnvWhenNoFlags(t *testing.T) {
	defaults := DefaultGlobalOptions()
	env := MapEnv{"EBO_OUTPUT": "json", "EBO_NO_COLOR": "1", "EBO_VERBOSE": "true"}

	opts := PeekGlobalOptions([]string{}, env, defaults)
	if opts.Output != OutputJSON {
		t.Fatalf("output: got %q", opts.Output)
	}
	if !opts.NoColor {
		t.Fatalf("noColor: got %v", opts.NoColor)
	}
	if !opts.Verbose {
		t.Fatalf("verbose: got %v", opts.Verbose)
	}
}

func TestPeekGlobalOptions_EnvInvalidDoesNotCrash(t *testing.T) {
	defaults := DefaultGlobalOptions()
	env := MapEnv{"EBO_TIMEOUT": "nope", "EBO_NO_COLOR": "nope"}

	opts := PeekGlobalOptions([]string{}, env, defaults)
	// stays at defaults
	if opts.Timeout != defaults.Timeout {
		t.Fatalf("timeout: got %s", opts.Timeout)
	}
	if opts.NoColor != defaults.NoColor {
		t.Fatalf("noColor: got %v", opts.NoColor)
	}
}

func TestPeekGlobalOptions_TimeoutFlagAndEqualsForm(t *testing.T) {
	defaults := DefaultGlobalOptions()

	opts := PeekGlobalOptions([]string{"--timeout", "2m", "--timeout=1s"}, MapEnv{}, defaults)
	// last one wins in our simple scan
	if opts.Timeout.String() != "1s" {
		t.Fatalf("timeout: got %s", opts.Timeout)
	}
}

func TestPeekGlobalOptions_VerboseEnvFalse(t *testing.T) {
	defaults := DefaultGlobalOptions()
	env := MapEnv{"EBO_VERBOSE": "false"}

	opts := PeekGlobalOptions([]string{}, env, defaults)
	if opts.Verbose {
		t.Fatalf("verbose: got %v", opts.Verbose)
	}
}
