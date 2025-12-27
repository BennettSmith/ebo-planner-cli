package config

import "testing"

func TestNewEmptyDocument_ViewAndMutations(t *testing.T) {
	doc := NewEmptyDocument()

	v, err := ViewOf(doc)
	if err != nil {
		t.Fatalf("view: %v", err)
	}
	if v.CurrentProfile != "" {
		t.Fatalf("currentProfile: got %q", v.CurrentProfile)
	}

	doc, err = WithCurrentProfile(doc, "staging")
	if err != nil {
		t.Fatalf("with current: %v", err)
	}
	doc, err = WithProfileAPIURL(doc, "staging", "https://api")
	if err != nil {
		t.Fatalf("with api: %v", err)
	}

	v, err = ViewOf(doc)
	if err != nil {
		t.Fatalf("view2: %v", err)
	}
	if v.CurrentProfile != "staging" {
		t.Fatalf("currentProfile: got %q", v.CurrentProfile)
	}
	if v.Profiles["staging"].APIURL != "https://api" {
		t.Fatalf("apiUrl: got %q", v.Profiles["staging"].APIURL)
	}
}

func TestRootMapping_InvalidDocs(t *testing.T) {
	_, err := rootMapping(Document{})
	if err == nil {
		t.Fatalf("expected error")
	}
}
