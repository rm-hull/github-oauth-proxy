package github

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/cockroachdb/errors"
)

const (
	defaultWebURL = "https://github.com"
	defaultAPIURL = "https://api.github.com"

	// Use a version you have verified against GitHub's "API versions" docs.
	apiVersion = "2026-03-10"
	userAgent  = "github-oauth-proxy (https://github.com/rm-hull/github-oauth-proxy)"

	maxErrorBodyBytes = 1 << 15 // 32KB
	maxBodyBytes      = 1 << 20 // 1MB
)

type TokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	RedirectURI  string `json:"redirect_uri"`
}

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	Scope            string `json:"scope"`
	TokenType        string `json:"token_type"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

type Client interface {
	ExchangeToken(ctx context.Context, req TokenRequest) (*TokenResponse, error)
	RevokeToken(ctx context.Context, clientID, clientSecret, accessToken string) error
}

type client struct {
	httpClient *http.Client
	webURL     string // OAuth web flow endpoints (github.com)
	apiURL     string // REST API endpoints (api.github.com)
}

func NewClient() Client {
	return &client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		webURL:     defaultWebURL,
		apiURL:     defaultAPIURL,
	}
}

// RevokeToken revokes a single OAuth access token. It is idempotent: a 404
// from GitHub (token already revoked or invalid) is treated as success.
func (c *client) RevokeToken(ctx context.Context, clientID, clientSecret, accessToken string) error {
	body, err := json.Marshal(map[string]string{"access_token": accessToken})
	if err != nil {
		return errors.Wrap(err, "failed to marshal revoke request")
	}

	endpoint := c.apiURL + "/applications/" + url.PathEscape(clientID) + "/token"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "failed to create http request")
	}

	httpReq.SetBasicAuth(clientID, clientSecret)
	httpReq.Header.Set("Accept", "application/vnd.github+json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-GitHub-Api-Version", apiVersion)
	httpReq.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return errors.Wrap(err, "failed to execute http request")
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes))

	switch resp.StatusCode {
	case http.StatusNoContent, http.StatusOK:
		return nil
	case http.StatusNotFound:
		slog.DebugContext(ctx, "GitHub revoke returned 404; token already revoked or invalid")
		return nil
	default:
		return errors.Errorf("GitHub revoke API returned status: %s, body: %s", resp.Status, string(respBody))
	}
}

// ExchangeToken exchanges an authorization code for an access token.
//
// GitHub frequently returns HTTP 200 with an "error" field set (for example
// bad_verification_code), so callers must check TokenResponse.Error.
func (c *client) ExchangeToken(ctx context.Context, req TokenRequest) (*TokenResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal token request")
	}

	endpoint := c.webURL + "/login/oauth/access_token"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute http request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		var errResp TokenResponse
		decodeErr := json.NewDecoder(io.LimitReader(resp.Body, maxErrorBodyBytes)).Decode(&errResp)
		if decodeErr == nil && errResp.Error != "" {
			return &errResp, nil
		}
		return nil, errors.Errorf("unexpected HTTP response: %s", resp.Status)
	}

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		slog.ErrorContext(ctx, "failed to parse GitHub token response", "error", err)
		return &TokenResponse{Error: "internal_error", ErrorDescription: "failed to parse response"}, nil
	}

	// A 200 with neither an error nor a token is not a success.
	if tokenResp.Error == "" && tokenResp.AccessToken == "" {
		return nil, errors.New("GitHub token response contained neither an access token nor an error")
	}

	return &tokenResp, nil
}