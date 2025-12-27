package specpin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadSpecTag(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "spec.lock")
	if err := os.WriteFile(p, []byte("v1\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	tag, err := ReadSpecTag(p)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if tag != "v1" {
		t.Fatalf("got %q", tag)
	}
}

func TestReadSpecTag_EmptyErrors(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "spec.lock")
	if err := os.WriteFile(p, []byte("\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := ReadSpecTag(p); err == nil {
		t.Fatalf("expected error")
	}
}
