package cli

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/cliopts"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/config"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
)

type noopOpener struct{}

func (noopOpener) Open(url string) error { return nil }

type memStore2 struct {
	path string
	doc  config.Document
}

func (m memStore2) Path(ctx context.Context) (string, error)             { return m.path, nil }
func (m memStore2) Load(ctx context.Context) (config.Document, error)    { return m.doc, nil }
func (m *memStore2) Save(ctx context.Context, doc config.Document) error { m.doc = doc; return nil }

func TestAuthLogin_MissingOIDCIsUsageExit2(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://x")
	store := &memStore2{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: cliopts.MapEnv{}, ConfigStore: store, Stdout: stdout, Stderr: stderr, BrowserOpener: noopOpener{}})
	cmd.SetArgs([]string{"auth", "login"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected usage exit 2, got %d", exitcode.Code(err))
	}
}

func TestAuthLogin_SuccessPersistsAndPrintsGuidanceToStderr(t *testing.T) {
	var base string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_, _ = w.Write([]byte(`{"device_authorization_endpoint":"` + base + `/device","token_endpoint":"` + base + `/token"}`))
		case "/device":
			_, _ = w.Write([]byte(`{"device_code":"dc","user_code":"UC","verification_uri":"` + base + `/verify","expires_in":600,"interval":0}`))
		case "/token":
			_, _ = w.Write([]byte(`{"access_token":"a.b.c","token_type":"Bearer","expires_in":60}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	base = srv.URL

	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://x")
	doc, _ = config.WithProfileOIDC(doc, "default", base, "cid", []string{"openid"})
	store := &memStore2{path: "/x", doc: doc}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCmd(RootDeps{Env: cliopts.MapEnv{}, ConfigStore: store, Stdout: stdout, Stderr: stderr, BrowserOpener: noopOpener{}})
	cmd.SetArgs([]string{"--timeout", "2s", "auth", "login"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("Open:")) || !bytes.Contains(stderr.Bytes(), []byte("Code:")) {
		t.Fatalf("stderr=%q", stderr.String())
	}
	if bytes.Contains(stderr.Bytes(), []byte("a.b.c")) || bytes.Contains(stdout.Bytes(), []byte("a.b.c")) {
		t.Fatalf("token leaked")
	}
	// persisted
	got, _ := config.Get(store.doc, "profiles.default.auth.accessToken")
	if got != "a.b.c" {
		t.Fatalf("token %q", got)
	}
}
