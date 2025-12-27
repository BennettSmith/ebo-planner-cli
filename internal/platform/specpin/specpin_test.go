package specpin

import (
	"errors"
	"path/filepath"
	"testing"
)

var errFake = errors.New("fake")

type fakeRunner struct {
	out map[string][]byte
	err map[string]error
}

func (f fakeRunner) Run(name string, args ...string) ([]byte, error) {
	k := name + " " + join(args)
	if err, ok := f.err[k]; ok {
		return f.out[k], err
	}
	return f.out[k], nil
}

func join(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += " "
		}
		out += s
	}
	return out
}

func TestResolveSpecDir_DefaultSibling(t *testing.T) {
	cwd := "/repo/ebo-planner-cli"
	got := ResolveSpecDir(map[string]string{}, cwd)
	want := filepath.Clean("/repo/ebo-planner-spec")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveSpecDir_EnvAbsolute(t *testing.T) {
	got := ResolveSpecDir(map[string]string{"EBO_SPEC_DIR": "/x/spec"}, "/repo")
	if got != "/x/spec" {
		t.Fatalf("got %q", got)
	}
}

func TestVerifyRefExists_OK(t *testing.T) {
	r := fakeRunner{out: map[string][]byte{
		"git -C /spec rev-parse --verify v1": []byte("v1\n"),
	}, err: map[string]error{}}
	if err := VerifyRefExists(r, "/spec", "v1"); err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
}

func TestExtractOpenAPIAtRef_OK(t *testing.T) {
	r := fakeRunner{out: map[string][]byte{
		"git -C /spec rev-parse --verify v1":        []byte("v1\n"),
		"git -C /spec show v1:openapi/openapi.yaml": []byte("openapi: 3.0.0\n"),
	}, err: map[string]error{}}

	d := t.TempDir()
	outPath := filepath.Join(d, "openapi.yaml")
	if err := ExtractOpenAPIAtRef(r, "/spec", "v1", outPath); err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
}

func TestVerifyRefExists_MissingRef(t *testing.T) {
	r := fakeRunner{out: map[string][]byte{}, err: map[string]error{
		"git -C /spec rev-parse --verify nope": errFake,
	}}
	if err := VerifyRefExists(r, "/spec", "nope"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestExtractOpenAPIAtRef_GitShowFails(t *testing.T) {
	r := fakeRunner{out: map[string][]byte{
		"git -C /spec rev-parse --verify v1": []byte("v1\n"),
	}, err: map[string]error{
		"git -C /spec show v1:openapi/openapi.yaml": errFake,
	}}
	d := t.TempDir()
	outPath := filepath.Join(d, "openapi.yaml")
	if err := ExtractOpenAPIAtRef(r, "/spec", "v1", outPath); err == nil {
		t.Fatalf("expected error")
	}
}

func TestExecRunner_RunSuccessAndFailure(t *testing.T) {
	r := ExecRunner{}
	if _, err := r.Run("/usr/bin/true"); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if _, err := r.Run("/usr/bin/false"); err == nil {
		t.Fatalf("expected failure")
	}
}
