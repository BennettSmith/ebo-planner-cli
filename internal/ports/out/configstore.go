package out

import (
	"context"

	"github.com/BennettSmith/ebo-planner-cli/internal/platform/config"
)

// ConfigStore persists and loads the CLI configuration.
//
// It must preserve unknown YAML fields when saving an existing config.
//
// See docs/cli-spec.md "Config and profiles".
// See docs/architecture.md for layering rules.
type ConfigStore interface {
	Path(ctx context.Context) (string, error)
	Load(ctx context.Context) (config.Document, error)
	Save(ctx context.Context, doc config.Document) error
}
