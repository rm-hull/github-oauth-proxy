package api

import (
	"fmt"
	"time"

	"github.com/Depado/ginprom"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rm-hull/github-oauth-proxy/internal/config"
	"github.com/rm-hull/github-oauth-proxy/internal/middleware"
	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/tavsec/gin-healthcheck/checks"
	hc_config "github.com/tavsec/gin-healthcheck/config"
)

func NewServer(cfg *config.Config, handlers *Handlers) (*gin.Engine, error) {
	if cfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	if len(cfg.TrustedProxies) > 0 {
		err := r.SetTrustedProxies(cfg.TrustedProxies)
		if err != nil {
			return nil, fmt.Errorf("failed to set trusted proxies: %w", err)
		}
	}
	r.Use(gin.Recovery())
	r.Use(middleware.Logger("/healthz", "/metrics"))
	r.Use(callersMiddleware("/healthz", "/metrics"))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"POST", "GET", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	p := ginprom.New(
		ginprom.Engine(r),
		ginprom.Path("/metrics"),
		ginprom.Ignore("/healthz"),
	)
	r.Use(p.Instrument())

	_ = healthcheck.New(r, hc_config.DefaultConfig(), []checks.Check{})

	v1 := r.Group("/v1")
	v1.POST("/github/token", handlers.ExchangeToken)
	v1.POST("/github/token/revoke", handlers.RevokeToken)

	return r, nil
}

func Run(cfg *config.Config, handlers *Handlers) error {
	r, err := NewServer(cfg, handlers)
	if err != nil {
		return err
	}
	return r.Run(fmt.Sprintf(":%d", cfg.Port))
}
