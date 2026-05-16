package cmd

import (
	"os"

	"github.com/rm-hull/github-oauth-proxy/internal/api"
	"github.com/rm-hull/github-oauth-proxy/internal/config"
	"github.com/rm-hull/github-oauth-proxy/internal/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
				log.Fatal().Err(err).Msg("Failed to load config")
			}

			// Override with CLI flags if provided
			if port != 0 {
				cfg.Port = port
			}
			if logLevel != "" {
				cfg.LogLevel = logLevel
			}

			setupLogger(cfg.LogLevel)

			githubClient := github.NewClient()
			handlers := api.NewHandlers(cfg, githubClient)

			log.Info().Int("port", cfg.Port).Msg("Starting server")
			if err := api.Run(cfg, handlers); err != nil {
				log.Fatal().Err(err).Msg("Server failed")
			}
		},
	}

	rootCmd.Flags().IntVarP(&port, "port", "p", 0, "Port to run the server on")
	rootCmd.Flags().StringVarP(&logLevel, "log-level", "l", "", "Log level (debug, info, warn, error)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func setupLogger(level string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	l, err := zerolog.ParseLevel(level)
	if err != nil {
		l = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(l)

	if l == zerolog.DebugLevel {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
