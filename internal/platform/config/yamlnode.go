package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func NewEmptyDocument() Document {
	doc := &yaml.Node{Kind: yaml.DocumentNode}
	root := &yaml.Node{Kind: yaml.MappingNode}
	doc.Content = []*yaml.Node{root}
	return Document{Root: doc}
}

func rootMapping(doc Document) (*yaml.Node, error) {
	if doc.Root == nil || doc.Root.Kind != yaml.DocumentNode || len(doc.Root.Content) == 0 {
		return nil, fmt.Errorf("invalid yaml document")
	}
	root := doc.Root.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("invalid yaml root: expected mapping")
	}
	return root, nil
}

func mapGet(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		k := m.Content[i]
		v := m.Content[i+1]
		if k.Kind == yaml.ScalarNode && k.Value == key {
			return v
		}
	}
	return nil
}

func mapSetScalar(m *yaml.Node, key, value string) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		k := m.Content[i]
		if k.Kind == yaml.ScalarNode && k.Value == key {
			m.Content[i+1] = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
			return
		}
	}
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value},
	)
}

func mapEnsureMapping(m *yaml.Node, key string) *yaml.Node {
	if v := mapGet(m, key); v != nil {
		if v.Kind == yaml.MappingNode {
			return v
		}
		// Overwrite non-mapping with mapping (known key path).
		newMap := &yaml.Node{Kind: yaml.MappingNode}
		mapSetNode(m, key, newMap)
		return newMap
	}
	newMap := &yaml.Node{Kind: yaml.MappingNode}
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		newMap,
	)
	return newMap
}

func mapSetNode(m *yaml.Node, key string, node *yaml.Node) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		k := m.Content[i]
		if k.Kind == yaml.ScalarNode && k.Value == key {
			m.Content[i+1] = node
			return
		}
	}
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		node,
	)
}
