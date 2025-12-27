package config

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ErrNotFound indicates a missing key path.
type ErrNotFound struct{ Key string }

func (e ErrNotFound) Error() string { return fmt.Sprintf("key not found: %s", e.Key) }

func splitPath(key string) ([]string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, fmt.Errorf("empty key")
	}
	parts := strings.Split(key, ".")
	for _, p := range parts {
		if p == "" {
			return nil, fmt.Errorf("invalid key %q", key)
		}
	}
	return parts, nil
}

// Get returns the scalar value at a dot-path key.
// If the key is missing, ErrNotFound is returned.
func Get(doc Document, key string) (string, error) {
	parts, err := splitPath(key)
	if err != nil {
		return "", err
	}
	root, err := rootMapping(doc)
	if err != nil {
		return "", err
	}

	n := root
	for i, p := range parts {
		if n.Kind != yaml.MappingNode {
			return "", ErrNotFound{Key: key}
		}
		v := mapGet(n, p)
		if v == nil {
			return "", ErrNotFound{Key: key}
		}
		if i == len(parts)-1 {
			if v.Kind != yaml.ScalarNode {
				return "", fmt.Errorf("key %s is not a scalar", key)
			}
			return v.Value, nil
		}
		n = v
	}
	return "", ErrNotFound{Key: key}
}

// SetString sets a scalar string value at a dot-path key, creating missing maps.
func SetString(doc Document, key string, value string) (Document, error) {
	parts, err := splitPath(key)
	if err != nil {
		return Document{}, err
	}
	root, err := rootMapping(doc)
	if err != nil {
		return Document{}, err
	}

	n := root
	for i, p := range parts {
		if i == len(parts)-1 {
			mapSetScalar(n, p, value)
			return doc, nil
		}
		n = mapEnsureMapping(n, p)
	}
	return doc, nil
}

// Unset removes the key at the dot-path.
// If the key doesn't exist, it is a no-op.
func Unset(doc Document, key string) (Document, error) {
	parts, err := splitPath(key)
	if err != nil {
		return Document{}, err
	}
	root, err := rootMapping(doc)
	if err != nil {
		return Document{}, err
	}
	_ = unsetAt(root, parts)
	return doc, nil
}

func unsetAt(m *yaml.Node, parts []string) bool {
	if m == nil || m.Kind != yaml.MappingNode || len(parts) == 0 {
		return false
	}
	key := parts[0]
	for i := 0; i+1 < len(m.Content); i += 2 {
		k := m.Content[i]
		v := m.Content[i+1]
		if k.Kind == yaml.ScalarNode && k.Value == key {
			if len(parts) == 1 {
				// remove key/value
				m.Content = append(m.Content[:i], m.Content[i+2:]...)
				return true
			}
			// recurse
			removed := unsetAt(v, parts[1:])
			return removed
		}
	}
	return false
}

func IsSecretKey(key string) bool {
	// Spec-defined secret: profiles.<name>.auth.accessToken
	parts, err := splitPath(key)
	if err != nil {
		return false
	}
	if len(parts) >= 4 && parts[0] == "profiles" && parts[2] == "auth" && parts[3] == "accessToken" {
		return true
	}
	return false
}

// RedactSecrets returns a copy of doc where known secret values are replaced with "REDACTED".
func RedactSecrets(doc Document) (Document, error) {
	_, err := rootMapping(doc)
	if err != nil {
		return Document{}, err
	}

	copyRoot := deepCopyNode(doc.Root)
	out := Document{Root: copyRoot}

	outRoot, _ := rootMapping(out)
	// Walk known secret paths: profiles.*.auth.accessToken
	profiles := mapGet(outRoot, "profiles")
	if profiles == nil || profiles.Kind != yaml.MappingNode {
		return out, nil
	}
	for i := 0; i+1 < len(profiles.Content); i += 2 {
		pnode := profiles.Content[i+1]
		auth := mapGet(pnode, "auth")
		if auth == nil || auth.Kind != yaml.MappingNode {
			continue
		}
		at := mapGet(auth, "accessToken")
		if at != nil && at.Kind == yaml.ScalarNode {
			at.Value = "REDACTED"
		}
	}
	return out, nil
}

func deepCopyNode(n *yaml.Node) *yaml.Node {
	if n == nil {
		return nil
	}
	c := *n
	if len(n.Content) > 0 {
		c.Content = make([]*yaml.Node, len(n.Content))
		for i := range n.Content {
			c.Content[i] = deepCopyNode(n.Content[i])
		}
	}
	return &c
}
