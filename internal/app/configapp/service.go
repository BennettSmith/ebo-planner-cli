package configapp

import (
	"context"
	"fmt"

	"github.com/BennettSmith/ebo-planner-cli/internal/platform/config"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
	"github.com/BennettSmith/ebo-planner-cli/internal/ports/out"
)

type Service struct {
	Store out.ConfigStore
}

func (s Service) Path(ctx context.Context) (string, error) {
	p, err := s.Store.Path(ctx)
	if err != nil {
		return "", exitcode.New(exitcode.KindServer, "config path", err)
	}
	return p, nil
}

func (s Service) Get(ctx context.Context, key string) (string, error) {
	doc, err := s.Store.Load(ctx)
	if err != nil {
		return "", exitcode.New(exitcode.KindServer, "load config", err)
	}
	v, err := config.Get(doc, key)
	if err != nil {
		if _, ok := err.(config.ErrNotFound); ok {
			return "", exitcode.New(exitcode.KindNotFound, "config key not found", err)
		}
		return "", exitcode.New(exitcode.KindUsage, "invalid config key", err)
	}
	if config.IsSecretKey(key) {
		return "REDACTED", nil
	}
	return v, nil
}

func (s Service) Set(ctx context.Context, key, value string) error {
	doc, err := s.Store.Load(ctx)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "load config", err)
	}
	doc, err = config.SetString(doc, key, value)
	if err != nil {
		return exitcode.New(exitcode.KindUsage, "invalid config key", err)
	}
	if err := s.Store.Save(ctx, doc); err != nil {
		return exitcode.New(exitcode.KindServer, "save config", err)
	}
	return nil
}

func (s Service) Unset(ctx context.Context, key string) error {
	doc, err := s.Store.Load(ctx)
	if err != nil {
		return exitcode.New(exitcode.KindServer, "load config", err)
	}
	doc, err = config.Unset(doc, key)
	if err != nil {
		return exitcode.New(exitcode.KindUsage, "invalid config key", err)
	}
	if err := s.Store.Save(ctx, doc); err != nil {
		return exitcode.New(exitcode.KindServer, "save config", err)
	}
	return nil
}

func (s Service) ListYAML(ctx context.Context, includeSecrets bool) (string, error) {
	doc, err := s.Store.Load(ctx)
	if err != nil {
		return "", exitcode.New(exitcode.KindServer, "load config", err)
	}
	if !includeSecrets {
		doc, err = config.RedactSecrets(doc)
		if err != nil {
			return "", exitcode.New(exitcode.KindServer, "redact secrets", err)
		}
	}
	b, err := config.MarshalYAML(doc)
	if err != nil {
		return "", exitcode.New(exitcode.KindServer, "marshal config", err)
	}
	return string(b), nil
}

func (s Service) ListJSON(ctx context.Context, includeSecrets bool) (any, error) {
	doc, err := s.Store.Load(ctx)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "load config", err)
	}
	if !includeSecrets {
		doc, err = config.RedactSecrets(doc)
		if err != nil {
			return nil, exitcode.New(exitcode.KindServer, "redact secrets", err)
		}
	}
	m, err := config.ToInterface(doc)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "marshal config", err)
	}
	return m, nil
}

func (s Service) EnsureStore() error {
	if s.Store == nil {
		return fmt.Errorf("nil store")
	}
	return nil
}
