package oidcdevice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Sleeper interface {
	Sleep(ctx context.Context, d time.Duration) error
}

type RealSleeper struct{}

func (RealSleeper) Sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

type Discovery struct {
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
	TokenEndpoint               string `json:"token_endpoint"`
}

type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type TokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e TokenError) IsPending() bool  { return e.Error == "authorization_pending" }
func (e TokenError) IsSlowDown() bool { return e.Error == "slow_down" }

type Client struct {
	HTTP    HTTPDoer
	Sleeper Sleeper
}

func (c Client) ensure() error {
	if c.HTTP == nil {
		return fmt.Errorf("nil http client")
	}
	if c.Sleeper == nil {
		c.Sleeper = RealSleeper{}
	}
	return nil
}

func Discover(ctx context.Context, httpc HTTPDoer, issuerURL string) (Discovery, error) {
	issuerURL = strings.TrimRight(strings.TrimSpace(issuerURL), "/")
	if issuerURL == "" {
		return Discovery{}, fmt.Errorf("empty issuer url")
	}
	u := issuerURL + "/.well-known/openid-configuration"
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := httpc.Do(req)
	if err != nil {
		return Discovery{}, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return Discovery{}, fmt.Errorf("discovery http %d: %s", resp.StatusCode, string(b))
	}
	var d Discovery
	if err := json.Unmarshal(b, &d); err != nil {
		return Discovery{}, err
	}
	if d.DeviceAuthorizationEndpoint == "" || d.TokenEndpoint == "" {
		return Discovery{}, fmt.Errorf("discovery missing required endpoints")
	}
	return d, nil
}

func (c Client) RequestDeviceCode(ctx context.Context, deviceEndpoint, clientID string, scopes []string) (DeviceCodeResponse, error) {
	if err := c.ensure(); err != nil {
		return DeviceCodeResponse{}, err
	}
	form := url.Values{}
	form.Set("client_id", clientID)
	if len(scopes) > 0 {
		form.Set("scope", strings.Join(scopes, " "))
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, deviceEndpoint, strings.NewReader(form.Encode()))
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return DeviceCodeResponse{}, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return DeviceCodeResponse{}, fmt.Errorf("device code http %d: %s", resp.StatusCode, string(b))
	}
	var out DeviceCodeResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return DeviceCodeResponse{}, err
	}
	if out.DeviceCode == "" || out.UserCode == "" || out.VerificationURI == "" {
		return DeviceCodeResponse{}, fmt.Errorf("device code response missing required fields")
	}
	return out, nil
}

func (c Client) PollToken(ctx context.Context, tokenEndpoint, clientID, deviceCode string, interval time.Duration) (TokenResponse, error) {
	if err := c.ensure(); err != nil {
		return TokenResponse{}, err
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}
	for {
		select {
		case <-ctx.Done():
			return TokenResponse{}, ctx.Err()
		default:
		}

		form := url.Values{}
		form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
		form.Set("device_code", deviceCode)
		form.Set("client_id", clientID)
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
		req.Header.Set("content-type", "application/x-www-form-urlencoded")
		resp, err := c.HTTP.Do(req)
		if err != nil {
			return TokenResponse{}, err
		}
		b, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode/100 == 2 {
			var tr TokenResponse
			if err := json.Unmarshal(b, &tr); err != nil {
				return TokenResponse{}, err
			}
			if tr.AccessToken == "" {
				return TokenResponse{}, fmt.Errorf("token response missing access_token")
			}
			return tr, nil
		}

		// error response
		var te TokenError
		if err := json.Unmarshal(b, &te); err != nil || te.Error == "" {
			return TokenResponse{}, fmt.Errorf("token http %d: %s", resp.StatusCode, string(b))
		}
		if te.IsPending() {
			if err := c.Sleeper.Sleep(ctx, interval); err != nil {
				return TokenResponse{}, err
			}
			continue
		}
		if te.IsSlowDown() {
			interval += 5 * time.Second
			if err := c.Sleeper.Sleep(ctx, interval); err != nil {
				return TokenResponse{}, err
			}
			continue
		}
		return TokenResponse{}, errors.New(te.Error)
	}
}
