package authloginapp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BennettSmith/ebo-planner-cli/internal/platform/config"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/oidcdevice"
	"github.com/BennettSmith/ebo-planner-cli/internal/ports/out"
)

type BrowserOpener interface {
	Open(url string) error
}

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

type Service struct {
	Store out.ConfigStore
	OIDC  oidcdevice.Client
	Open  BrowserOpener
	Clock Clock
}

type LoginResult struct {
	Profile                 string
	VerificationURI         string
	VerificationURIComplete string
	UserCode                string
	ExpiresAtRFC3339        string
}

func (s Service) Login(ctx context.Context, effective config.Effective) (LoginResult, error) {
	if s.Store == nil {
		return LoginResult{}, exitcode.New(exitcode.KindUnexpected, "config store", fmt.Errorf("nil store"))
	}
	if s.Open == nil {
		return LoginResult{}, exitcode.New(exitcode.KindUnexpected, "browser opener", fmt.Errorf("nil opener"))
	}
	if s.Clock == nil {
		s.Clock = RealClock{}
	}

	doc, err := s.Store.Load(ctx)
	if err != nil {
		return LoginResult{}, exitcode.New(exitcode.KindServer, "load config", err)
	}

	profile := strings.TrimSpace(effective.Profile)
	if profile == "" {
		profile = "default"
	}

	oc, err := config.OIDCOf(doc, profile)
	if err != nil {
		return LoginResult{}, exitcode.New(exitcode.KindUsage,
			"missing OIDC config for profile; set profiles.<name>.oidc.issuerUrl, oidc.clientId, and oidc.scopes",
			err,
		)
	}

	d, err := oidcdevice.Discover(ctx, s.OIDC.HTTP, oc.IssuerURL)
	if err != nil {
		return LoginResult{}, exitcode.New(exitcode.KindServer, "oidc discovery", err)
	}

	dc, err := s.OIDC.RequestDeviceCode(ctx, d.DeviceAuthorizationEndpoint, oc.ClientID, oc.Scopes)
	if err != nil {
		return LoginResult{}, exitcode.New(exitcode.KindServer, "device code", err)
	}

	openURL := dc.VerificationURIComplete
	if openURL == "" {
		openURL = dc.VerificationURI
	}
	_ = s.Open.Open(openURL)

	pollInterval := time.Duration(dc.Interval) * time.Second
	tr, err := s.OIDC.PollToken(ctx, d.TokenEndpoint, oc.ClientID, dc.DeviceCode, pollInterval)
	if err != nil {
		return LoginResult{}, exitcode.New(exitcode.KindAuth, "login failed", err)
	}

	// Persist access token + token type.
	doc, err = config.SetString(doc, "profiles."+profile+".auth.accessToken", tr.AccessToken)
	if err != nil {
		return LoginResult{}, exitcode.New(exitcode.KindServer, "update config", err)
	}
	doc, err = config.SetString(doc, "profiles."+profile+".auth.tokenType", "Bearer")
	if err != nil {
		return LoginResult{}, exitcode.New(exitcode.KindServer, "update config", err)
	}

	expiresAt := ""
	if tr.ExpiresIn > 0 {
		expiresAt = s.Clock.Now().Add(time.Duration(tr.ExpiresIn) * time.Second).UTC().Format(time.RFC3339)
		doc, err = config.SetString(doc, "profiles."+profile+".auth.expiresAt", expiresAt)
		if err != nil {
			return LoginResult{}, exitcode.New(exitcode.KindServer, "update config", err)
		}
	}

	if err := s.Store.Save(ctx, doc); err != nil {
		return LoginResult{}, exitcode.New(exitcode.KindServer, "save config", err)
	}

	return LoginResult{
		Profile:                 profile,
		VerificationURI:         dc.VerificationURI,
		VerificationURIComplete: dc.VerificationURIComplete,
		UserCode:                dc.UserCode,
		ExpiresAtRFC3339:        expiresAt,
	}, nil
}
