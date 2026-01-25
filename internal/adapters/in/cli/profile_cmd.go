package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/app/profileapp"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/cliopts"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/envelope"
	"github.com/spf13/cobra"
)

func addProfileCommands(root *cobra.Command, deps RootDeps) {
	svc := profileapp.Service{Store: deps.ConfigStore}

	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage named profiles and select the active profile",
	}

	profileCmd.AddCommand(newProfileListCmd(deps, svc))
	profileCmd.AddCommand(newProfileShowCmd(deps, svc))
	profileCmd.AddCommand(newProfileCreateCmd(deps, svc))
	profileCmd.AddCommand(newProfileSetCmd(deps, svc))
	profileCmd.AddCommand(newProfileUseCmd(deps, svc))
	profileCmd.AddCommand(newProfileDeleteCmd(deps, svc))

	root.AddCommand(profileCmd)
}

func newProfileListCmd(deps RootDeps, svc profileapp.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			profiles, current, err := svc.List(ctx)
			if err != nil {
				return err
			}

			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}
			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: map[string]any{
						"currentProfile": current,
						"profiles":       profiles,
					},
					Meta: envelope.Meta{APIURL: resolved.Options.APIURL, Profile: resolved.Options.Profile},
				})
			}

			_, _ = io.WriteString(deps.Stdout, "NAME\tAPI URL\tCURRENT\n")
			for _, p := range profiles {
				mark := ""
				if p.Name == current {
					mark = "*"
				}
				_, _ = fmt.Fprintf(deps.Stdout, "%s\t%s\t%s\n", p.Name, p.APIURL, mark)
			}
			return nil
		},
	}
}

func newProfileShowCmd(deps RootDeps, svc profileapp.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "show [profile]",
		Short: "Show a profile (defaults to current profile)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) == 1 {
				name = args[0]
			}

			ctx := context.Background()
			p, err := svc.Show(ctx, name)
			if err != nil {
				return err
			}

			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}
			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: map[string]any{"profile": p},
					Meta: envelope.Meta{APIURL: resolved.Options.APIURL, Profile: resolved.Options.Profile},
				})
			}

			_, _ = fmt.Fprintf(deps.Stdout, "Name: %s\nAPI URL: %s\n", p.Name, p.APIURL)
			return nil
		},
	}
}

func newProfileCreateCmd(deps RootDeps, svc profileapp.Service) *cobra.Command {
	var apiURL string
	cmd := &cobra.Command{
		Use:   "create <profile> --api-url <url>",
		Short: "Create a new profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			ctx := context.Background()
			if err := svc.Create(ctx, name, apiURL); err != nil {
				return err
			}

			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}
			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: map[string]any{"profile": name, "apiUrl": apiURL},
					Meta: envelope.Meta{APIURL: resolved.Options.APIURL, Profile: resolved.Options.Profile},
				})
			}
			_, _ = io.WriteString(deps.Stdout, "OK\n")
			return nil
		},
	}
	cmd.Flags().StringVar(&apiURL, "api-url", "", "Planner API base URL")
	_ = cmd.MarkFlagRequired("api-url")
	return cmd
}

func newProfileSetCmd(deps RootDeps, svc profileapp.Service) *cobra.Command {
	var apiURL string
	cmd := &cobra.Command{
		Use:   "set <profile> --api-url <url>",
		Short: "Set profile fields (creates profile if missing)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			ctx := context.Background()
			if err := svc.SetAPIURL(ctx, name, apiURL); err != nil {
				return err
			}

			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}
			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: map[string]any{"profile": name, "apiUrl": apiURL},
					Meta: envelope.Meta{APIURL: resolved.Options.APIURL, Profile: resolved.Options.Profile},
				})
			}
			_, _ = io.WriteString(deps.Stdout, "OK\n")
			return nil
		},
	}
	cmd.Flags().StringVar(&apiURL, "api-url", "", "Planner API base URL")
	_ = cmd.MarkFlagRequired("api-url")
	return cmd
}

func newProfileUseCmd(deps RootDeps, svc profileapp.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "use <profile>",
		Short: "Set the current profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			ctx := context.Background()
			if err := svc.Use(ctx, name); err != nil {
				return err
			}

			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}
			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: map[string]any{"currentProfile": name},
					Meta: envelope.Meta{APIURL: resolved.Options.APIURL, Profile: resolved.Options.Profile},
				})
			}
			_, _ = io.WriteString(deps.Stdout, "OK\n")
			return nil
		},
	}
}

func newProfileDeleteCmd(deps RootDeps, svc profileapp.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <profile>",
		Short: "Delete a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			ctx := context.Background()
			if err := svc.Delete(ctx, name); err != nil {
				return err
			}

			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}
			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: map[string]any{"deleted": name},
					Meta: envelope.Meta{APIURL: resolved.Options.APIURL, Profile: resolved.Options.Profile},
				})
			}
			_, _ = io.WriteString(deps.Stdout, "OK\n")
			return nil
		},
	}
}
