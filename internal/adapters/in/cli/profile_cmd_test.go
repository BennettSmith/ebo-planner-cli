package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/config"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
)

func TestProfileList_JSON(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://a")
	doc, _ = config.WithProfileAPIURL(doc, "dev", "http://b")
	doc, _ = config.WithCurrentProfile(doc, "dev")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "profile", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout not json: %v", err)
	}
	data := got["data"].(map[string]any)
	if data["currentProfile"] != "dev" {
		t.Fatalf("currentProfile: %#v", data["currentProfile"])
	}
	profiles := data["profiles"].([]any)
	if len(profiles) != 2 {
		t.Fatalf("profiles len: %d", len(profiles))
	}
}

func TestProfileShow_Table(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://a")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"profile", "show", "default"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("Name: default")) {
		t.Fatalf("expected Name line, got %q", stdout.String())
	}
}

func TestProfileCreate_ConflictExitCode(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "dev", "http://x")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"profile", "create", "dev", "--api-url", "http://y"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Conflict {
		t.Fatalf("expected conflict exit 5, got %d", exitcode.Code(err))
	}
}

func TestProfileUse_MissingIsUsageExitCode(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://a")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"profile", "use", "nope"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected usage exit 2, got %d", exitcode.Code(err))
	}
}

func TestProfileDelete_LastProfileConflictExitCode(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://a")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"profile", "delete", "default"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Conflict {
		t.Fatalf("expected conflict exit 5, got %d", exitcode.Code(err))
	}
}

func TestProfileDelete_CurrentProfileSwitchesToDefault(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://a")
	doc, _ = config.WithProfileAPIURL(doc, "dev", "http://b")
	doc, _ = config.WithCurrentProfile(doc, "dev")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"profile", "delete", "dev"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	v, _ := config.ViewOf(store.doc)
	if v.CurrentProfile != "default" {
		t.Fatalf("currentProfile: got %q", v.CurrentProfile)
	}
	if _, ok := v.Profiles["dev"]; ok {
		t.Fatalf("expected dev profile deleted")
	}
}

func TestProfileSet_CreatesIfMissing(t *testing.T) {
	store := &memStore{path: "/x", doc: config.NewEmptyDocument()}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"profile", "set", "staging", "--api-url", "http://s"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	v, _ := config.ViewOf(store.doc)
	if v.Profiles["staging"].APIURL != "http://s" {
		t.Fatalf("apiUrl: got %q", v.Profiles["staging"].APIURL)
	}
}

func TestProfileList_Table(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://a")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"profile", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("NAME")) {
		t.Fatalf("expected header, got %q", stdout.String())
	}
}

func TestProfileCreate_Success_JSONPersists(t *testing.T) {
	store := &memStore{path: "/x", doc: config.NewEmptyDocument()}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "profile", "create", "dev", "--api-url", "http://x"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	val, err := config.Get(store.doc, "profiles.dev.apiUrl")
	if err != nil || val != "http://x" {
		t.Fatalf("persisted apiUrl %q err %v", val, err)
	}
}

func TestProfileShow_JSON(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://a")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "profile", "show", "default"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("\"profile\"")) {
		t.Fatalf("expected json profile object, got %q", stdout.String())
	}
}

func TestProfileUse_Success_JSON(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://a")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "profile", "use", "default"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	v, _ := config.ViewOf(store.doc)
	if v.CurrentProfile != "default" {
		t.Fatalf("currentProfile: got %q", v.CurrentProfile)
	}
}

func TestProfileDelete_JSON(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://a")
	doc, _ = config.WithProfileAPIURL(doc, "dev", "http://b")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "profile", "delete", "dev"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("\"deleted\"")) {
		t.Fatalf("expected deleted field, got %q", stdout.String())
	}
}

func TestProfileCreate_TableOK(t *testing.T) {
	store := &memStore{path: "/x", doc: config.NewEmptyDocument()}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"profile", "create", "dev", "--api-url", "http://x"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("OK")) {
		t.Fatalf("expected OK, got %q", stdout.String())
	}
}

func TestProfileUse_TableOK(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://a")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"profile", "use", "default"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("OK")) {
		t.Fatalf("expected OK, got %q", stdout.String())
	}
}

func TestProfileSet_JSON(t *testing.T) {
	store := &memStore{path: "/x", doc: config.NewEmptyDocument()}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "profile", "set", "staging", "--api-url", "http://s"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("\"apiUrl\"")) {
		t.Fatalf("expected apiUrl in json, got %q", stdout.String())
	}
}
