package authapp

import (
	"context"
	"strings"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/config"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/ports/out"
)

type Service struct {
	Store out.ConfigStore
}

type Status struct {
	Profile         string
	TokenConfigured bool
	TokenType       string
	ExpiresAt       string
}

func (s Service) Status(ctx context.Context) (Status, error) {
	doc, err := s.Store.Load(ctx)
	if err != nil {
		return Status{}, exitcode.New(exitcode.KindServer, "load config", err)
	}
	v, err := config.ViewOf(doc)
	if err != nil {
		return Status{}, exitcode.New(exitcode.KindServer, "parse config", err)
	}
	profile := v.CurrentProfile
	if profile == "" {
		profile = "default"
	}

	accessToken, _ := config.Get(doc, "profiles."+profile+".auth.accessToken")
	tokenType, _ := config.Get(doc, "profiles."+profile+".auth.tokenType")
	expiresAt, _ := config.Get(doc, "profiles."+profile+".auth.expiresAt")

	configured := strings.TrimSpace(accessToken) != ""
	st := Status{Profile: profile, TokenConfigured: configured, TokenType: tokenType, ExpiresAt: expiresAt}
	if !configured {
		return st, exitcode.New(exitcode.KindAuth, "no token configured", nil)
	}
	return st, nil
}

func (s Service) Logout(ctx context.Context) error {
	doc, err := s.Store.Load(ctx)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "load config", err)
	}
	v, err := config.ViewOf(doc)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "parse config", err)
	}
	profile := v.CurrentProfile
	if profile == "" {
		profile = "default"
	}

	accessToken, _ := config.Get(doc, "profiles."+profile+".auth.accessToken")
	if strings.TrimSpace(accessToken) == "" {
		return exitcode.New(exitcode.KindAuth, "no token configured", nil)
	}

	doc, err = config.Unset(doc, "profiles."+profile+".auth.accessToken")
	if err != nil {
		return exitcode.New(exitcode.KindServer, "update config", err)
	}
	doc, err = config.Unset(doc, "profiles."+profile+".auth.tokenType")
	if err != nil {
		return exitcode.New(exitcode.KindServer, "update config", err)
	}
	doc, err = config.Unset(doc, "profiles."+profile+".auth.expiresAt")
	if err != nil {
		return exitcode.New(exitcode.KindServer, "update config", err)
	}

	if err := s.Store.Save(ctx, doc); err != nil {
		return exitcode.New(exitcode.KindServer, "save config", err)
	}
	return nil
}

func (s Service) TokenSet(ctx context.Context, profile string, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return exitcode.New(exitcode.KindUsage, "empty token", nil)
	}
	if !looksLikeJWT(token) {
		return exitcode.New(exitcode.KindUsage, "token must look like a JWT (three dot-separated parts)", nil)
	}

	doc, err := s.Store.Load(ctx)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "load config", err)
	}

	if profile == "" {
		v, err := config.ViewOf(doc)
		if err != nil {
			return exitcode.New(exitcode.KindServer, "parse config", err)
		}
		profile = v.CurrentProfile
		if profile == "" {
			profile = "default"
		}
	}

	doc, err = config.SetString(doc, "profiles."+profile+".auth.accessToken", token)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "update config", err)
	}
	doc, err = config.SetString(doc, "profiles."+profile+".auth.tokenType", "Bearer")
	if err != nil {
		return exitcode.New(exitcode.KindServer, "update config", err)
	}

	if err := s.Store.Save(ctx, doc); err != nil {
		return exitcode.New(exitcode.KindServer, "save config", err)
	}
	return nil
}

func (s Service) TokenPrint(ctx context.Context) (string, string, error) {
	doc, err := s.Store.Load(ctx)
	if err != nil {
		return "", "", exitcode.New(exitcode.KindServer, "load config", err)
	}
	v, err := config.ViewOf(doc)
	if err != nil {
		return "", "", exitcode.New(exitcode.KindServer, "parse config", err)
	}
	profile := v.CurrentProfile
	if profile == "" {
		profile = "default"
	}

	token, err := config.Get(doc, "profiles."+profile+".auth.accessToken")
	if err != nil || strings.TrimSpace(token) == "" {
		return "", profile, exitcode.New(exitcode.KindAuth, "no token configured", nil)
	}
	return token, profile, nil
}

func looksLikeJWT(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
	}
	return true
}
