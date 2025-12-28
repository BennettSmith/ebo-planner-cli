package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestJSONOutput_HasNoANSIEscapeCodes(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeTripReadAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "trip", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if strings.Contains(stdout.String(), "\x1b") {
		t.Fatalf("stdout contains ansi: %q", stdout.String())
	}
}
