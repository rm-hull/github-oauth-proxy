package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rm-hull/github-oauth-proxy/internal/config"
	"github.com/rm-hull/github-oauth-proxy/internal/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGithubClient struct {
	mock.Mock
}

func (m *MockGithubClient) ExchangeToken(ctx context.Context, req github.TokenRequest) (*github.TokenResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*github.TokenResponse), args.Error(1)
}

func (m *MockGithubClient) RevokeToken(ctx context.Context, clientID, clientSecret, accessToken string) error {
	args := m.Called(ctx, clientID, clientSecret, accessToken)
	return args.Error(0)
}

func TestServer(t *testing.T) {
	cfg := &config.Config{
		Port:           8080,
		LogLevel:       "info",
		AllowedOrigins: []string{"http://localhost:5173"},
		GithubSecrets: map[string]config.Secret{
			"test_client_id": {
				Name:         "test_app",
				ClientID:     "test_client_id",
				ClientSecret: "test_client_secret",
			},
		},
	}

	mockClient := new(MockGithubClient)
	handlers := NewHandlers(cfg, mockClient)
	server := NewServer(cfg, handlers)

	t.Run("GET /healthz", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/healthz", nil)
		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		// gin-healthcheck returns "[]" if no checks are provided but it's 200 OK
	})

	t.Run("GET /metrics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/metrics", nil)
		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), "requests_total")
	})

	t.Run("POST /v1/github/token - Success", func(t *testing.T) {
		tokenReq := struct {
			ClientID     string `json:"client_id"`
			Code         string `json:"code"`
			CodeVerifier string `json:"code_verifier"`
			RedirectURI  string `json:"redirect_uri"`
		}{
			ClientID:     "test_client_id",
			Code:         "test_code",
			CodeVerifier: "test_verifier",
			RedirectURI:  "http://localhost:5173/callback",
		}

		body, _ := json.Marshal(tokenReq)
		req, _ := http.NewRequest("POST", "/v1/github/token", bytes.NewBuffer(body))
		resp := httptest.NewRecorder()

		mockClient.On("ExchangeToken", mock.Anything, github.TokenRequest{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			Code:         "test_code",
			CodeVerifier: "test_verifier",
			RedirectURI:  "http://localhost:5173/callback",
		}).Return(&github.TokenResponse{
			AccessToken: "test_access_token",
			TokenType:   "bearer",
			Scope:       "repo",
		}, nil)

		server.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		var tokenResp github.TokenResponse
		err := json.Unmarshal(resp.Body.Bytes(), &tokenResp)
		assert.NoError(t, err)
		assert.Equal(t, "test_access_token", tokenResp.AccessToken)
	})

	t.Run("POST /v1/github/token/revoke - Success", func(t *testing.T) {
		revokeReq := struct {
			ClientID    string `json:"client_id"`
			AccessToken string `json:"access_token"`
		}{
			ClientID:    "test_client_id",
			AccessToken: "test_access_token",
		}

		body, _ := json.Marshal(revokeReq)
		req, _ := http.NewRequest("POST", "/v1/github/token/revoke", bytes.NewBuffer(body))
		resp := httptest.NewRecorder()

		mockClient.On("RevokeToken", mock.Anything, "test_client_id", "test_client_secret", "test_access_token").Return(nil)

		server.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		var respBody map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, "Token revoked successfully", respBody["message"])
	})

	t.Run("POST /v1/github/token/revoke - Invalid client_id", func(t *testing.T) {
		revokeReq := struct {
			ClientID    string `json:"client_id"`
			AccessToken string `json:"access_token"`
		}{
			ClientID:    "unknown_client_id",
			AccessToken: "test_access_token",
		}

		body, _ := json.Marshal(revokeReq)
		req, _ := http.NewRequest("POST", "/v1/github/token/revoke", bytes.NewBuffer(body))
		resp := httptest.NewRecorder()

		server.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "Invalid client_id")
	})

	t.Run("POST /v1/github/token - Invalid Origin", func(t *testing.T) {
		tokenReq := struct {
			ClientID     string `json:"client_id"`
			Code         string `json:"code"`
			CodeVerifier string `json:"code_verifier"`
			RedirectURI  string `json:"redirect_uri"`
		}{
			ClientID:     "test_client_id",
			Code:         "test_code",
			CodeVerifier: "test_verifier",
			RedirectURI:  "http://malicious.com/callback",
		}

		body, _ := json.Marshal(tokenReq)
		req, _ := http.NewRequest("POST", "/v1/github/token", bytes.NewBuffer(body))
		resp := httptest.NewRecorder()

		server.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "Invalid redirect_uri")
	})
}
