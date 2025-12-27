package config

import "gopkg.in/yaml.v3"

// Document is a YAML-backed configuration document that preserves unknown fields.
//
// The canonical on-disk format is YAML.
//
// This type is used as the value passed through the ConfigStore port.
// Keeping the raw yaml.Node allows round-tripping unknown keys.
//
// This is intentionally minimal and will grow as issues #15/#16 add CLI commands.
//
// See docs/cli-spec.md "Config and profiles".
type Document struct {
	Root *yaml.Node
}

type View struct {
	CurrentProfile string
	Profiles       map[string]ProfileView
}

type ProfileView struct {
	APIURL string
}
