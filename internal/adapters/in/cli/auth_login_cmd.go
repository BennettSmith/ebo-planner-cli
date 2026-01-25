package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/app/authloginapp"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/browseropen"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/cliopts"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/config"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/envelope"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/oidcdevice"
	"github.com/spf13/cobra"
)

func newAuthLoginCmd(deps RootDeps) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Interactive login (OIDC device flow)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if deps.ConfigStore == nil {
				return exitcode.New(exitcode.KindUnexpected, "config store", fmt.Errorf("nil store"))
			}

			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}

			doc, err := deps.ConfigStore.Load(ctx)
			if err != nil {
				return exitcode.New(exitcode.KindServer, "load config", err)
			}
			view, err := config.ViewOf(doc)
			if err != nil {
				return exitcode.New(exitcode.KindServer, "parse config", err)
			}
			eff := config.ResolveEffective(resolved, view)

			// Polling timeout default: 5 minutes unless user overrides via --timeout.
			totalTimeout := resolved.Options.Timeout
			if resolved.Sources["timeout"] == "default" {
				totalTimeout = 5 * time.Minute
			}
			loginCtx, cancel := context.WithTimeout(ctx, totalTimeout)
			defer cancel()

			client := oidcdevice.Client{HTTP: &http.Client{}}
			opener := deps.BrowserOpener
			if opener == nil {
				opener = browseropen.DefaultOpener{}
			}
			svc := authloginapp.Service{Store: deps.ConfigStore, OIDC: client, Open: opener}
			res, err := svc.Login(loginCtx, eff)
			if err != nil {
				return err
			}

			verify := res.VerificationURIComplete
			if verify == "" {
				verify = res.VerificationURI
			}

			// Print guidance to stderr.
			_, _ = fmt.Fprintf(deps.Stderr, "Open: %s\nCode: %s\n", verify, res.UserCode)

			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: map[string]any{"ok": true, "profile": res.Profile, "expiresAt": res.ExpiresAtRFC3339},
					Meta: envelope.Meta{APIURL: eff.APIURL, Profile: eff.Profile},
				})
			}

			_, _ = io.WriteString(deps.Stdout, "OK\n")
			return nil
		},
	}
}
