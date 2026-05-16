package api

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/rm-hull/github-oauth-proxy/internal/config"
	"github.com/rm-hull/github-oauth-proxy/internal/github"
	"github.com/rs/zerolog/log"
)

type Handlers struct {
	cfg          *config.Config
	githubClient github.Client
}

func NewHandlers(cfg *config.Config, githubClient github.Client) *Handlers {
	return &Handlers{
		cfg:          cfg,
		githubClient: githubClient,
	}
}

func (h *Handlers) ExchangeToken(c *gin.Context) {
	var req struct {
		ClientID     string `json:"client_id" binding:"required"`
		Code         string `json:"code" binding:"required"`
		CodeVerifier string `json:"code_verifier" binding:"required"`
		RedirectURI  string `json:"redirect_uri" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid parameters"})
		return
	}

	// Validate Redirect URI Origin
	u, err := url.Parse(req.RedirectURI)
	if err != nil {
		log.Warn().Str("redirect_uri", req.RedirectURI).Msg("Invalid redirect_uri format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid redirect_uri format"})
		return
	}

	origin := u.Scheme + "://" + u.Host
	if !h.cfg.IsOriginAllowed(origin) {
		log.Warn().Str("redirect_uri", req.RedirectURI).Str("origin", origin).Msg("Invalid redirect_uri origin")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid redirect_uri"})
		return
	}

	// Lookup Secret
	secret, ok := h.cfg.GithubSecrets[req.ClientID]
	if !ok {
		log.Warn().Str("client_id", req.ClientID).Msg("Unknown client_id received")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client_id"})
		return
	}

	log.Info().
		Str("app", secret.Name).
		Str("clientId", req.ClientID).
		Str("redirectUri", req.RedirectURI).
		Str("ip", c.ClientIP()).
		Msg("Exchanging code for token")

	resp, err := h.githubClient.ExchangeToken(c.Request.Context(), github.TokenRequest{
		ClientID:     secret.ClientID,
		ClientSecret: secret.ClientSecret,
		Code:         req.Code,
		CodeVerifier: req.CodeVerifier,
		RedirectURI:  req.RedirectURI,
	})

	if err != nil {
		log.Error().Err(err).Msg("Error exchanging code for token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if resp.Error != "" {
		log.Error().
			Str("error", resp.Error).
			Str("description", resp.ErrorDescription).
			Msg("GitHub OAuth error")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	log.Info().
		Str("scope", resp.Scope).
		Str("tokenType", resp.TokenType).
		Msg("Token exchange successful")

	c.JSON(http.StatusOK, resp)
}
