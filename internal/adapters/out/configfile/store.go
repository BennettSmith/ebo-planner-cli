package configfile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/config"
	"gopkg.in/yaml.v3"
)

type Env interface {
	LookupEnv(key string) (string, bool)
}

type OSEnv struct{}

func (OSEnv) LookupEnv(key string) (string, bool) { return os.LookupEnv(key) }

type Store struct {
	Env Env
}

func (s Store) Path(ctx context.Context) (string, error) {
	_ = ctx
	base, err := s.configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "ebo", "config.yaml"), nil
}

func (s Store) Load(ctx context.Context) (config.Document, error) {
	path, err := s.Path(ctx)
	if err != nil {
		return config.Document{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return config.NewEmptyDocument(), nil
		}
		return config.Document{}, err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(b, &doc); err != nil {
		return config.Document{}, err
	}
	// Ensure document shape.
	if doc.Kind == 0 {
		return config.NewEmptyDocument(), nil
	}
	return config.Document{Root: &doc}, nil
}

func (s Store) Save(ctx context.Context, doc config.Document) error {
	path, err := s.Path(ctx)
	if err != nil {
		return err
	}
	if doc.Root == nil {
		return fmt.Errorf("nil document")
	}

	// Ensure parent dir exists.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	b, err := yaml.Marshal(doc.Root)
	if err != nil {
		return err
	}

	// Write atomically with restrictive permissions.
	tmp, err := os.CreateTemp(filepath.Dir(path), "config-*.yaml")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func (s Store) configDir() (string, error) {
	if s.Env == nil {
		s.Env = OSEnv{}
	}
	if v, ok := s.Env.LookupEnv("EBO_CONFIG_DIR"); ok && v != "" {
		return v, nil
	}
	return os.UserConfigDir()
}
