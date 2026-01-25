package specpin

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const defaultSiblingSpecDir = "../trip-planner-spec"

type Runner interface {
	Run(name string, args ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, bytes.TrimSpace(out))
	}
	return out, nil
}

// ReadSpecTag reads the pinned spec tag from spec.lock.
func ReadSpecTag(specLockPath string) (string, error) {
	b, err := os.ReadFile(specLockPath)
	if err != nil {
		return "", err
	}
	tag := strings.TrimSpace(string(b))
	if tag == "" {
		return "", fmt.Errorf("spec.lock is empty")
	}
	return tag, nil
}

// ResolveSpecDir resolves the spec repo directory.
//
// - If EBO_SPEC_DIR is set, it is used.
// - Otherwise, default to a sibling checkout at ../trip-planner-spec.
func ResolveSpecDir(env map[string]string, cwd string) string {
	if v := strings.TrimSpace(env["EBO_SPEC_DIR"]); v != "" {
		if filepath.IsAbs(v) {
			return v
		}
		return filepath.Clean(filepath.Join(cwd, v))
	}
	return filepath.Clean(filepath.Join(cwd, defaultSiblingSpecDir))
}

// VerifyTagExists ensures the tag exists in the spec repo.
func VerifyRefExists(r Runner, specDir, tag string) error {
	if r == nil {
		r = ExecRunner{}
	}
	if _, err := r.Run("git", "-C", specDir, "rev-parse", "--verify", tag); err != nil {
		return fmt.Errorf("spec repo missing ref %q (from spec.lock). Try: git -C %s fetch --tags (if using tags)\n%w", tag, specDir, err)
	}
	return nil
}

// ExtractOpenAPIAtRef writes openapi/openapi.yaml at the given git ref to destPath.
//
// This avoids requiring the spec repo working tree to be checked out to the tag.
func ExtractOpenAPIAtRef(r Runner, specDir, tag, destPath string) error {
	if r == nil {
		r = ExecRunner{}
	}
	if err := VerifyRefExists(r, specDir, tag); err != nil {
		return err
	}

	blob := fmt.Sprintf("%s:openapi/openapi.yaml", tag)
	b, err := r.Run("git", "-C", specDir, "show", blob)
	if err != nil {
		return fmt.Errorf("failed to read %s from spec repo: %w", blob, err)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destPath, b, 0o644)
}
