package configapp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/config"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
)

type memStore struct {
	path  string
	doc   config.Document
	saved int
}

func (m *memStore) Path(ctx context.Context) (string, error)          { return m.path, nil }
func (m *memStore) Load(ctx context.Context) (config.Document, error) { return m.doc, nil }
func (m *memStore) Save(ctx context.Context, doc config.Document) error {
	m.doc = doc
	m.saved++
	return nil
}

func TestService_GetNotFoundIsExit4(t *testing.T) {
	m := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	s := Service{Store: m}
	_, err := s.Get(context.Background(), "nope")
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.NotFound {
		t.Fatalf("expected exit 4, got %d", exitcode.Code(err))
	}
}

func TestService_SetAndUnsetPersists(t *testing.T) {
	m := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	s := Service{Store: m}
	ctx := context.Background()

	if err := s.Set(ctx, "profiles.dev.apiUrl", "http://x"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if m.saved != 1 {
		t.Fatalf("saved: %d", m.saved)
	}
	v, err := config.Get(m.doc, "profiles.dev.apiUrl")
	if err != nil || v != "http://x" {
		t.Fatalf("got %q err %v", v, err)
	}

	if err := s.Unset(ctx, "profiles.dev.apiUrl"); err != nil {
		t.Fatalf("unset: %v", err)
	}
	if m.saved != 2 {
		t.Fatalf("saved: %d", m.saved)
	}
}

func TestService_ListYAML_RedactsByDefault(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.dev.auth.accessToken", "secret")
	m := &memStore{path: "/x", doc: doc}
	s := Service{Store: m}

	out, err := s.ListYAML(context.Background(), false)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !contains(out, "REDACTED") {
		t.Fatalf("expected redacted, got %q", out)
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (func() bool { return stringContains(s, sub) })())
}
func stringContains(s, sub string) bool {
	return (len(sub) == 0) || (len(s) >= len(sub) && (indexOf(s, sub) >= 0))
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestService_GetSecretIsRedacted(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.dev.auth.accessToken", "secret")
	m := &memStore{path: "/x", doc: doc}
	s := Service{Store: m}

	val, err := s.Get(context.Background(), "profiles.dev.auth.accessToken")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if val != "REDACTED" {
		t.Fatalf("got %q", val)
	}
}

func TestService_ListJSON_SecretsOptional(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.dev.auth.accessToken", "secret")
	m := &memStore{path: "/x", doc: doc}
	s := Service{Store: m}

	red, err := s.ListJSON(context.Background(), false)
	if err != nil {
		t.Fatalf("list redacted: %v", err)
	}
	b, _ := json.Marshal(red)
	if string(b) == "" {
		t.Fatalf("expected json")
	}
	if string(b) == "secret" || contains(string(b), "secret") {
		t.Fatalf("expected redacted")
	}

	full, err := s.ListJSON(context.Background(), true)
	if err != nil {
		t.Fatalf("list full: %v", err)
	}
	b2, _ := json.Marshal(full)
	if !contains(string(b2), "secret") {
		t.Fatalf("expected secret")
	}
}

type errStore struct {
	loadErr error
	saveErr error
	pathErr error
}

func (e errStore) Path(ctx context.Context) (string, error) { return "", e.pathErr }
func (e errStore) Load(ctx context.Context) (config.Document, error) {
	return config.Document{}, e.loadErr
}
func (e errStore) Save(ctx context.Context, doc config.Document) error { return e.saveErr }

func TestService_Path(t *testing.T) {
	m := &memStore{path: "/p", doc: config.NewEmptyDocument()}
	s := Service{Store: m}
	p, err := s.Path(context.Background())
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if p != "/p" {
		t.Fatalf("got %q", p)
	}
}

func TestService_Path_ErrorIsServer(t *testing.T) {
	s := Service{Store: errStore{pathErr: context.Canceled}}
	_, err := s.Path(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Server {
		t.Fatalf("expected server exit, got %d", exitcode.Code(err))
	}
}

func TestService_Set_InvalidKeyIsUsage(t *testing.T) {
	m := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	s := Service{Store: m}
	err := s.Set(context.Background(), "a..b", "x")
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected usage, got %d", exitcode.Code(err))
	}
}

func TestService_Set_OIDCScopes_JSONArray_PersistsAndParses(t *testing.T) {
	m := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	s := Service{Store: m}
	ctx := context.Background()

	if err := s.Set(ctx, "profiles.dev.oidc.issuerUrl", "http://localhost:8082/realms/ebo"); err != nil {
		t.Fatalf("issuer: %v", err)
	}
	if err := s.Set(ctx, "profiles.dev.oidc.clientId", "ebo-client"); err != nil {
		t.Fatalf("client: %v", err)
	}
	if err := s.Set(ctx, "profiles.dev.oidc.scopes", `["openid","profile","email"]`); err != nil {
		t.Fatalf("scopes: %v", err)
	}

	oc, err := config.OIDCOf(m.doc, "dev")
	if err != nil {
		t.Fatalf("oidc: %v", err)
	}
	if oc.Scopes[0] != "openid" || len(oc.Scopes) != 3 {
		t.Fatalf("scopes %#v", oc.Scopes)
	}
}

func TestService_Set_OIDCScopes_CommaList_PersistsAndParses(t *testing.T) {
	m := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	s := Service{Store: m}
	ctx := context.Background()

	_ = s.Set(ctx, "profiles.dev.oidc.issuerUrl", "http://localhost:8082/realms/ebo")
	_ = s.Set(ctx, "profiles.dev.oidc.clientId", "ebo-client")

	if err := s.Set(ctx, "profiles.dev.oidc.scopes", "openid,profile"); err != nil {
		t.Fatalf("scopes: %v", err)
	}
	oc, err := config.OIDCOf(m.doc, "dev")
	if err != nil {
		t.Fatalf("oidc: %v", err)
	}
	if len(oc.Scopes) != 2 || oc.Scopes[0] != "openid" || oc.Scopes[1] != "profile" {
		t.Fatalf("scopes %#v", oc.Scopes)
	}
}

func TestService_Set_OIDCScopes_SpaceList_PersistsAndParses(t *testing.T) {
	m := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	s := Service{Store: m}
	ctx := context.Background()

	_ = s.Set(ctx, "profiles.dev.oidc.issuerUrl", "http://localhost:8082/realms/ebo")
	_ = s.Set(ctx, "profiles.dev.oidc.clientId", "ebo-client")

	if err := s.Set(ctx, "profiles.dev.oidc.scopes", "openid profile"); err != nil {
		t.Fatalf("scopes: %v", err)
	}
	oc, err := config.OIDCOf(m.doc, "dev")
	if err != nil {
		t.Fatalf("oidc: %v", err)
	}
	if len(oc.Scopes) != 2 || oc.Scopes[0] != "openid" || oc.Scopes[1] != "profile" {
		t.Fatalf("scopes %#v", oc.Scopes)
	}
}

func TestService_Set_OIDCScopes_InvalidJSON_IsUsage(t *testing.T) {
	m := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	s := Service{Store: m}
	err := s.Set(context.Background(), "profiles.dev.oidc.scopes", `["openid",]`)
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected usage, got %d", exitcode.Code(err))
	}
}

func TestService_Set_OIDCScopes_Empty_IsUsage(t *testing.T) {
	m := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	s := Service{Store: m}
	err := s.Set(context.Background(), "profiles.dev.oidc.scopes", "")
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected usage, got %d", exitcode.Code(err))
	}
}

