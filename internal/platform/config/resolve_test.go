package config

import (
	"testing"
	"time"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/cliopts"
	"github.com/spf13/pflag"
)

func resolvedFromArgs(t *testing.T, args []string) cliopts.Resolved {
	t.Helper()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	defaults := cliopts.DefaultGlobalOptions()
	cliopts.AddGlobalFlags(fs, defaults)
	if err := fs.Parse(args); err != nil {
		t.Fatalf("parse: %v", err)
	}
	r, err := cliopts.ResolveGlobalOptions(fs, cliopts.MapEnv{}, defaults)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	return r
}

func TestResolveEffective_UsesConfigCurrentProfileWhenProfileNotSpecified(t *testing.T) {
	cli := resolvedFromArgs(t, []string{})

	cfg := View{
		CurrentProfile: "staging",
		Profiles: map[string]ProfileView{
			"staging": {APIURL: "https://staging"},
		},
	}

	e := ResolveEffective(cli, cfg)
	if e.Profile != "staging" {
		t.Fatalf("profile: got %q", e.Profile)
	}
	if e.APIURL != "https://staging" {
		t.Fatalf("apiUrl: got %q", e.APIURL)
	}
}

func TestResolveEffective_ProfileFlagOverridesConfig(t *testing.T) {
	cli := resolvedFromArgs(t, []string{"--profile", "dev"})
	cfg := View{CurrentProfile: "staging", Profiles: map[string]ProfileView{"staging": {APIURL: "https://staging"}, "dev": {APIURL: "https://dev"}}}

	e := ResolveEffective(cli, cfg)
	if e.Profile != "dev" {
		t.Fatalf("profile: got %q", e.Profile)
	}
	if e.APIURL != "https://dev" {
		t.Fatalf("apiUrl: got %q", e.APIURL)
	}
}

func TestResolveEffective_APIURLFlagOverridesProfileAPIURL(t *testing.T) {
	cli := resolvedFromArgs(t, []string{"--api-url", "http://override"})
	cfg := View{CurrentProfile: "staging", Profiles: map[string]ProfileView{"staging": {APIURL: "https://staging"}}}

	e := ResolveEffective(cli, cfg)
	if e.APIURL != "http://override" {
		t.Fatalf("apiUrl: got %q", e.APIURL)
	}
}

func TestResolveEffective_ProfileDefaultIfNoConfigCurrentProfile(t *testing.T) {
	cli := resolvedFromArgs(t, []string{})
	cfg := View{CurrentProfile: "", Profiles: map[string]ProfileView{"default": {APIURL: "https://d"}}}

	e := ResolveEffective(cli, cfg)
	if e.Profile != "default" {
		t.Fatalf("profile: got %q", e.Profile)
	}
	if e.APIURL != "https://d" {
		t.Fatalf("apiUrl: got %q", e.APIURL)
	}
}

func TestResolveEffective_IgnoresOtherClioptsFields(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	defaults := cliopts.DefaultGlobalOptions()
	cliopts.AddGlobalFlags(fs, defaults)
	_ = fs.Parse([]string{"--timeout", "1s"})
	r, err := cliopts.ResolveGlobalOptions(fs, cliopts.MapEnv{}, defaults)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if r.Options.Timeout != 1*time.Second {
		t.Fatalf("timeout: got %s", r.Options.Timeout)
	}
	_ = ResolveEffective(r, View{})
}
