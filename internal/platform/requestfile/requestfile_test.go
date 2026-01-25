package requestfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	gen "github.com/Overland-East-Bay/trip-planner-cli/internal/gen/plannerapi"
)

func writeTemp(t *testing.T, ext string, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "req-*"+ext)
	if err != nil {
		t.Fatalf("tempfile: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	return f.Name()
}

func TestLoadStrict_JSON_ByExtension(t *testing.T) {
	p := writeTemp(t, ".json", `{"name":"Trip 1"}`)
	var req gen.CreateTripDraftRequest
	if err := LoadStrict(p, &req); err != nil {
		t.Fatalf("LoadStrict: %v", err)
	}
	if req.Name != "Trip 1" {
		t.Fatalf("name: %q", req.Name)
	}
}

func TestLoadStrict_YAML_ByExtension_MultilinePreserved(t *testing.T) {
	p := writeTemp(t, ".yaml", ""+
		"description: |\n"+
		"  line1\n"+
		"  line2\n")

	var req gen.UpdateTripRequest
	if err := LoadStrict(p, &req); err != nil {
		t.Fatalf("LoadStrict: %v", err)
	}
	if req.Description == nil {
		t.Fatalf("expected description set")
	}
	if *req.Description != "line1\nline2\n" {
		t.Fatalf("description: %#v", *req.Description)
	}
}

func TestLoadStrict_UnknownExtension_FallbackJSONThenYAML(t *testing.T) {
	p := writeTemp(t, ".txt", `name: Trip 2`)
	var req gen.CreateTripDraftRequest
	if err := LoadStrict(p, &req); err != nil {
		t.Fatalf("LoadStrict: %v", err)
	}
	if req.Name != "Trip 2" {
		t.Fatalf("name: %q", req.Name)
	}
}

func TestLoadStrict_StrictUnknownFields(t *testing.T) {
	p := writeTemp(t, ".json", `{"name":"Trip 1","extra":1}`)
	var req gen.CreateTripDraftRequest
	if err := LoadStrict(p, &req); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoadStrict_JSON_ExtensionDoesNotFallbackToYAML(t *testing.T) {
	// YAML content but .json extension must be treated as JSON and fail.
	p := writeTemp(t, ".json", "name: Trip 3")
	var req gen.CreateTripDraftRequest
	if err := LoadStrict(p, &req); err == nil {
		t.Fatalf("expected error")
	}
}

func TestNormalizeYAML_RejectsNonStringKeys(t *testing.T) {
	_, err := normalizeYAML(map[any]any{1: "x"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestNormalizeYAML_MapStringAny_Recurses(t *testing.T) {
	v, err := normalizeYAML(map[string]any{
		"a": map[string]any{"b": "c"},
	})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	m := v.(map[string]any)
	a := m["a"].(map[string]any)
	if a["b"] != "c" {
		t.Fatalf("got: %#v", v)
	}
}

func TestNormalizeYAML_MapAnyAny_StringKeys_Succeeds(t *testing.T) {
	v, err := normalizeYAML(map[any]any{"a": map[any]any{"b": "c"}})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	m := v.(map[string]any)
	a := m["a"].(map[string]any)
	if a["b"] != "c" {
		t.Fatalf("got: %#v", v)
	}
}

func TestNormalizeYAML_Slice_Recurses(t *testing.T) {
	v, err := normalizeYAML([]any{map[any]any{"k": "v"}})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	s := v.([]any)
	m := s[0].(map[string]any)
	if m["k"] != "v" {
		t.Fatalf("got: %#v", v)
	}
}

func TestLoadStrict_PathIsInErrorMessageOnFallbackFailure(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.req")
	if err := os.WriteFile(p, []byte("{"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	var req gen.CreateTripDraftRequest
	if err := LoadStrict(p, &req); err == nil {
		t.Fatalf("expected error")
	} else if !strings.Contains(err.Error(), p) {
		t.Fatalf("expected path in error, got: %v", err)
	}
}

func TestLoadStrict_YML_ByExtension(t *testing.T) {
	p := writeTemp(t, ".yml", "name: Trip\n")
	var req gen.CreateTripDraftRequest
	if err := LoadStrict(p, &req); err != nil {
		t.Fatalf("LoadStrict: %v", err)
	}
	if req.Name != "Trip" {
		t.Fatalf("name: %q", req.Name)
	}
}

func TestLoadStrict_UnknownExtension_FallbackBothFail_IncludesBothErrors(t *testing.T) {
	p := writeTemp(t, ".req", "{")
	var req gen.CreateTripDraftRequest
	err := LoadStrict(p, &req)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "JSON:") || !strings.Contains(err.Error(), "YAML:") {
		t.Fatalf("expected both JSON and YAML errors, got: %v", err)
	}
}

func TestLoadStrict_JSON_TrailingContentRejected(t *testing.T) {
	p := writeTemp(t, ".json", `{"name":"Trip"} {"name":"Trip2"}`)
	var req gen.CreateTripDraftRequest
	if err := LoadStrict(p, &req); err == nil {
		t.Fatalf("expected error")
	}
}
