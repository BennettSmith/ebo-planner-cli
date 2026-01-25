package profileapp

import (
	"context"
	"fmt"
	"sort"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/config"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/ports/out"
)

type Service struct {
	Store out.ConfigStore
}

type ProfileSummary struct {
	Name   string
	APIURL string
}

func (s Service) List(ctx context.Context) ([]ProfileSummary, string, error) {
	doc, err := s.Store.Load(ctx)
	if err != nil {
		return nil, "", exitcode.New(exitcode.KindServer, "load config", err)
	}
	v, err := config.ViewOf(doc)
	if err != nil {
		return nil, "", exitcode.New(exitcode.KindServer, "parse config", err)
	}
	current := v.CurrentProfile
	if current == "" {
		current = "default"
	}

	outList := make([]ProfileSummary, 0, len(v.Profiles))
	for name, p := range v.Profiles {
		outList = append(outList, ProfileSummary{Name: name, APIURL: p.APIURL})
	}
	sort.Slice(outList, func(i, j int) bool { return outList[i].Name < outList[j].Name })
	return outList, current, nil
}

func (s Service) Show(ctx context.Context, name string) (ProfileSummary, error) {
	doc, err := s.Store.Load(ctx)
	if err != nil {
		return ProfileSummary{}, exitcode.New(exitcode.KindServer, "load config", err)
	}
	v, err := config.ViewOf(doc)
	if err != nil {
		return ProfileSummary{}, exitcode.New(exitcode.KindServer, "parse config", err)
	}

	if name == "" {
		name = v.CurrentProfile
		if name == "" {
			name = "default"
		}
	}
	p, ok := v.Profiles[name]
	if !ok {
		return ProfileSummary{}, exitcode.New(exitcode.KindUsage, "profile does not exist", fmt.Errorf("%s", name))
	}
	return ProfileSummary{Name: name, APIURL: p.APIURL}, nil
}

func (s Service) Create(ctx context.Context, name, apiURL string) error {
	doc, err := s.Store.Load(ctx)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "load config", err)
	}
	v, err := config.ViewOf(doc)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "parse config", err)
	}
	if _, ok := v.Profiles[name]; ok {
		return exitcode.New(exitcode.KindConflict, "profile already exists", fmt.Errorf("%s", name))
	}
	doc, err = config.WithProfileAPIURL(doc, name, apiURL)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "update config", err)
	}
	if err := s.Store.Save(ctx, doc); err != nil {
		return exitcode.New(exitcode.KindServer, "save config", err)
	}
	return nil
}

func (s Service) SetAPIURL(ctx context.Context, name, apiURL string) error {
	doc, err := s.Store.Load(ctx)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "load config", err)
	}
	doc, err = config.WithProfileAPIURL(doc, name, apiURL)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "update config", err)
	}
	if err := s.Store.Save(ctx, doc); err != nil {
		return exitcode.New(exitcode.KindServer, "save config", err)
	}
	return nil
}

func (s Service) Use(ctx context.Context, name string) error {
	doc, err := s.Store.Load(ctx)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "load config", err)
	}
	v, err := config.ViewOf(doc)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "parse config", err)
	}
	if _, ok := v.Profiles[name]; !ok {
		return exitcode.New(exitcode.KindUsage, "profile does not exist", fmt.Errorf("%s", name))
	}
	doc, err = config.WithCurrentProfile(doc, name)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "update config", err)
	}
	if err := s.Store.Save(ctx, doc); err != nil {
		return exitcode.New(exitcode.KindServer, "save config", err)
	}
	return nil
}

func (s Service) Delete(ctx context.Context, name string) error {
	doc, err := s.Store.Load(ctx)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "load config", err)
	}
	v, err := config.ViewOf(doc)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "parse config", err)
	}
	if _, ok := v.Profiles[name]; !ok {
		return exitcode.New(exitcode.KindConflict, "profile does not exist", fmt.Errorf("%s", name))
	}

	// Cannot delete last profile.
	if len(v.Profiles) <= 1 {
		return exitcode.New(exitcode.KindConflict, "cannot delete last profile; create another profile first", nil)
	}

	// Remove profiles.<name>
	doc, err = config.Unset(doc, "profiles."+name)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "update config", err)
	}

	// If deleting current profile, switch to default if it exists.
	current := v.CurrentProfile
	if current == "" {
		current = "default"
	}
	if current == name {
		if _, ok := v.Profiles["default"]; ok && name != "default" {
			doc, err = config.WithCurrentProfile(doc, "default")
			if err != nil {
				return exitcode.New(exitcode.KindServer, "update config", err)
			}
		} else {
			return exitcode.New(exitcode.KindConflict, "cannot delete current profile; create a default profile (or switch currentProfile first)", nil)
		}
	}

	if err := s.Store.Save(ctx, doc); err != nil {
		return exitcode.New(exitcode.KindServer, "save config", err)
	}
	return nil
}
