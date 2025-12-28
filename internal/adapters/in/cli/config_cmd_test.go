package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/BennettSmith/ebo-planner-cli/internal/platform/config"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
)

type memStore struct {
	path string
	doc  config.Document
}

func (m memStore) Path(ctx context.Context) (string, error)             { return m.path, nil }
func (m memStore) Load(ctx context.Context) (config.Document, error)    { return m.doc, nil }
func (m *memStore) Save(ctx context.Context, doc config.Document) error { m.doc = doc; return nil }

func TestConfigPath_JSON(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	store := &memStore{path: "/tmp/config.yaml", doc: config.NewEmptyDocument()}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "config", "path"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout not json: %v", err)
	}
	data := got["data"].(map[string]any)
	if data["path"] != "/tmp/config.yaml" {
		t.Fatalf("path: %#v", data["path"])
	}
}

func TestConfigList_RedactsInTable(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.dev.auth.accessToken", "secret")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"config", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if bytes.Contains(stdout.Bytes(), []byte("secret")) {
		t.Fatalf("expected redacted")
	}
	if !bytes.Contains(stdout.Bytes(), []byte("REDACTED")) {
		t.Fatalf("expected REDACTED marker")
	}
}

func TestConfigList_IncludeSecrets_JSONOnly(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.dev.auth.accessToken", "secret")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "config", "list", "--include-secrets"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("secret")) {
		t.Fatalf("expected secret included in json")
	}
}

func TestConfigGet_NotFoundExitCode(t *testing.T) {
	store := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"config", "get", "nope"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.NotFound {
		t.Fatalf("expected exit 4, got %d", exitcode.Code(err))
	}
}

func TestConfigGet_SecretRedacted_Table(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.dev.auth.accessToken", "secret")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"config", "get", "profiles.dev.auth.accessToken"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if bytes.Contains(stdout.Bytes(), []byte("secret")) {
		t.Fatalf("expected redacted")
	}
	if !bytes.Contains(stdout.Bytes(), []byte("REDACTED")) {
		t.Fatalf("expected REDACTED")
	}
}

func TestConfigSet_JSONRedactsSecretValue(t *testing.T) {
	store := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "config", "set", "profiles.dev.auth.accessToken", "secret"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if bytes.Contains(stdout.Bytes(), []byte("secret")) {
		t.Fatalf("expected secret redacted in json response")
	}
}

func TestConfigUnset_JSON(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.dev.apiUrl", "http://x")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "config", "unset", "profiles.dev.apiUrl"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
}

func TestConfigList_JSONRedactsByDefault(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.dev.auth.accessToken", "secret")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "config", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if bytes.Contains(stdout.Bytes(), []byte("secret")) {
		t.Fatalf("expected redacted")
	}
	if !bytes.Contains(stdout.Bytes(), []byte("REDACTED")) {
		t.Fatalf("expected REDACTED")
	}
}

func TestConfigSet_TableOKAndPersists(t *testing.T) {
	store := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"config", "set", "profiles.dev.apiUrl", "http://x"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("OK")) {
		t.Fatalf("expected OK")
	}
	// Verify persisted
	val, err := config.Get(store.doc, "profiles.dev.apiUrl")
	if err != nil || val != "http://x" {
		t.Fatalf("val %q err %v", val, err)
	}
}

func TestConfigUnset_TableOK(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.dev.apiUrl", "http://x")
	store := &memStore{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"config", "unset", "profiles.dev.apiUrl"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("OK")) {
		t.Fatalf("expected OK")
	}
}

func TestConfigGet_InvalidKeyIsUsage(t *testing.T) {
	store := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"config", "get", "a..b"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected usage, got %d", exitcode.Code(err))
	}
}

func TestConfigPath_Table(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	store := &memStore{path: "/tmp/config.yaml", doc: config.NewEmptyDocument()}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"config", "path"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("/tmp/config.yaml")) {
		t.Fatalf("expected path, got %q", stdout.String())
	}
}

func TestConfigPath_NilStoreIsUnexpectedExitCode(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: nil, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"config", "path"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Unexpected {
		t.Fatalf("expected unexpected exit 1, got %d", exitcode.Code(err))
	}
}
