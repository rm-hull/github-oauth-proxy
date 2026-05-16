package api

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/rm-hull/github-oauth-proxy/internal/config"
	"github.com/rm-hull/github-oauth-proxy/internal/github"
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
		slog.Warn("Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid parameters"})
		return
	}

	// Validate Redirect URI Origin
	u, err := url.Parse(req.RedirectURI)
	if err != nil {
		slog.Warn("Invalid redirect_uri format", "redirect_uri", req.RedirectURI)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid redirect_uri format"})
		return
	}

	origin := u.Scheme + "://" + u.Host
	if !h.cfg.IsOriginAllowed(origin) {
		slog.Warn("Invalid redirect_uri origin", "redirect_uri", req.RedirectURI, "origin", origin)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid redirect_uri"})
		return
	}

	// Lookup Secret
	secret, ok := h.cfg.GithubSecrets[req.ClientID]
	if !ok {
		slog.Warn("Unknown client_id received", "client_id", req.ClientID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client_id"})
		return
	}

	slog.Info("Exchanging code for token",
		"app", secret.Name,
		"clientId", req.ClientID,
		"redirectUri", req.RedirectURI,
		"ip", c.ClientIP(),
	)

	resp, err := h.githubClient.ExchangeToken(c.Request.Context(), github.TokenRequest{
		ClientID:     secret.ClientID,
		ClientSecret: secret.ClientSecret,
		Code:         req.Code,
		CodeVerifier: req.CodeVerifier,
		RedirectURI:  req.RedirectURI,
	})

	if err != nil {
		slog.Error("Error exchanging code for token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if resp.Error != "" {
		slog.Error("GitHub OAuth error", "error", resp.Error, "description", resp.ErrorDescription)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	slog.Info("Token exchange successful",
		"scope", resp.Scope,
		"tokenType", resp.TokenType,
	)

	c.JSON(http.StatusOK, resp)
}
