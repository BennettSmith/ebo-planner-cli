package main

import (
	"errors"
	"os"
	"strings"

	"github.com/BennettSmith/ebo-planner-cli/internal/adapters/in/cli"
	"github.com/BennettSmith/ebo-planner-cli/internal/adapters/out/configfile"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/cliopts"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/envelope"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
)

func main() {
	env := cliopts.OSEnv{}
	defaults := cliopts.DefaultGlobalOptions()
	peek := cliopts.PeekGlobalOptions(os.Args[1:], env, defaults)

	store := configfile.Store{Env: configfile.OSEnv{}}
	cmd := cli.NewRootCmd(cli.RootDeps{Env: env, ConfigStore: store, Stdout: os.Stdout, Stderr: os.Stderr})
	if err := cmd.Execute(); err != nil {
		// Best-effort classify errors into the required exit code contract.
		mapped := err
		// Cobra/pflag parsing errors don't expose a stable exported type; use a best-effort heuristic.
		if looksLikeUsageError(err) {
			mapped = exitcode.New(exitcode.KindUsage, "usage error", err)
		}

		code := exitcode.Code(mapped)

		if peek.Output == cliopts.OutputJSON {
			_ = envelope.WriteJSON(os.Stdout, envelope.Envelope{
				Meta: envelope.Meta{APIURL: peek.APIURL, Profile: peek.Profile},
				Error: &envelope.ErrorBody{
					Code:    stringExitCodeKind(mapped),
					Message: mapped.Error(),
				},
			})
		} else {
			_, _ = os.Stderr.WriteString(mapped.Error() + "\n")
		}

		os.Exit(code)
	}
}

func stringExitCodeKind(err error) string {
	var e *exitcode.Error
	if errors.As(err, &e) {
		return string(e.Kind)
	}
	return string(exitcode.KindUnexpected)
}

func looksLikeUsageError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "unknown flag") ||
		strings.Contains(msg, "flag needs an argument") ||
		strings.Contains(msg, "requires an argument") ||
		strings.Contains(msg, "invalid argument")
}
