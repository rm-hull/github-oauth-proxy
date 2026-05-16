package github

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
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
}

type client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient() Client {
	return &client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://github.com",
	}
}

func (c *client) ExchangeToken(ctx context.Context, req TokenRequest) (*TokenResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal token request")
	}

	url := c.baseURL + "/login/oauth/access_token"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "github-oauth-proxy (https://github.com/rm-hull/github-oauth-proxy)")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute http request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		var errResp TokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return &errResp, nil
		}
		return nil, errors.Errorf("unexpected HTTP response: %s", resp.Status)
	}

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // Limit to 1MB
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return &TokenResponse{Error: "internal_error", ErrorDescription: "failed to parse response"}, nil
	}

	return &tokenResp, nil
}
