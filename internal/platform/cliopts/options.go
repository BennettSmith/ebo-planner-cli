package cliopts

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

type OutputFormat string

const (
	OutputTable OutputFormat = "table"
	OutputJSON  OutputFormat = "json"
)

type GlobalOptions struct {
	APIURL  string
	Profile string
	Output  OutputFormat
	NoColor bool
	Timeout time.Duration
	Verbose bool
}

func DefaultGlobalOptions() GlobalOptions {
	return GlobalOptions{
		Profile: "default",
		Output:  OutputTable,
		Timeout: 30 * time.Second,
	}
}

type EnvProvider interface {
	LookupEnv(key string) (string, bool)
}

type OSEnv struct{}

func (OSEnv) LookupEnv(key string) (string, bool) { return os.LookupEnv(key) }

type MapEnv map[string]string

func (m MapEnv) LookupEnv(key string) (string, bool) {
	v, ok := m[key]
	return v, ok
}

func AddGlobalFlags(fs *pflag.FlagSet, defaults GlobalOptions) {
	fs.String("api-url", defaults.APIURL, "Override API base URL (or set EBO_API_URL)")
	fs.String("profile", defaults.Profile, "Select profile (or set EBO_PROFILE)")
	fs.String("output", string(defaults.Output), "Output format: table|json (or set EBO_OUTPUT)")
	fs.Bool("no-color", defaults.NoColor, "Disable ANSI color (or set EBO_NO_COLOR=1)")
	fs.Duration("timeout", defaults.Timeout, "Request timeout (e.g., 10s, 2m) (or set EBO_TIMEOUT)")
	fs.Bool("verbose", defaults.Verbose, "Verbose logging to stderr (or set EBO_VERBOSE=1)")
}

type Resolved struct {
	Options GlobalOptions
	// Sources indicates where each setting was resolved from: "flag", "env", or "default".
	Sources map[string]string
}

func ResolveGlobalOptions(fs *pflag.FlagSet, env EnvProvider, defaults GlobalOptions) (Resolved, error) {
	out := Resolved{Options: defaults, Sources: map[string]string{}}

	getString := func(flagName, envKey string, dst *string) error {
		if f := fs.Lookup(flagName); f != nil && f.Changed {
			v, err := fs.GetString(flagName)
			if err != nil {
				return err
			}
			*dst = v
			out.Sources[flagName] = "flag"
			return nil
		}
		if v, ok := env.LookupEnv(envKey); ok {
			*dst = v
			out.Sources[flagName] = "env"
			return nil
		}
		out.Sources[flagName] = "default"
		return nil
	}

	getBoolOne := func(flagName, envKey string, dst *bool) error {
		if f := fs.Lookup(flagName); f != nil && f.Changed {
			v, err := fs.GetBool(flagName)
			if err != nil {
				return err
			}
			*dst = v
			out.Sources[flagName] = "flag"
			return nil
		}
		if v, ok := env.LookupEnv(envKey); ok {
			b, err := parseTruthy(v)
			if err != nil {
				return fmt.Errorf("%s: %w", envKey, err)
			}
			*dst = b
			out.Sources[flagName] = "env"
			return nil
		}
		out.Sources[flagName] = "default"
		return nil
	}

	getDuration := func(flagName, envKey string, dst *time.Duration) error {
		if f := fs.Lookup(flagName); f != nil && f.Changed {
			v, err := fs.GetDuration(flagName)
			if err != nil {
				return err
			}
			*dst = v
			out.Sources[flagName] = "flag"
			return nil
		}
		if v, ok := env.LookupEnv(envKey); ok {
			d, err := time.ParseDuration(v)
			if err != nil {
				return fmt.Errorf("%s: %w", envKey, err)
			}
			*dst = d
			out.Sources[flagName] = "env"
			return nil
		}
		out.Sources[flagName] = "default"
		return nil
	}

	outputStr := string(defaults.Output)
	if err := getString("api-url", "EBO_API_URL", &out.Options.APIURL); err != nil {
		return Resolved{}, err
	}
	if err := getString("profile", "EBO_PROFILE", &out.Options.Profile); err != nil {
		return Resolved{}, err
	}
	if err := getString("output", "EBO_OUTPUT", &outputStr); err != nil {
		return Resolved{}, err
	}
	if err := getBoolOne("no-color", "EBO_NO_COLOR", &out.Options.NoColor); err != nil {
		return Resolved{}, err
	}
	if err := getDuration("timeout", "EBO_TIMEOUT", &out.Options.Timeout); err != nil {
		return Resolved{}, err
	}
	if err := getBoolOne("verbose", "EBO_VERBOSE", &out.Options.Verbose); err != nil {
		return Resolved{}, err
	}

	out.Options.Output = OutputFormat(strings.ToLower(strings.TrimSpace(outputStr)))
	switch out.Options.Output {
	case OutputTable, OutputJSON:
		// ok
	default:
		return Resolved{}, fmt.Errorf("invalid --output %q (expected table|json)", outputStr)
	}

	if out.Options.Profile == "" {
		return Resolved{}, fmt.Errorf("invalid --profile: must be non-empty")
	}

	if out.Options.Timeout < 0 {
		return Resolved{}, fmt.Errorf("invalid --timeout: must be non-negative")
	}

	return out, nil
}

func parseTruthy(v string) (bool, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return false, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("invalid boolean value %q", v)
	}
	return b, nil
}
