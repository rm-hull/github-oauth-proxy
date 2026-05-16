package config

import (
	"slices"
	"os"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/joho/godotenv"
)

type Secret struct {
	Name         string
	ClientID     string
	ClientSecret string
}

type Config struct {
	Port           int
	LogLevel       string
	AllowedOrigins []string
	GithubSecrets  map[string]Secret
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "" {
		allowedOrigins = []string{"http://localhost:5173"}
	}

	secrets := make(map[string]Secret)
	prefix := "GITHUB_OAUTH_SECRET_"

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		key := pair[0]
		value := pair[1]

		if strings.HasPrefix(key, prefix) && value != "" {
			name := strings.TrimPrefix(key, prefix)
			parts := strings.Split(value, "|")
			if len(parts) != 2 {
				return nil, errors.Errorf("invalid secret format for %s. Expected 'clientId|clientSecret'", key)
			}
			clientID := parts[0]
			clientSecret := parts[1]
			secrets[clientID] = Secret{
				Name:         name,
				ClientID:     clientID,
				ClientSecret: clientSecret,
			}
		}
	}

	if len(secrets) == 0 {
		return nil, errors.New("missing GitHub OAuth credentials")
	}

	return &Config{
		Port:           port,
		LogLevel:       logLevel,
		AllowedOrigins: allowedOrigins,
		GithubSecrets:  secrets,
	}, nil
}

func (c *Config) IsOriginAllowed(origin string) bool {
	return slices.Contains(c.AllowedOrigins, origin)
}
