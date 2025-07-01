package config

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func NewLogger(settings *Settings) zerolog.Logger {
	var output io.Writer

	var logLevel zerolog.Level

	env := strings.ToLower(settings.Environment)
	switch env {
	case "production":
		output = os.Stdout
		log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()

		logLevel = zerolog.ErrorLevel
		zerolog.SetGlobalLevel(logLevel)

	case "test-environment":
		output = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.Out = os.Stderr
			w.TimeFormat = time.RFC3339
		})

		log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()

		logLevel = zerolog.InfoLevel
		zerolog.SetGlobalLevel(logLevel)

	case "local":
		output = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.Out = os.Stderr
			w.TimeFormat = time.RFC3339
		})

		log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()

		log.Warn().Msgf("Unknown log level: '%s'. Selected default level: 'info'", settings.LogLevel)
		logLevel, err := zerolog.ParseLevel(strings.ToLower(settings.LogLevel))
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to read logging level")
		}

		zerolog.SetGlobalLevel(logLevel)
	}

	log.Info().Str("logLevel", logLevel.String()).Msg("Logger successfully initialized")
	return log.Logger
}
