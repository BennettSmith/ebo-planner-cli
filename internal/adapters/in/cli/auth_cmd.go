package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/BennettSmith/ebo-planner-cli/internal/app/authapp"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/cliopts"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/envelope"
	"github.com/spf13/cobra"
)

func addAuthCommands(root *cobra.Command, deps RootDeps) {
	svc := authapp.Service{Store: deps.ConfigStore}

	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication helpers",
	}

	authCmd.AddCommand(newAuthStatusCmd(deps, svc))
	authCmd.AddCommand(newAuthLogoutCmd(deps, svc))
	authCmd.AddCommand(newAuthTokenCmd(deps, svc))

	root.AddCommand(authCmd)
}

func newAuthStatusCmd(deps RootDeps, svc authapp.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show whether a token is configured for the active profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			st, err := svc.Status(ctx)
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
						"profile":         st.Profile,
						"tokenConfigured": st.TokenConfigured,
						"tokenType":       st.TokenType,
						"expiresAt":       st.ExpiresAt,
					},
					Meta: envelope.Meta{APIURL: resolved.Options.APIURL, Profile: resolved.Options.Profile},
				})
			}
			// Single-line output.
			state := "not configured"
			if st.TokenConfigured {
				state = "configured"
			}
			_, _ = fmt.Fprintf(deps.Stdout, "profile=%s token=%s\n", st.Profile, state)
			return nil
		},
	}
}

func newAuthLogoutCmd(deps RootDeps, svc authapp.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear stored token fields for the active profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if err := svc.Logout(ctx); err != nil {
				return err
			}
			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}
			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: map[string]any{"ok": true},
					Meta: envelope.Meta{APIURL: resolved.Options.APIURL, Profile: resolved.Options.Profile},
				})
			}
			_, _ = io.WriteString(deps.Stdout, "OK\n")
			return nil
		},
	}
}

func newAuthTokenCmd(deps RootDeps, svc authapp.Service) *cobra.Command {
	tokenCmd := &cobra.Command{
		Use:   "token",
		Short: "Manage stored bearer tokens",
	}

	tokenCmd.AddCommand(newAuthTokenSetCmd(deps, svc))
	tokenCmd.AddCommand(newAuthTokenPrintCmd(deps, svc))
	return tokenCmd
}

func newAuthTokenSetCmd(deps RootDeps, svc authapp.Service) *cobra.Command {
	var token string
	cmd := &cobra.Command{
		Use:   "set --token <jwt>",
		Short: "Store a bearer access token for the active profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if err := svc.TokenSet(ctx, "", token); err != nil {
				return err
			}
			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}
			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: map[string]any{"ok": true},
					Meta: envelope.Meta{APIURL: resolved.Options.APIURL, Profile: resolved.Options.Profile},
				})
			}
			_, _ = io.WriteString(deps.Stdout, "OK\n")
			return nil
		},
	}
	cmd.Flags().StringVar(&token, "token", "", "Bearer JWT")
	_ = cmd.MarkFlagRequired("token")
	return cmd
}

func newAuthTokenPrintCmd(deps RootDeps, svc authapp.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "print",
		Short: "Print the current token to stdout (only when explicitly requested)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			tok, _, err := svc.TokenPrint(ctx)
			if err != nil {
				return err
			}
			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}

			// Special-case: must be stdout-only (no extra noise). In JSON mode we still output a single JSON object.
			if resolved.Options.Output == cliopts.OutputJSON {
				b, _ := json.Marshal(map[string]any{"token": tok})
				_, _ = deps.Stdout.Write(append(b, '\n'))
				return nil
			}
			_, _ = io.WriteString(deps.Stdout, tok+"\n")
			return nil
		},
	}
}
