package cli

import (
	"bytes"
	"fmt"
	"io"

	"github.com/BennettSmith/ebo-planner-cli/internal/platform/cliopts"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/envelope"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
	"github.com/BennettSmith/ebo-planner-cli/internal/ports/out"
	"github.com/spf13/cobra"
)

type RootDeps struct {
	Env cliopts.EnvProvider

	ConfigStore out.ConfigStore

	Stdout io.Writer
	Stderr io.Writer

	// OnResolved is a test hook invoked after flags/env are resolved.
	OnResolved func(cliopts.Resolved)
}

func NewRootCmd(deps RootDeps) *cobra.Command {
	if deps.Env == nil {
		deps.Env = cliopts.OSEnv{}
	}

	defaults := cliopts.DefaultGlobalOptions()
	var resolved cliopts.Resolved

	cmd := &cobra.Command{
		Use:           "ebo",
		Short:         "ebo is a CLI for the Overland Trip Planning service",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			r, err := cliopts.ResolveGlobalOptions(cmd.Flags(), deps.Env, defaults)
			if err != nil {
				return exitcode.New(exitcode.KindUsage, "invalid flags", err)
			}
			resolved = r
			if deps.OnResolved != nil {
				deps.OnResolved(r)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Root currently has no subcommands (Issue #9 scope). Print help.
			if resolved.Options.Output == cliopts.OutputJSON {
				b := &bytes.Buffer{}
				cmd.SetOut(b)
				_ = cmd.Help()
				cmd.SetOut(deps.Stdout)

				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: map[string]any{
						"help": b.String(),
					},
					Meta: envelope.Meta{APIURL: resolved.Options.APIURL, Profile: resolved.Options.Profile},
				})
			}
			return cmd.Help()
		},
	}

	cmd.SetOut(deps.Stdout)
	cmd.SetErr(deps.Stderr)

	cliopts.AddGlobalFlags(cmd.PersistentFlags(), defaults)

	// Ensure PersistentPreRunE sees persistent flags as well.
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return fmt.Errorf("%w", exitcode.New(exitcode.KindUsage, "usage error", err))
	})

	addConfigCommands(cmd, deps)
	addProfileCommands(cmd, deps)
	addAuthCommands(cmd, deps)

	return cmd
}
