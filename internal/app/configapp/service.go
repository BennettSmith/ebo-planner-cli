package configapp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/config"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/ports/out"
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

	// Special-case: OIDC scopes must be a YAML array. `ebo config set` takes a string input,
	// but we allow providing the array as a JSON string array (recommended) or a comma/space list.
	if strings.HasSuffix(strings.TrimSpace(key), ".oidc.scopes") {
		raw := strings.TrimSpace(value)
		var scopes []string
		if strings.HasPrefix(raw, "[") {
			if err := json.Unmarshal([]byte(raw), &scopes); err != nil {
				return exitcode.New(exitcode.KindUsage, "invalid oidc scopes (expected JSON string array like [\"openid\",\"profile\"])", err)
			}
		} else if strings.Contains(raw, ",") {
			for _, p := range strings.Split(raw, ",") {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				scopes = append(scopes, p)
			}
		} else {
			for _, p := range strings.Fields(raw) {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				scopes = append(scopes, p)
			}
		}
		if len(scopes) == 0 {
			return exitcode.New(exitcode.KindUsage, "invalid oidc scopes (must include at least one scope, e.g. openid)", nil)
		}
		doc, err = config.SetStringList(doc, key, scopes)
	} else {
		doc, err = config.SetString(doc, key, value)
	}
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
