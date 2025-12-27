package configfile

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BennettSmith/ebo-planner-cli/internal/platform/config"
	"gopkg.in/yaml.v3"
)

type mapEnv map[string]string

func (m mapEnv) LookupEnv(key string) (string, bool) {
	v, ok := m[key]
	return v, ok
}

func TestPath_UsesEBOConfigDirOverride(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()

	s := Store{Env: mapEnv{"EBO_CONFIG_DIR": tmp}}
	p, err := s.Path(ctx)
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	want := filepath.Join(tmp, "ebo", "config.yaml")
	if p != want {
		t.Fatalf("got %q want %q", p, want)
	}
}

func TestSaveLoad_PreservesUnknownFields(t *testing.T) {
	ctx := context.Background()
	base := t.TempDir()

	s := Store{Env: mapEnv{"EBO_CONFIG_DIR": base}}
	p, err := s.Path(ctx)
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	original := strings.TrimSpace(`currentProfile: staging
x-top: 123
profiles:
  staging:
    apiUrl: https://old.example
    x-unknown: keepme
`) + "\n"
	if err := os.WriteFile(p, []byte(original), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	doc, err := s.Load(ctx)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	doc, err = config.WithCurrentProfile(doc, "staging")
	if err != nil {
		t.Fatalf("with current: %v", err)
	}
	doc, err = config.WithProfileAPIURL(doc, "staging", "https://new.example")
	if err != nil {
		t.Fatalf("with apiUrl: %v", err)
	}

	if err := s.Save(ctx, doc); err != nil {
		t.Fatalf("save: %v", err)
	}

	reloaded, err := s.Load(ctx)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	b, err := yaml.Marshal(reloaded.Root)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := string(b)
	if !strings.Contains(out, "x-top: 123") {
		t.Fatalf("expected top-level unknown field preserved, got:\n%s", out)
	}
	if !strings.Contains(out, "x-unknown: keepme") {
		t.Fatalf("expected profile unknown field preserved, got:\n%s", out)
	}
	if !strings.Contains(out, "apiUrl: https://new.example") {
		t.Fatalf("expected apiUrl updated, got:\n%s", out)
	}

	st, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("expected 0600 perms, got %o", st.Mode().Perm())
	}
}

func TestLoad_MissingFileReturnsEmptyDoc(t *testing.T) {
	ctx := context.Background()
	base := t.TempDir()
	s := Store{Env: mapEnv{"EBO_CONFIG_DIR": base}}
	doc, err := s.Load(ctx)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	v, err := config.ViewOf(doc)
	if err != nil {
		t.Fatalf("view: %v", err)
	}
	if v.CurrentProfile != "" {
		t.Fatalf("currentProfile: got %q", v.CurrentProfile)
	}
}

func TestSave_NilDocumentErrors(t *testing.T) {
	ctx := context.Background()
	base := t.TempDir()
	s := Store{Env: mapEnv{"EBO_CONFIG_DIR": base}}
	if err := s.Save(ctx, config.Document{}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestPath_UsesUserConfigDirWhenNoOverride(t *testing.T) {
	ctx := context.Background()
	s := Store{Env: mapEnv{}}
	p, err := s.Path(ctx)
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if !strings.HasSuffix(p, string(os.PathSeparator)+"ebo"+string(os.PathSeparator)+"config.yaml") {
		t.Fatalf("unexpected path: %q", p)
	}
}

func TestLoad_InvalidYAMLReturnsError(t *testing.T) {
	ctx := context.Background()
	base := t.TempDir()
	s := Store{Env: mapEnv{"EBO_CONFIG_DIR": base}}
	p, err := s.Path(ctx)
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(p, []byte(": ["), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := s.Load(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSave_CreatesParentDir(t *testing.T) {
	ctx := context.Background()
	base := t.TempDir()
	s := Store{Env: mapEnv{"EBO_CONFIG_DIR": base}}

	doc := config.NewEmptyDocument()
	doc, _ = config.SetString(doc, "currentProfile", "default")

	if err := s.Save(ctx, doc); err != nil {
		t.Fatalf("save: %v", err)
	}

	p, err := s.Path(ctx)
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("stat: %v", err)
	}
}

func TestPath_UsesOSEnv(t *testing.T) {
	ctx := context.Background()
	base := t.TempDir()
	// Set process env so OSEnv.LookupEnv is exercised.
	if err := os.Setenv("EBO_CONFIG_DIR", base); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("EBO_CONFIG_DIR") })

	s := Store{Env: OSEnv{}}
	p, err := s.Path(ctx)
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	want := filepath.Join(base, "ebo", "config.yaml")
	if p != want {
		t.Fatalf("got %q want %q", p, want)
	}
}

func TestSave_FailsWhenConfigDirIsFile(t *testing.T) {
	ctx := context.Background()
	base := t.TempDir()
	// Create a file named "ebo" where a directory is expected.
	if err := os.WriteFile(filepath.Join(base, "ebo"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	s := Store{Env: mapEnv{"EBO_CONFIG_DIR": base}}

	doc := config.NewEmptyDocument()
	if err := s.Save(ctx, doc); err == nil {
		t.Fatalf("expected error")
	}
}
