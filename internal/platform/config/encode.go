package config

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

func MarshalYAML(doc Document) ([]byte, error) {
	return yaml.Marshal(doc.Root)
}

func ToInterface(doc Document) (any, error) {
	// Decode yaml.Node into interface{} for JSON output.
	var v any
	b, err := MarshalYAML(doc)
	if err != nil {
		return nil, err
	}
	dec := yaml.NewDecoder(bytes.NewReader(b))
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}
