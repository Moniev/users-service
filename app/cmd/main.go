package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"users-service/app/config"
	_ "users-service/app/docs"

	"github.com/swaggo/swag"
)

var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "users-service.users-service.svc.cluster.local:8000",
	BasePath:         "/",
	Schemes:          []string{"http", "https"},
	Title:            "Users Service API",
	Description:      "Main users management microservice responsible for user related operations.",
	InfoInstanceName: "swagger",
}

// @title Users Service API
// @version 1.0.0
// @description Main users management microservice responsible for user related operations.
// @termsOfService http://not-yet-hosted.com/terms/

// @contact.name API Support
// @contact.url http://not-yet-hosted.com/support
// @contact.email m0niev@gmail.com

// @host users-service.users-service.svc.cluster.local:8000
// @BasePath /
// @externalDocs.description OpenAPI 3.0.0 Specification
// @externalDocs.url https://not-yet-hosted.com/specification
// @schemes http https

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	settings := config.GetSettings()
	app := config.NewApp(settings)

	server := &http.Server{
		Addr:    settings.ApiPort,
		Handler: app.Engine,
	}

	go func() {
		app.Logger.Info().Str("address", server.Addr).Msg("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.Logger.Fatal().Err(err).Msgf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		app.Logger.Error().Err(err).Msg("Server forced to shutdown")
	} else {
		app.Logger.Info().Msg("Server shut down gracefully.")
	}

	app.Shutdown()
	app.Logger.Info().Msg("Application shut down successfully")
}
