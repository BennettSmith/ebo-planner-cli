package main

import (
	"os"

	"github.com/BennettSmith/ebo-planner-cli/internal/adapters/in/cli"
)

func main() {
	cmd := cli.NewRootCmd(cli.RootDeps{Env: nil, Stdout: os.Stdout, Stderr: os.Stderr})
	if err := cmd.Execute(); err != nil {
		// Exit code mapping is implemented in Issue #10; use 1 for now.
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
