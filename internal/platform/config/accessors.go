package config

import (
	"gopkg.in/yaml.v3"
)

func ViewOf(doc Document) (View, error) {
	root, err := rootMapping(doc)
	if err != nil {
		return View{}, err
	}

	v := View{Profiles: map[string]ProfileView{}}
	if cp := mapGet(root, "currentProfile"); cp != nil && cp.Kind == yaml.ScalarNode {
		v.CurrentProfile = cp.Value
	}

	profiles := mapGet(root, "profiles")
	if profiles != nil && profiles.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(profiles.Content); i += 2 {
			k := profiles.Content[i]
			pv := profiles.Content[i+1]
			if k.Kind != yaml.ScalarNode || pv.Kind != yaml.MappingNode {
				continue
			}
			p := ProfileView{}
			if au := mapGet(pv, "apiUrl"); au != nil && au.Kind == yaml.ScalarNode {
				p.APIURL = au.Value
			}
			v.Profiles[k.Value] = p
		}
	}

	return v, nil
}

func WithCurrentProfile(doc Document, profile string) (Document, error) {
	root, err := rootMapping(doc)
	if err != nil {
		return Document{}, err
	}
	mapSetScalar(root, "currentProfile", profile)
	return doc, nil
}

func WithProfileAPIURL(doc Document, profile, apiURL string) (Document, error) {
	root, err := rootMapping(doc)
	if err != nil {
		return Document{}, err
	}
	profiles := mapEnsureMapping(root, "profiles")
	pnode := mapEnsureMapping(profiles, profile)
	mapSetScalar(pnode, "apiUrl", apiURL)
	return doc, nil
}
