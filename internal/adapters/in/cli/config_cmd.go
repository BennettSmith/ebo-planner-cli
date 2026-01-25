package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/app/configapp"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/cliopts"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/envelope"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
	"github.com/spf13/cobra"
)

func addConfigCommands(root *cobra.Command, deps RootDeps) {
	svc := configapp.Service{Store: deps.ConfigStore}

	cfgCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}

	cfgCmd.AddCommand(newConfigPathCmd(deps, svc))
	cfgCmd.AddCommand(newConfigGetCmd(deps, svc))
	cfgCmd.AddCommand(newConfigSetCmd(deps, svc))
	cfgCmd.AddCommand(newConfigUnsetCmd(deps, svc))
	cfgCmd.AddCommand(newConfigListCmd(deps, svc))

	root.AddCommand(cfgCmd)
}

func resolvedFromRoot(cmd *cobra.Command, deps RootDeps) (cliopts.Resolved, error) {
	defaults := cliopts.DefaultGlobalOptions()
	return cliopts.ResolveGlobalOptions(cmd.InheritedFlags(), deps.Env, defaults)
}

func newConfigPathCmd(deps RootDeps, svc configapp.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := svc.EnsureStore(); err != nil {
				return exitcode.New(exitcode.KindUnexpected, "config store", err)
			}
			ctx := context.Background()
			p, err := svc.Path(ctx)
			if err != nil {
				return err
			}

			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}
			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: map[string]any{"path": p},
					Meta: envelope.Meta{APIURL: resolved.Options.APIURL, Profile: resolved.Options.Profile},
				})
			}
			_, _ = io.WriteString(deps.Stdout, p+"\n")
			return nil
		},
	}
}

func newConfigGetCmd(deps RootDeps, svc configapp.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a config value by dot-path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			ctx := context.Background()
			val, err := svc.Get(ctx, key)
			if err != nil {
				return err
			}

			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}
			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: map[string]any{"key": key, "value": val},
					Meta: envelope.Meta{APIURL: resolved.Options.APIURL, Profile: resolved.Options.Profile},
				})
			}
			_, _ = io.WriteString(deps.Stdout, val+"\n")
			return nil
		},
	}
}

func newConfigSetCmd(deps RootDeps, svc configapp.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value by dot-path",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			key, val := args[0], args[1]
			if err := svc.Set(ctx, key, val); err != nil {
				return err
			}

			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}
			if resolved.Options.Output == cliopts.OutputJSON {
				outVal := val
				if strings.Contains(key, "accessToken") {
					outVal = "REDACTED"
				}
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: map[string]any{"key": key, "value": outVal},
					Meta: envelope.Meta{APIURL: resolved.Options.APIURL, Profile: resolved.Options.Profile},
				})
			}
			_, _ = io.WriteString(deps.Stdout, "OK\n")
			return nil
		},
	}
}

func newConfigUnsetCmd(deps RootDeps, svc configapp.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "unset <key>",
		Short: "Remove a config value by dot-path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			key := args[0]
			if err := svc.Unset(ctx, key); err != nil {
				return err
			}

			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}
			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: map[string]any{"key": key},
					Meta: envelope.Meta{APIURL: resolved.Options.APIURL, Profile: resolved.Options.Profile},
				})
			}
			_, _ = io.WriteString(deps.Stdout, "OK\n")
			return nil
		},
	}
}

func newConfigListCmd(deps RootDeps, svc configapp.Service) *cobra.Command {
	var includeSecrets bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print the entire config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}

			if resolved.Options.Output == cliopts.OutputJSON {
				data, err := svc.ListJSON(ctx, includeSecrets)
				if err != nil {
					return err
				}
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: data,
					Meta: envelope.Meta{APIURL: resolved.Options.APIURL, Profile: resolved.Options.Profile},
				})
			}

			// Table mode always redacts.
			y, err := svc.ListYAML(ctx, false)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprint(deps.Stdout, y)
			return nil
		},
	}
	cmd.Flags().BoolVar(&includeSecrets, "include-secrets", false, "Include secrets in JSON output only")
	return cmd
}
