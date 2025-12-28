package editmode

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestResolveEditor_Order(t *testing.T) {
	_ = os.Unsetenv("EBO_EDITOR")
	_ = os.Unsetenv("EDITOR")
	t.Cleanup(func() {
		_ = os.Unsetenv("EBO_EDITOR")
		_ = os.Unsetenv("EDITOR")
	})

	name, args := ResolveEditor()
	if name != "vi" || len(args) != 0 {
		t.Fatalf("default: %q %#v", name, args)
	}

	if err := os.Setenv("EDITOR", "nano"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	name, args = ResolveEditor()
	if name != "nano" || len(args) != 0 {
		t.Fatalf("EDITOR: %q %#v", name, args)
	}

	if err := os.Setenv("EBO_EDITOR", "code --wait"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	name, args = ResolveEditor()
	if name != "code" || strings.Join(args, " ") != "--wait" {
		t.Fatalf("EBO_EDITOR: %q %#v", name, args)
	}
}

func TestEditTemp_UsesResolvedEditorAndReturnsEditedBuffer(t *testing.T) {
	old := execCommand
	t.Cleanup(func() { execCommand = old })

	_ = os.Unsetenv("EBO_EDITOR")
	_ = os.Unsetenv("EDITOR")
	if err := os.Setenv("EBO_EDITOR", "fake-editor"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("EBO_EDITOR") })

	var seenPath string
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name != "fake-editor" {
			t.Fatalf("editor name: %q", name)
		}
		if len(args) != 1 {
			t.Fatalf("expected 1 arg (file), got %d", len(args))
		}
		seenPath = args[0]
		// Simulate an editor by overwriting the file before returning.
		if err := os.WriteFile(seenPath, []byte("{\"k\":\"v\"}\n"), 0o600); err != nil {
			t.Fatalf("write edited: %v", err)
		}
		return exec.Command("true")
	}

	b, err := EditTemp("k: template\n")
	if err != nil {
		t.Fatalf("EditTemp: %v", err)
	}
	if seenPath == "" {
		t.Fatalf("expected editor called with a path")
	}
	if string(b) != "{\"k\":\"v\"}\n" {
		t.Fatalf("buffer: %q", string(b))
	}
	// temp file should be removed
	if _, err := os.Stat(seenPath); err == nil {
		t.Fatalf("expected temp file removed")
	}
}

func TestEditTemp_EmptyTemplate_DefaultsToEmptyMapping(t *testing.T) {
	_ = os.Unsetenv("EBO_EDITOR")
	_ = os.Unsetenv("EDITOR")
	if err := os.Setenv("EBO_EDITOR", "true"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("EBO_EDITOR") })

	b, err := EditTemp("")
	if err != nil {
		t.Fatalf("EditTemp: %v", err)
	}
	if string(b) != "{}\n" {
		t.Fatalf("buffer: %q", string(b))
	}
}

func TestEditTemp_AddsTrailingNewline(t *testing.T) {
	_ = os.Unsetenv("EBO_EDITOR")
	_ = os.Unsetenv("EDITOR")
	if err := os.Setenv("EBO_EDITOR", "true"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("EBO_EDITOR") })

	b, err := EditTemp("{}")
	if err != nil {
		t.Fatalf("EditTemp: %v", err)
	}
	if string(b) != "{}\n" {
		t.Fatalf("buffer: %q", string(b))
	}
}

func TestEditFile_RunErrorIsReturned(t *testing.T) {
	old := execCommand
	t.Cleanup(func() { execCommand = old })
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}

	f, err := os.CreateTemp(t.TempDir(), "x-*.req")
	if err != nil {
		t.Fatalf("temp: %v", err)
	}
	_ = f.Close()

	if _, err := EditFile(f.Name()); err == nil {
		t.Fatalf("expected error")
	}
}
