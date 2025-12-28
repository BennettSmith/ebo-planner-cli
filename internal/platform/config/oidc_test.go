package config

import "testing"

func TestOIDCOf_MissingScopesFails(t *testing.T) {
	doc := NewEmptyDocument()
	doc, _ = WithProfileAPIURL(doc, "default", "http://x")
	doc, _ = WithProfileOIDC(doc, "default", "https://issuer", "cid", nil)
	_, err := OIDCOf(doc, "default")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestOIDCOf_RoundTrip(t *testing.T) {
	doc := NewEmptyDocument()
	doc, _ = WithProfileAPIURL(doc, "default", "http://x")
	doc, _ = WithProfileOIDC(doc, "default", "https://issuer", "cid", []string{"openid", "email"})
	got, err := OIDCOf(doc, "default")
	if err != nil {
		t.Fatalf("oidc: %v", err)
	}
	if got.IssuerURL != "https://issuer" || got.ClientID != "cid" {
		t.Fatalf("got %#v", got)
	}
	if len(got.Scopes) != 2 || got.Scopes[0] != "openid" {
		t.Fatalf("scopes %#v", got.Scopes)
	}
}

func TestOIDCOf_MissingParts(t *testing.T) {
	cases := []struct {
		name string
		doc  Document
	}{
		{"missing profiles", NewEmptyDocument()},
		{"missing profile", func() Document {
			d := NewEmptyDocument()
			// profiles exists but has no default
			root, _ := rootMapping(d)
			_ = mapEnsureMapping(root, "profiles")
			return d
		}()},
		{"missing oidc", func() Document {
			d := NewEmptyDocument()
			d, _ = WithProfileAPIURL(d, "default", "http://x")
			return d
		}()},
		{"missing issuer", func() Document {
			d := NewEmptyDocument()
			d, _ = WithProfileAPIURL(d, "default", "http://x")
			d, _ = WithProfileOIDC(d, "default", "", "cid", []string{"openid"})
			return d
		}()},
		{"missing clientId", func() Document {
			d := NewEmptyDocument()
			d, _ = WithProfileAPIURL(d, "default", "http://x")
			d, _ = WithProfileOIDC(d, "default", "https://issuer", "", []string{"openid"})
			return d
		}()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := OIDCOf(tc.doc, "default"); err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}
