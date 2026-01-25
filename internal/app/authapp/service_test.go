package authapp

import (
	"context"
	"testing"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/config"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
)

type memStore struct{ doc config.Document }

func (m memStore) Path(ctx context.Context) (string, error)             { return "/x", nil }
func (m memStore) Load(ctx context.Context) (config.Document, error)    { return m.doc, nil }
func (m *memStore) Save(ctx context.Context, doc config.Document) error { m.doc = doc; return nil }

func TestTokenSet_ValidatesJWTShape(t *testing.T) {
	s := Service{Store: &memStore{doc: config.NewEmptyDocument()}}
	if err := s.TokenSet(context.Background(), "default", "notajwt"); err == nil {
		t.Fatalf("expected error")
	} else if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected usage, got %d", exitcode.Code(err))
	}
}

func TestTokenSet_PersistsAccessTokenAndTokenType(t *testing.T) {
	m := &memStore{doc: config.NewEmptyDocument()}
	s := Service{Store: m}

	tok := "a.b.c"
	if err := s.TokenSet(context.Background(), "default", tok); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, _ := config.Get(m.doc, "profiles.default.auth.accessToken")
	if got != tok {
		t.Fatalf("token: got %q", got)
	}
	tt, _ := config.Get(m.doc, "profiles.default.auth.tokenType")
	if tt != "Bearer" {
		t.Fatalf("tokenType: got %q", tt)
	}
	if _, err := config.Get(m.doc, "profiles.default.auth.expiresAt"); err == nil {
		t.Fatalf("expected expiresAt untouched")
	}
}

func TestStatus_NoTokenIsAuthError3(t *testing.T) {
	m := &memStore{doc: config.NewEmptyDocument()}
	s := Service{Store: m}
	_, err := s.Status(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Auth {
		t.Fatalf("expected auth exit 3, got %d", exitcode.Code(err))
	}
}

func TestLogout_NoTokenIsAuthError3(t *testing.T) {
	m := &memStore{doc: config.NewEmptyDocument()}
	s := Service{Store: m}
	err := s.Logout(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Auth {
		t.Fatalf("expected auth exit 3, got %d", exitcode.Code(err))
	}
}

func TestLogout_ClearsFields(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.default.auth.accessToken", "a.b.c")
	doc, _ = config.SetString(doc, "profiles.default.auth.tokenType", "Bearer")
	doc, _ = config.SetString(doc, "profiles.default.auth.expiresAt", "2026-01-01T00:00:00Z")

	m := &memStore{doc: doc}
	s := Service{Store: m}
	if err := s.Logout(context.Background()); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := config.Get(m.doc, "profiles.default.auth.accessToken"); err == nil {
		t.Fatalf("expected token removed")
	}
	if _, err := config.Get(m.doc, "profiles.default.auth.tokenType"); err == nil {
		t.Fatalf("expected tokenType removed")
	}
	if _, err := config.Get(m.doc, "profiles.default.auth.expiresAt"); err == nil {
		t.Fatalf("expected expiresAt removed")
	}
}

func TestTokenPrint_Success(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.default.auth.accessToken", "a.b.c")
	m := &memStore{doc: doc}
	s := Service{Store: m}

	tok, profile, err := s.TokenPrint(context.Background())
	if err != nil {
		t.Fatalf("print: %v", err)
	}
	if profile != "default" {
		t.Fatalf("profile: %q", profile)
	}
	if tok != "a.b.c" {
		t.Fatalf("token: %q", tok)
	}
}

func TestTokenSet_UsesCurrentProfileWhenEmpty(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithCurrentProfile(doc, "dev")
	m := &memStore{doc: doc}
	s := Service{Store: m}

	if err := s.TokenSet(context.Background(), "", "a.b.c"); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, _ := config.Get(m.doc, "profiles.dev.auth.accessToken")
	if got != "a.b.c" {
		t.Fatalf("token: %q", got)
	}
}

type loadErrStore struct{ err error }

func (l loadErrStore) Path(ctx context.Context) (string, error) { return "/x", nil }
func (l loadErrStore) Load(ctx context.Context) (config.Document, error) {
	return config.Document{}, l.err
}
func (l loadErrStore) Save(ctx context.Context, doc config.Document) error { return nil }

func TestStatus_LoadErrorIsServer(t *testing.T) {
	s := Service{Store: loadErrStore{err: context.Canceled}}
	_, err := s.Status(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Server {
		t.Fatalf("expected server, got %d", exitcode.Code(err))
	}
}

func TestTokenPrint_LoadErrorIsServer(t *testing.T) {
	s := Service{Store: loadErrStore{err: context.Canceled}}
	_, _, err := s.TokenPrint(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Server {
		t.Fatalf("expected server, got %d", exitcode.Code(err))
	}
}

func TestTokenSet_EmptyPartIsUsage(t *testing.T) {
	s := Service{Store: &memStore{doc: config.NewEmptyDocument()}}
	if err := s.TokenSet(context.Background(), "default", "a..c"); err == nil {
		t.Fatalf("expected error")
	} else if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected usage, got %d", exitcode.Code(err))
	}
}
