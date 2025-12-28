package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDotPath_SetGetUnset(t *testing.T) {
	doc := NewEmptyDocument()
	var err error

	doc, err = SetString(doc, "profiles.dev.apiUrl", "http://x")
	if err != nil {
		t.Fatalf("set: %v", err)
	}
	got, err := Get(doc, "profiles.dev.apiUrl")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "http://x" {
		t.Fatalf("got %q", got)
	}

	doc, err = Unset(doc, "profiles.dev.apiUrl")
	if err != nil {
		t.Fatalf("unset: %v", err)
	}
	if _, err := Get(doc, "profiles.dev.apiUrl"); err == nil {
		t.Fatalf("expected not found")
	}
}

func TestGet_NotFound(t *testing.T) {
	doc := NewEmptyDocument()
	if _, err := Get(doc, "nope"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestGet_NonScalar(t *testing.T) {
	doc := NewEmptyDocument()
	root := doc.Root.Content[0]
	mapSetNode(root, "profiles", &yaml.Node{Kind: yaml.MappingNode})
	if _, err := Get(doc, "profiles"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestRedactSecrets(t *testing.T) {
	doc := NewEmptyDocument()
	var err error
	doc, err = SetString(doc, "profiles.dev.auth.accessToken", "secret")
	if err != nil {
		t.Fatalf("set: %v", err)
	}
	red, err := RedactSecrets(doc)
	if err != nil {
		t.Fatalf("redact: %v", err)
	}
	got, err := Get(red, "profiles.dev.auth.accessToken")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "REDACTED" {
		t.Fatalf("got %q", got)
	}
	orig, _ := Get(doc, "profiles.dev.auth.accessToken")
	if orig != "secret" {
		t.Fatalf("original mutated")
	}
}

func TestIsSecretKey(t *testing.T) {
	if !IsSecretKey("profiles.dev.auth.accessToken") {
		t.Fatalf("expected secret")
	}
	if IsSecretKey("profiles.dev.apiUrl") {
		t.Fatalf("unexpected secret")
	}
}

func TestSetString_InvalidKey(t *testing.T) {
	doc := NewEmptyDocument()
	if _, err := SetString(doc, "", "x"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := SetString(doc, "a..b", "x"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestErrNotFound_ErrorString(t *testing.T) {
	err := ErrNotFound{Key: "a.b"}
	if got := err.Error(); got == "" {
		t.Fatalf("expected non-empty error")
	}
	if got := err.Error(); got != "key not found: a.b" {
		t.Fatalf("got %q", got)
	}
}
