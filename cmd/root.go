package cmd

import (
	"log/slog"
	"os"

	"github.com/rm-hull/github-oauth-proxy/internal/api"
	"github.com/rm-hull/github-oauth-proxy/internal/config"
	"github.com/rm-hull/github-oauth-proxy/internal/github"
	"github.com/rm-hull/godx"
	"github.com/spf13/cobra"
)

func Execute() {
	var port int
	var logLevel string

	rootCmd := &cobra.Command{
		Use:   "github-oauth-proxy",
		Short: "A secure intermediary for GitHub OAuth authentication",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load()
			if err != nil {
				slog.Error("Failed to load config", "error", err)
				os.Exit(1)
			}

			// Override with CLI flags if provided
			if port != 0 {
				cfg.Port = port
			}
			if logLevel != "" {
				cfg.LogLevel = logLevel
			}

			logger := setupLogger(cfg.LogLevel)
			godx.Diagnostics(logger)

			githubClient := github.NewClient()
			handlers := api.NewHandlers(cfg, githubClient)

			slog.Info("Starting server", "port", cfg.Port)
			if err := api.Run(cfg, handlers); err != nil {
				slog.Error("Server failed", "error", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.Flags().IntVarP(&port, "port", "p", 0, "Port to run the server on")
	rootCmd.Flags().StringVarP(&logLevel, "log-level", "l", "", "Log level (debug, info, warn, error)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func setupLogger(level string) *slog.Logger {
	var programLevel = new(slog.LevelVar)
	switch level {
	case "debug":
		programLevel.Set(slog.LevelDebug)
	case "warn":
		programLevel.Set(slog.LevelWarn)
	case "error":
		programLevel.Set(slog.LevelError)
	default:
		programLevel.Set(slog.LevelInfo)
	}

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: programLevel,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
