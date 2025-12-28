package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/BennettSmith/ebo-planner-cli/internal/platform/cliopts"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/config"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
)

type apiContext struct {
	Profile     string
	APIURL      string
	BearerToken string
}

func resolveAPIContext(ctx context.Context, deps RootDeps, resolved cliopts.Resolved) (apiContext, error) {
	if deps.ConfigStore == nil {
		return apiContext{}, exitcode.New(exitcode.KindUnexpected, "config store", fmt.Errorf("nil store"))
	}
	doc, err := deps.ConfigStore.Load(ctx)
	if err != nil {
		return apiContext{}, exitcode.New(exitcode.KindServer, "load config", err)
	}
	view, err := config.ViewOf(doc)
	if err != nil {
		return apiContext{}, exitcode.New(exitcode.KindServer, "parse config", err)
	}

	eff := config.ResolveEffective(resolved, view)
	if strings.TrimSpace(eff.APIURL) == "" {
		return apiContext{}, exitcode.New(
			exitcode.KindUsage,
			fmt.Sprintf("missing api url (set one with `ebo profile set %s --api-url <url>` or pass --api-url)", eff.Profile),
			nil,
		)
	}

	tok, err := config.Get(doc, "profiles."+eff.Profile+".auth.accessToken")
	if err != nil {
		var nf config.ErrNotFound
		if errors.As(err, &nf) {
			return apiContext{}, exitcode.New(exitcode.KindAuth, "no token configured (try `ebo auth login`)", nil)
		}
		return apiContext{}, exitcode.New(exitcode.KindServer, "read token from config", err)
	}
	if strings.TrimSpace(tok) == "" {
		return apiContext{}, exitcode.New(exitcode.KindAuth, "no token configured (try `ebo auth login`)", nil)
	}

	return apiContext{Profile: eff.Profile, APIURL: eff.APIURL, BearerToken: tok}, nil
}

// NOTE: additional shared API helpers belong here as the command surface grows.
