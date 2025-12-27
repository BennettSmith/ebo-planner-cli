package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestViewOf_IgnoresNonScalarAndNonMappingShapes(t *testing.T) {
	// Build a doc with odd shapes: currentProfile as mapping, profiles as scalar.
	doc := NewEmptyDocument()
	root := doc.Root.Content[0]
	mapSetNode(root, "currentProfile", &yaml.Node{Kind: yaml.MappingNode})
	mapSetNode(root, "profiles", &yaml.Node{Kind: yaml.ScalarNode, Value: "nope"})

	v, err := ViewOf(doc)
	if err != nil {
		t.Fatalf("view: %v", err)
	}
	if v.CurrentProfile != "" {
		t.Fatalf("currentProfile: got %q", v.CurrentProfile)
	}
	if len(v.Profiles) != 0 {
		t.Fatalf("profiles: expected empty")
	}
}
