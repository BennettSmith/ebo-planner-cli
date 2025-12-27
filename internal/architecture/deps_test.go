package architecture_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"testing"
)

type goListPkg struct {
	ImportPath string   `json:"ImportPath"`
	Imports    []string `json:"Imports"`
	Standard   bool     `json:"Standard"`
}

func TestArchitecture_DependencyInvariants(t *testing.T) {
	pkgs := mustGoList(t)

	var violations []string
	for _, p := range pkgs {
		if !strings.Contains(p.ImportPath, "/internal/") {
			continue
		}

		for _, imp := range p.Imports {
			// Only enforce invariants within this repo's internal packages.
			if !strings.Contains(imp, "/internal/") {
				continue
			}

			// Rules are *direct* import rules.
			// (Transitive imports are allowed; this matches how Go enforces imports.)
			if strings.Contains(p.ImportPath, "/internal/app/") || strings.HasSuffix(p.ImportPath, "/internal/app") {
				if strings.Contains(imp, "/internal/adapters/") {
					violations = append(violations, fmt.Sprintf("%s must not import %s (app must not depend on adapters)", p.ImportPath, imp))
				}
				if strings.Contains(imp, "/internal/gen/") {
					violations = append(violations, fmt.Sprintf("%s must not import %s (app must not depend on generated code)", p.ImportPath, imp))
				}
			}

			if strings.Contains(p.ImportPath, "/internal/platform/") || strings.HasSuffix(p.ImportPath, "/internal/platform") {
				if strings.Contains(imp, "/internal/app/") || strings.HasSuffix(imp, "/internal/app") {
					violations = append(violations, fmt.Sprintf("%s must not import %s (platform must not depend on app)", p.ImportPath, imp))
				}
				if strings.Contains(imp, "/internal/adapters/") {
					violations = append(violations, fmt.Sprintf("%s must not import %s (platform must not depend on adapters)", p.ImportPath, imp))
				}
				if strings.Contains(imp, "/internal/gen/") {
					violations = append(violations, fmt.Sprintf("%s must not import %s (platform must not depend on generated code)", p.ImportPath, imp))
				}
			}

			if strings.Contains(p.ImportPath, "/internal/adapters/in/") {
				if strings.Contains(imp, "/internal/gen/") {
					violations = append(violations, fmt.Sprintf("%s must not import %s (inbound adapter must not depend on generated code)", p.ImportPath, imp))
				}
			}

			if strings.Contains(p.ImportPath, "/internal/adapters/out/") {
				if strings.Contains(imp, "/internal/adapters/in/") {
					violations = append(violations, fmt.Sprintf("%s must not import %s (outbound adapters must not depend on inbound adapters)", p.ImportPath, imp))
				}
			}
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("architecture dependency violations:\n- %s", strings.Join(violations, "\n- "))
	}
}

func mustGoList(t *testing.T) []goListPkg {
	t.Helper()

	cmd := exec.Command("go", "list", "-json", "./...")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list failed: %v\n%s", err, string(out))
	}

	dec := json.NewDecoder(bytes.NewReader(out))
	var pkgs []goListPkg
	for dec.More() {
		var p goListPkg
		if err := dec.Decode(&p); err != nil {
			t.Fatalf("decode go list json: %v", err)
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}
