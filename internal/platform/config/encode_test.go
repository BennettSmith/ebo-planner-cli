package config

import (
	"testing"
)

func TestToInterface(t *testing.T) {
	doc := NewEmptyDocument()
	var err error
	doc, err = SetString(doc, "currentProfile", "default")
	if err != nil {
		t.Fatalf("set: %v", err)
	}

	v, err := ToInterface(doc)
	if err != nil {
		t.Fatalf("to interface: %v", err)
	}
	m, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", v)
	}
	if m["currentProfile"] != "default" {
		t.Fatalf("got %#v", m["currentProfile"])
	}
}