func TestService_Set_OIDCScopes_EmptyJSONArray_IsUsage(t *testing.T) {
	m := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	s := Service{Store: m}
	err := s.Set(context.Background(), "profiles.dev.oidc.scopes", "[]")
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected usage, got %d", exitcode.Code(err))
	}
}

func TestService_ListYAML_IncludeSecrets(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "profiles.dev.auth.accessToken", "secret")
	m := &memStore{path: "/x", doc: doc}
	s := Service{Store: m}

	out, err := s.ListYAML(context.Background(), true)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !contains(out, "secret") {
		t.Fatalf("expected secret present")
	}
}

func TestService_ListJSON_LoadErrorIsServer(t *testing.T) {
	s := Service{Store: errStore{loadErr: context.Canceled}}
	_, err := s.ListJSON(context.Background(), false)
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Server {
		t.Fatalf("expected server, got %d", exitcode.Code(err))
	}
}

func TestService_EnsureStore(t *testing.T) {
	// nil store
	s := Service{}
	if err := s.EnsureStore(); err == nil {
		t.Fatalf("expected error")
	}

	// non-nil store
	m := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	s = Service{Store: m}
	if err := s.EnsureStore(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

type saveErrStore struct {
	doc     config.Document
	saveErr error
}

func (s *saveErrStore) Path(ctx context.Context) (string, error) { return "/x", nil }
func (s *saveErrStore) Load(ctx context.Context) (config.Document, error) {
	return s.doc, nil
}
func (s *saveErrStore) Save(ctx context.Context, doc config.Document) error { return s.saveErr }

func TestService_SaveErrorIsServer(t *testing.T) {
	st := &saveErrStore{doc: config.NewEmptyDocument(), saveErr: context.Canceled}
	svc := Service{Store: st}

	err := svc.Set(context.Background(), "profiles.dev.apiUrl", "http://x")
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Server {
		t.Fatalf("expected server, got %d", exitcode.Code(err))
	}
}

func TestService_Get_InvalidKeyIsUsage(t *testing.T) {
	m := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	s := Service{Store: m}
	_, err := s.Get(context.Background(), "a..b")
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected usage, got %d", exitcode.Code(err))
	}
}

func TestService_Unset_InvalidKeyIsUsage(t *testing.T) {
	m := &memStore{path: "/x", doc: config.NewEmptyDocument()}
	s := Service{Store: m}
	err := s.Unset(context.Background(), "a..b")
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected usage, got %d", exitcode.Code(err))
	}
}

func TestService_Unset_SaveErrorIsServer(t *testing.T) {
	st := &saveErrStore{doc: config.NewEmptyDocument(), saveErr: context.Canceled}
	svc := Service{Store: st}
	if err := svc.Unset(context.Background(), "profiles.dev.apiUrl"); err == nil {
		t.Fatalf("expected error")
	} else if exitcode.Code(err) != exitcode.Server {
		t.Fatalf("expected server, got %d", exitcode.Code(err))
	}
}
