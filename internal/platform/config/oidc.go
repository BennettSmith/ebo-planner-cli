package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type OIDCConfig struct {
	IssuerURL string
	ClientID  string
	Scopes    []string
}

func OIDCOf(doc Document, profile string) (OIDCConfig, error) {
	root, err := rootMapping(doc)
	if err != nil {
		return OIDCConfig{}, err
	}
	profiles := mapGet(root, "profiles")
	if profiles == nil || profiles.Kind != yaml.MappingNode {
		return OIDCConfig{}, fmt.Errorf("missing profiles")
	}
	pnode := mapGet(profiles, profile)
	if pnode == nil || pnode.Kind != yaml.MappingNode {
		return OIDCConfig{}, fmt.Errorf("missing profile %q", profile)
	}
	oidc := mapGet(pnode, "oidc")
	if oidc == nil || oidc.Kind != yaml.MappingNode {
		return OIDCConfig{}, fmt.Errorf("missing oidc config")
	}
	issuer := mapGet(oidc, "issuerUrl")
	clientID := mapGet(oidc, "clientId")
	scopes := mapGet(oidc, "scopes")
	if issuer == nil || issuer.Kind != yaml.ScalarNode || issuer.Value == "" {
		return OIDCConfig{}, fmt.Errorf("missing oidc.issuerUrl")
	}
	if clientID == nil || clientID.Kind != yaml.ScalarNode || clientID.Value == "" {
		return OIDCConfig{}, fmt.Errorf("missing oidc.clientId")
	}
	if scopes == nil || scopes.Kind != yaml.SequenceNode || len(scopes.Content) == 0 {
		return OIDCConfig{}, fmt.Errorf("missing oidc.scopes")
	}
	outScopes := make([]string, 0, len(scopes.Content))
	for _, n := range scopes.Content {
		if n.Kind != yaml.ScalarNode || n.Value == "" {
			continue
		}
		outScopes = append(outScopes, n.Value)
	}
	if len(outScopes) == 0 {
		return OIDCConfig{}, fmt.Errorf("missing oidc.scopes")
	}
	return OIDCConfig{IssuerURL: issuer.Value, ClientID: clientID.Value, Scopes: outScopes}, nil
}

func WithProfileOIDC(doc Document, profile string, issuerURL, clientID string, scopes []string) (Document, error) {
	root, err := rootMapping(doc)
	if err != nil {
		return Document{}, err
	}
	profiles := mapEnsureMapping(root, "profiles")
	pnode := mapEnsureMapping(profiles, profile)
	oidc := mapEnsureMapping(pnode, "oidc")
	mapSetScalar(oidc, "issuerUrl", issuerURL)
	mapSetScalar(oidc, "clientId", clientID)

	seq := &yaml.Node{Kind: yaml.SequenceNode}
	for _, s := range scopes {
		seq.Content = append(seq.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: s})
	}
	mapSetNode(oidc, "scopes", seq)
	return doc, nil
}
