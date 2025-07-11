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
// @swagger 2.0

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
package main

import (
	"users-service/app/config"

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

func main() {
	settings := config.GetSettings()

	router := config.NewApp(settings)
	router.Run(settings.ApiPort)
}
