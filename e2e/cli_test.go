package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

var eboPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "ebo-e2e-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmp)

	bin := filepath.Join(tmp, "ebo")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", bin, "../cmd/ebo")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}

	eboPath = bin
	os.Exit(m.Run())
}

type runResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func runEBO(t *testing.T, env map[string]string, args ...string) runResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, eboPath, args...)
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	code := 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
		} else {
			code = 1
		}
	}
	return runResult{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: code}
}

func TestAuthTokenCommands_E2E(t *testing.T) {
	cfgDir := t.TempDir()
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(api.Close)

	env := map[string]string{
		"EBO_CONFIG_DIR": cfgDir,
		"EBO_NO_COLOR":   "1",
		"EBO_API_URL":    api.URL,
	}

	// status (no token) -> exit 3
	res := runEBO(t, env, "auth", "status")
	if res.ExitCode != 3 {
		t.Fatalf("status exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}

	// token set
	res = runEBO(t, env, "auth", "token", "set", "--token", "a.b.c")
	if res.ExitCode != 0 {
		t.Fatalf("token set exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
	if strings.Contains(res.Stdout, "a.b.c") || strings.Contains(res.Stderr, "a.b.c") {
		t.Fatalf("token leaked in set output")
	}

	// token status json
	res = runEBO(t, env, "--output", "json", "auth", "status")
	if res.ExitCode != 0 {
		t.Fatalf("status json exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(res.Stdout), &got); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, res.Stdout)
	}
	data := got["data"].(map[string]any)
	if data["tokenConfigured"] != true {
		t.Fatalf("tokenConfigured=%#v", data["tokenConfigured"])
	}

	// token print (table) -> stdout only, no stderr
	res = runEBO(t, env, "auth", "token", "print")
	if res.ExitCode != 0 {
		t.Fatalf("print exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
	if strings.TrimSpace(res.Stdout) != "a.b.c" {
		t.Fatalf("stdout=%q", res.Stdout)
	}
	if res.Stderr != "" {
		t.Fatalf("expected no stderr, got %q", res.Stderr)
	}

	// token print (json) -> simple json object
	res = runEBO(t, env, "--output", "json", "auth", "token", "print")
	if res.ExitCode != 0 {
		t.Fatalf("print json exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
	if res.Stderr != "" {
		t.Fatalf("expected no stderr, got %q", res.Stderr)
	}
	var tokObj map[string]any
	if err := json.Unmarshal([]byte(res.Stdout), &tokObj); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, res.Stdout)
	}
	if tokObj["token"] != "a.b.c" {
		t.Fatalf("token=%#v", tokObj["token"])
	}

	// logout clears
	res = runEBO(t, env, "auth", "logout")
	if res.ExitCode != 0 {
		t.Fatalf("logout exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
	if strings.Contains(res.Stdout, "a.b.c") || strings.Contains(res.Stderr, "a.b.c") {
		t.Fatalf("token leaked in logout output")
	}

	// print missing -> exit 3, token not printed
	res = runEBO(t, env, "auth", "token", "print")
	if res.ExitCode != 3 {
		t.Fatalf("print missing exit=%d", res.ExitCode)
	}
	if strings.Contains(res.Stdout, "a.b.c") || strings.Contains(res.Stderr, "a.b.c") {
		t.Fatalf("token leaked when missing")
	}
}
