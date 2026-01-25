package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/config"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
)

func TestAuthTokenSet_InvalidJWTIsUsageExit2(t *testing.T) {
	store := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"auth", "token", "set", "--token", "nope"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected usage exit 2, got %d", exitcode.Code(err))
	}
}

func TestAuthStatus_NoTokenIsAuthExit3(t *testing.T) {
	store := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"auth", "status"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Auth {
		t.Fatalf("expected auth exit 3, got %d", exitcode.Code(err))
	}
}

func TestAuthLogout_NoTokenIsAuthExit3(t *testing.T) {
	store := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"auth", "logout"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Auth {
		t.Fatalf("expected auth exit 3, got %d", exitcode.Code(err))
	}
}

func TestAuthTokenPrint_JSONIsSimpleObject(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.default.auth.accessToken", "a.b.c")
	store := &memStore{path: "/x", doc: doc}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "auth", "token", "print"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr, got %q", stderr.String())
	}

	var got map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout not json: %v", err)
	}
	if got["token"] != "a.b.c" {
		t.Fatalf("token: %#v", got["token"])
	}
}

func TestAuthTokenSet_OK_TableDoesNotLeakToken(t *testing.T) {
	store := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"auth", "token", "set", "--token", "a.b.c"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if bytes.Contains(stdout.Bytes(), []byte("a.b.c")) || bytes.Contains(stderr.Bytes(), []byte("a.b.c")) {
		t.Fatalf("token leaked")
	}
}

func TestAuthStatus_OK_TableSingleLine(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.default.auth.accessToken", "a.b.c")
	store := &memStore{path: "/x", doc: doc}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"auth", "status"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if bytes.Count(stdout.Bytes(), []byte("\n")) != 1 {
		t.Fatalf("expected single line, got %q", stdout.String())
	}
	if bytes.Contains(stdout.Bytes(), []byte("a.b.c")) || bytes.Contains(stderr.Bytes(), []byte("a.b.c")) {
		t.Fatalf("token leaked")
	}
}

func TestAuthLogout_OK_JSON(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.default.auth.accessToken", "a.b.c")
	store := &memStore{path: "/x", doc: doc}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "auth", "logout"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if bytes.Contains(stdout.Bytes(), []byte("a.b.c")) || bytes.Contains(stderr.Bytes(), []byte("a.b.c")) {
		t.Fatalf("token leaked")
	}
}

func TestAuthStatus_OK_JSON(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.default.auth.accessToken", "a.b.c")
	store := &memStore{path: "/x", doc: doc}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "auth", "status"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if bytes.Contains(stdout.Bytes(), []byte("a.b.c")) || bytes.Contains(stderr.Bytes(), []byte("a.b.c")) {
		t.Fatalf("token leaked")
	}

	var got map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout not json: %v", err)
	}
	data := got["data"].(map[string]any)
	if data["tokenConfigured"] != true {
		t.Fatalf("tokenConfigured=%#v", data["tokenConfigured"])
	}
	if data["profile"] != "default" {
		t.Fatalf("profile=%#v", data["profile"])
	}
}

func TestAuthLogout_OK_Table(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.default.auth.accessToken", "a.b.c")
	store := &memStore{path: "/x", doc: doc}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"auth", "logout"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("OK")) {
		t.Fatalf("stdout=%q", stdout.String())
	}
	if bytes.Contains(stdout.Bytes(), []byte("a.b.c")) || bytes.Contains(stderr.Bytes(), []byte("a.b.c")) {
		t.Fatalf("token leaked")
	}
}

func TestAuthTokenSet_OK_JSON(t *testing.T) {
	store := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "auth", "token", "set", "--token", "a.b.c"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if bytes.Contains(stdout.Bytes(), []byte("a.b.c")) || bytes.Contains(stderr.Bytes(), []byte("a.b.c")) {
		t.Fatalf("token leaked")
	}
}

func TestAuthTokenPrint_Table(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.default.auth.accessToken", "a.b.c")
	store := &memStore{path: "/x", doc: doc}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"auth", "token", "print"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr, got %q", stderr.String())
	}
	if string(bytes.TrimSpace(stdout.Bytes())) != "a.b.c" {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestAuthTokenPrint_MissingIsAuthExit3(t *testing.T) {
	store := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: nil, ConfigStore: store, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"auth", "token", "print"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Auth {
		t.Fatalf("expected auth exit 3, got %d", exitcode.Code(err))
	}
}
