package requestfile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadStrict reads the file at path and decodes it into out.
//
// Format detection rules:
//   - .json => JSON
//   - .yaml/.yml => YAML
//   - otherwise: attempt JSON first, then YAML; if both fail, returns an error
//
// Decoding is strict: unknown fields are rejected.
func LoadStrict(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return decodeJSONStrict(b, out)
	case ".yaml", ".yml":
		return decodeYAMLAsJSONStrict(b, out)
	default:
		jsonErr := decodeJSONStrict(b, out)
		if jsonErr == nil {
			return nil
		}
		yamlErr := decodeYAMLAsJSONStrict(b, out)
		if yamlErr == nil {
			return nil
		}
		return fmt.Errorf("parse %s: JSON: %v; YAML: %v", path, jsonErr, yamlErr)
	}
}

func decodeJSONStrict(b []byte, out any) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return err
	}
	// Ensure decoder is at EOF (covers trailing whitespace, rejects extra tokens).
	if err := dec.Decode(&struct{}{}); err == nil {
		return errors.New("unexpected trailing JSON content")
	} else if !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func decodeYAMLAsJSONStrict(b []byte, out any) error {
	var v any
	if err := yaml.Unmarshal(b, &v); err != nil {
		return err
	}
	norm, err := normalizeYAML(v)
	if err != nil {
		return err
	}
	j, err := json.Marshal(norm)
	if err != nil {
		return err
	}
	return decodeJSONStrict(j, out)
}

func normalizeYAML(v any) (any, error) {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, vv := range x {
			n, err := normalizeYAML(vv)
			if err != nil {
				return nil, err
			}
			out[k] = n
		}
		return out, nil
	case map[any]any:
		out := make(map[string]any, len(x))
		for k, vv := range x {
			ks, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("yaml key must be string, got %T", k)
			}
			n, err := normalizeYAML(vv)
			if err != nil {
				return nil, err
			}
			out[ks] = n
		}
		return out, nil
	case []any:
		out := make([]any, 0, len(x))
		for _, vv := range x {
			n, err := normalizeYAML(vv)
			if err != nil {
				return nil, err
			}
			out = append(out, n)
		}
		return out, nil
	default:
		return v, nil
	}
}
