package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMapSetScalar_UpdatesExistingKey(t *testing.T) {
	m := &yaml.Node{Kind: yaml.MappingNode}
	mapSetScalar(m, "k", "v1")
	mapSetScalar(m, "k", "v2")
	if got := mapGet(m, "k"); got == nil || got.Value != "v2" {
		t.Fatalf("got %#v", got)
	}
}

func TestMapEnsureMapping_OverwritesNonMapping(t *testing.T) {
	m := &yaml.Node{Kind: yaml.MappingNode}
	mapSetNode(m, "profiles", &yaml.Node{Kind: yaml.ScalarNode, Value: "bad"})
	pm := mapEnsureMapping(m, "profiles")
	if pm.Kind != yaml.MappingNode {
		t.Fatalf("expected mapping")
	}
}
