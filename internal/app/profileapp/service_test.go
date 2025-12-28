package profileapp

import (
	"context"
	"testing"

	"github.com/BennettSmith/ebo-planner-cli/internal/platform/config"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
)

type memStore struct {
	doc config.Document
}

func (m *memStore) Path(ctx context.Context) (string, error)            { return "/x", nil }
func (m *memStore) Load(ctx context.Context) (config.Document, error)   { return m.doc, nil }
func (m *memStore) Save(ctx context.Context, doc config.Document) error { m.doc = doc; return nil }

func TestList_DefaultCurrentProfileWhenUnset(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://a")
	doc, _ = config.WithProfileAPIURL(doc, "dev", "http://b")

	m := &memStore{doc: doc}
	s := Service{Store: m}

	list, current, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if current != "default" {
		t.Fatalf("current: got %q", current)
	}
	if len(list) != 2 {
		t.Fatalf("list size: got %d", len(list))
	}
}

func TestShow_DefaultsToCurrentProfileOrDefault(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://a")
	doc, _ = config.WithProfileAPIURL(doc, "dev", "http://b")
	doc, _ = config.WithCurrentProfile(doc, "dev")

	m := &memStore{doc: doc}
	s := Service{Store: m}

	got, err := s.Show(context.Background(), "")
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if got.Name != "dev" {
		t.Fatalf("name: got %q", got.Name)
	}
	if got.APIURL != "http://b" {
		t.Fatalf("apiUrl: got %q", got.APIURL)
	}
}

func TestCreateConflict(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "dev", "http://x")

	m := &memStore{doc: doc}
	s := Service{Store: m}
	if err := s.Create(context.Background(), "dev", "http://y"); err == nil {
		t.Fatalf("expected error")
	} else if exitcode.Code(err) != exitcode.Conflict {
		t.Fatalf("expected conflict, got %d", exitcode.Code(err))
	}
}

func TestSetAPIURL_CreatesIfMissing(t *testing.T) {
	m := &memStore{doc: config.NewEmptyDocument()}
	s := Service{Store: m}

	if err := s.SetAPIURL(context.Background(), "new", "http://x"); err != nil {
		t.Fatalf("set api url: %v", err)
	}
	v, _ := config.ViewOf(m.doc)
	if v.Profiles["new"].APIURL != "http://x" {
		t.Fatalf("apiUrl: got %q", v.Profiles["new"].APIURL)
	}
}

func TestUseMissingIsUsage(t *testing.T) {
	m := &memStore{doc: config.NewEmptyDocument()}
	s := Service{Store: m}
	if err := s.Use(context.Background(), "nope"); err == nil {
		t.Fatalf("expected error")
	} else if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected usage, got %d", exitcode.Code(err))
	}
}

func TestUse_SetsCurrentProfile(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "dev", "http://x")
	m := &memStore{doc: doc}
	s := Service{Store: m}

	if err := s.Use(context.Background(), "dev"); err != nil {
		t.Fatalf("use: %v", err)
	}
	v, _ := config.ViewOf(m.doc)
	if v.CurrentProfile != "dev" {
		t.Fatalf("currentProfile: got %q", v.CurrentProfile)
	}
}

func TestDeleteLastProfileConflict(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://x")
	m := &memStore{doc: doc}
	s := Service{Store: m}
	if err := s.Delete(context.Background(), "default"); err == nil {
		t.Fatalf("expected error")
	} else if exitcode.Code(err) != exitcode.Conflict {
		t.Fatalf("expected conflict")
	}
}

func TestDelete_CurrentProfileSwitchesToDefault(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://a")
	doc, _ = config.WithProfileAPIURL(doc, "dev", "http://b")
	doc, _ = config.WithCurrentProfile(doc, "dev")

	m := &memStore{doc: doc}
	s := Service{Store: m}

	if err := s.Delete(context.Background(), "dev"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	v, _ := config.ViewOf(m.doc)
	if v.CurrentProfile != "default" {
		t.Fatalf("currentProfile: got %q", v.CurrentProfile)
	}
	if _, ok := v.Profiles["dev"]; ok {
		t.Fatalf("expected dev profile deleted")
	}
}

func TestCreate_SuccessPersists(t *testing.T) {
	m := &memStore{doc: config.NewEmptyDocument()}
	s := Service{Store: m}

	if err := s.Create(context.Background(), "dev", "http://x"); err != nil {
		t.Fatalf("create: %v", err)
	}
	v, _ := config.ViewOf(m.doc)
	if v.Profiles["dev"].APIURL != "http://x" {
		t.Fatalf("apiUrl: got %q", v.Profiles["dev"].APIURL)
	}
}

func TestDelete_CurrentWithoutDefaultConflict(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "dev", "http://x")
	doc, _ = config.WithProfileAPIURL(doc, "staging", "http://y")
	doc, _ = config.WithCurrentProfile(doc, "dev")

	m := &memStore{doc: doc}
	s := Service{Store: m}
	if err := s.Delete(context.Background(), "dev"); err == nil {
		t.Fatalf("expected error")
	} else if exitcode.Code(err) != exitcode.Conflict {
		t.Fatalf("expected conflict, got %d", exitcode.Code(err))
	}
}

type errStore struct {
	loadErr error
	saveErr error
	doc     config.Document
}

func (e errStore) Path(ctx context.Context) (string, error) { return "/x", nil }
func (e errStore) Load(ctx context.Context) (config.Document, error) {
	return config.Document{}, e.loadErr
}
func (e errStore) Save(ctx context.Context, doc config.Document) error { return e.saveErr }

type saveErrStore struct {
	doc     config.Document
	saveErr error
}

func (s saveErrStore) Path(ctx context.Context) (string, error)             { return "/x", nil }
func (s saveErrStore) Load(ctx context.Context) (config.Document, error)    { return s.doc, nil }
func (s *saveErrStore) Save(ctx context.Context, doc config.Document) error { return s.saveErr }

func TestSetAPIURL_LoadErrorIsServer(t *testing.T) {
	s := Service{Store: errStore{loadErr: context.Canceled}}
	if err := s.SetAPIURL(context.Background(), "dev", "http://x"); err == nil {
		t.Fatalf("expected error")
	} else if exitcode.Code(err) != exitcode.Server {
		t.Fatalf("expected server, got %d", exitcode.Code(err))
	}
}

func TestSetAPIURL_SaveErrorIsServer(t *testing.T) {
	st := &saveErrStore{doc: config.NewEmptyDocument(), saveErr: context.Canceled}
	s := Service{Store: st}
	if err := s.SetAPIURL(context.Background(), "dev", "http://x"); err == nil {
		t.Fatalf("expected error")
	} else if exitcode.Code(err) != exitcode.Server {
		t.Fatalf("expected server, got %d", exitcode.Code(err))
	}
}

func TestList_LoadErrorIsServer(t *testing.T) {
	s := Service{Store: errStore{loadErr: context.Canceled}}
	_, _, err := s.List(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Server {
		t.Fatalf("expected server, got %d", exitcode.Code(err))
	}
}
