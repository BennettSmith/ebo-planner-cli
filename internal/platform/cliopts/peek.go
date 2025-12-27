package cliopts

import (
	"strings"
	"time"
)

// PeekGlobalOptions best-effort resolves global options from args/env/defaults.
//
// This is intentionally simpler than ResolveGlobalOptions: it exists so the
// entrypoint can decide whether to emit JSON output even when Cobra flag parsing
// fails (e.g., unknown flag).
func PeekGlobalOptions(args []string, env EnvProvider, defaults GlobalOptions) GlobalOptions {
	opts := defaults
	outputSet := false
	profileSet := false
	apiURLSet := false
	noColorSet := false
	timeoutSet := false
	verboseSet := false

	// flags first
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--api-url" && i+1 < len(args):
			opts.APIURL = args[i+1]
			apiURLSet = true
			i++
		case strings.HasPrefix(a, "--api-url="):
			opts.APIURL = strings.TrimPrefix(a, "--api-url=")
			apiURLSet = true
		case a == "--profile" && i+1 < len(args):
			opts.Profile = args[i+1]
			profileSet = true
			i++
		case strings.HasPrefix(a, "--profile="):
			opts.Profile = strings.TrimPrefix(a, "--profile=")
			profileSet = true
		case a == "--output" && i+1 < len(args):
			opts.Output = OutputFormat(strings.ToLower(args[i+1]))
			outputSet = true
			i++
		case strings.HasPrefix(a, "--output="):
			opts.Output = OutputFormat(strings.ToLower(strings.TrimPrefix(a, "--output=")))
			outputSet = true
		case a == "--no-color":
			opts.NoColor = true
			noColorSet = true
		case a == "--timeout" && i+1 < len(args):
			if d, err := time.ParseDuration(args[i+1]); err == nil {
				opts.Timeout = d
				timeoutSet = true
			}
			i++
		case strings.HasPrefix(a, "--timeout="):
			if d, err := time.ParseDuration(strings.TrimPrefix(a, "--timeout=")); err == nil {
				opts.Timeout = d
				timeoutSet = true
			}
		case a == "--verbose":
			opts.Verbose = true
			verboseSet = true
		}
	}

	// then env
	if env != nil {
		if !apiURLSet {
			if v, ok := env.LookupEnv("EBO_API_URL"); ok {
				opts.APIURL = v
			}
		}
		if !profileSet {
			if v, ok := env.LookupEnv("EBO_PROFILE"); ok {
				opts.Profile = v
			}
		}
		if !outputSet {
			if v, ok := env.LookupEnv("EBO_OUTPUT"); ok {
				opts.Output = OutputFormat(strings.ToLower(v))
			}
		}
		if !noColorSet {
			if v, ok := env.LookupEnv("EBO_NO_COLOR"); ok {
				if b, err := parseTruthy(v); err == nil {
					opts.NoColor = b
				}
			}
		}
		if !timeoutSet {
			if v, ok := env.LookupEnv("EBO_TIMEOUT"); ok {
				if d, err := time.ParseDuration(v); err == nil {
					opts.Timeout = d
				}
			}
		}
		if !verboseSet {
			if v, ok := env.LookupEnv("EBO_VERBOSE"); ok {
				if b, err := parseTruthy(v); err == nil {
					opts.Verbose = b
				}
			}
		}
	}

	return opts
}
