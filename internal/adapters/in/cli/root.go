package cli

import (
	"fmt"
	"io"

	"github.com/BennettSmith/ebo-planner-cli/internal/platform/cliopts"
	"github.com/spf13/cobra"
)

type RootDeps struct {
	Env cliopts.EnvProvider

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

	cmd := &cobra.Command{
		Use:           "ebo",
		Short:         "ebo is a CLI for the Overland Trip Planning service",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := cliopts.ResolveGlobalOptions(cmd.Flags(), deps.Env, defaults)
			if err != nil {
				return err
			}
			if deps.OnResolved != nil {
				deps.OnResolved(resolved)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Root currently has no subcommands (Issue #9 scope). Print help.
			return cmd.Help()
		},
	}

	cmd.SetOut(deps.Stdout)
	cmd.SetErr(deps.Stderr)

	cliopts.AddGlobalFlags(cmd.PersistentFlags(), defaults)

	// Ensure PersistentPreRunE sees persistent flags as well.
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return fmt.Errorf("%w", err)
	})

	return cmd
}
