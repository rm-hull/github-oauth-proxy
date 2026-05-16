package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExchangeToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedResp := TokenResponse{
			AccessToken: "test_token",
			TokenType:   "bearer",
			Scope:       "repo",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			json.NewEncoder(w).Encode(expectedResp)
		}))
		defer server.Close()

		c := &client{
			httpClient: server.Client(),
			baseURL:    server.URL,
		}

		resp, err := c.ExchangeToken(context.Background(), TokenRequest{})
		assert.NoError(t, err)
		assert.Equal(t, expectedResp.AccessToken, resp.AccessToken)
	})

	t.Run("API Error Response", func(t *testing.T) {
		errorResp := TokenResponse{
			Error:            "bad_verification_code",
			ErrorDescription: "The code passed is incorrect",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResp)
		}))
		defer server.Close()

		c := &client{
			httpClient: server.Client(),
			baseURL:    server.URL,
		}

		resp, err := c.ExchangeToken(context.Background(), TokenRequest{})
		assert.NoError(t, err)
		assert.Equal(t, errorResp.Error, resp.Error)
		assert.Equal(t, errorResp.ErrorDescription, resp.ErrorDescription)
	})

	t.Run("Server Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := &client{
			httpClient: server.Client(),
			baseURL:    server.URL,
		}

		resp, err := c.ExchangeToken(context.Background(), TokenRequest{})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
