package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BennettSmith/ebo-planner-cli/internal/platform/specpin"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fatal(err)
	}
	ref, err := specpin.ReadSpecTag(filepath.Join(cwd, "spec.lock"))
	if err != nil {
		fatal(err)
	}

	env := map[string]string{"EBO_SPEC_DIR": os.Getenv("EBO_SPEC_DIR")}
	specDir := specpin.ResolveSpecDir(env, cwd)

	outPath := filepath.Join(cwd, ".spec-cache", ref, "openapi.yaml")
	if err := specpin.ExtractOpenAPIAtRef(specpin.ExecRunner{}, specDir, ref, outPath); err != nil {
		fatal(err)
	}

	fmt.Println(outPath)
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, "ERROR:", err)
	os.Exit(2)
}
