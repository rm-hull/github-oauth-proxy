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
}

func NewClient() Client {
	return &client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *client) ExchangeToken(ctx context.Context, req TokenRequest) (*TokenResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal token request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute http request")
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // Limit to 1MB
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal token response")
	}

	return &tokenResp, nil
}
