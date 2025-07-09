package main

import "users-service/app/config"

// @title Users Service API
// @version 1.0.0
// @description Main users management microservice responsible for user related operations.
// @termsOfService http://not-yet-hosted/terms/

// @contact.name API Support
// @contact.url http://not-yet-hosted/support
// @contact.email m0niev@gmail.com

// @host users-service-users-service.svc.cluster.local:8000
// @BasePath /
// @externalDocs.description OpenAPI 3.0 Specification
// @externalDocs.url https://not-yet-hosted/specification
// @schemes http https

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @Swagger / @version 3.0
func main() {
	settings := config.GetSettings()

	router := config.NewApp(settings)
	router.Run(settings.ApiPort)
}
