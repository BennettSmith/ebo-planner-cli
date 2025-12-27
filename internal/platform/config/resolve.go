package config

import "github.com/BennettSmith/ebo-planner-cli/internal/platform/cliopts"

type Effective struct {
	Profile string
	APIURL  string
}

// ResolveEffective combines CLI options (including their source) with config-file
// settings.
//
// Precedence (highest -> lowest):
//  1. CLI flags
//  2. env vars
//  3. config file (currentProfile + profiles.<name>.apiUrl)
//  4. defaults
func ResolveEffective(cli cliopts.Resolved, cfg View) Effective {
	profile := cli.Options.Profile
	if cli.Sources["profile"] == "default" && cfg.CurrentProfile != "" {
		profile = cfg.CurrentProfile
	}

	apiURL := cli.Options.APIURL
	if cli.Sources["api-url"] == "default" {
		if p, ok := cfg.Profiles[profile]; ok && p.APIURL != "" {
			apiURL = p.APIURL
		}
	}

	return Effective{Profile: profile, APIURL: apiURL}
}
