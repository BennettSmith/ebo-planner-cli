package browseropen

import (
	"os/exec"
	"testing"
)

func TestDefaultOpener_EmptyURLIsError(t *testing.T) {
	if err := (DefaultOpener{}).Open(""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestDefaultOpener_CommandSelection(t *testing.T) {
	oldGOOS := goos
	oldExec := execCommand
	t.Cleanup(func() {
		goos = oldGOOS
		execCommand = oldExec
	})

	type tc struct {
		goos     string
		wantName string
		wantArgs []string
	}
	cases := []tc{
		{goos: "darwin", wantName: "open", wantArgs: []string{"http://example"}},
		{goos: "windows", wantName: "rundll32", wantArgs: []string{"url.dll,FileProtocolHandler", "http://example"}},
		{goos: "linux", wantName: "xdg-open", wantArgs: []string{"http://example"}},
	}

	for _, c := range cases {
		goos = c.goos

		var gotName string
		var gotArgs []string
		execCommand = func(name string, args ...string) *exec.Cmd {
			gotName = name
			gotArgs = append([]string{}, args...)
			// Use a harmless command so tests don't actually open a browser.
			return exec.Command("true")
		}

		if err := (DefaultOpener{}).Open("http://example"); err != nil {
			t.Fatalf("goos=%s: err=%v", c.goos, err)
		}
		if gotName != c.wantName {
			t.Fatalf("goos=%s: name got %q want %q", c.goos, gotName, c.wantName)
		}
		if len(gotArgs) != len(c.wantArgs) {
			t.Fatalf("goos=%s: args got %#v want %#v", c.goos, gotArgs, c.wantArgs)
		}
		for i := range gotArgs {
			if gotArgs[i] != c.wantArgs[i] {
				t.Fatalf("goos=%s: args got %#v want %#v", c.goos, gotArgs, c.wantArgs)
			}
		}
	}
}
