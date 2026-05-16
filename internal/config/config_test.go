package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	t.Run("Success with valid config", func(t *testing.T) {
		os.Setenv("PORT", "8080")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000")
		os.Setenv("GITHUB_OAUTH_SECRET_APP1", "id1|secret1")
		defer os.Clearenv()

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, 8080, cfg.Port)
		assert.Equal(t, "debug", cfg.LogLevel)
		assert.Contains(t, cfg.AllowedOrigins, "http://localhost:3000")
		assert.Equal(t, "id1", cfg.GithubSecrets["id1"].ClientID)
		assert.Equal(t, "secret1", cfg.GithubSecrets["id1"].ClientSecret)
	})

	t.Run("Success with multiple secrets", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("GITHUB_OAUTH_SECRET_APP1", "id1|secret1")
		os.Setenv("GITHUB_OAUTH_SECRET_APP2", "id2|secret2")
		defer os.Clearenv()

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Len(t, cfg.GithubSecrets, 2)
		assert.Equal(t, "secret1", cfg.GithubSecrets["id1"].ClientSecret)
		assert.Equal(t, "secret2", cfg.GithubSecrets["id2"].ClientSecret)
	})

	t.Run("Missing secrets", func(t *testing.T) {
		os.Clearenv()
		_, err := Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing GitHub OAuth credentials")
	})

	t.Run("Invalid secret format", func(t *testing.T) {
		os.Setenv("GITHUB_OAUTH_SECRET_APP1", "invalid_format")
		defer os.Clearenv()
		_, err := Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid secret format")
	})

	t.Run("IsOriginAllowed", func(t *testing.T) {
		cfg := &Config{
			AllowedOrigins: []string{"http://localhost:3000", "https://example.com"},
		}
		assert.True(t, cfg.IsOriginAllowed("http://localhost:3000"))
		assert.True(t, cfg.IsOriginAllowed("https://example.com"))
		assert.False(t, cfg.IsOriginAllowed("http://malicious.com"))
	})
}
