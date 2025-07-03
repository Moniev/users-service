package main

import "users-service/app/config"

// @title           Factory-Chainline Users-Service API
// @version         1.0.0
// @description     Microservice delegated toward user's schema management and it's business logic.
// @termsOfService

// @contact.name   API Support
// @contact.url    https://github.com/Moniev
// @contact.email  m0ni3v@gmail.com

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BearerApi

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://not-yet-hosted.com/
func main() {
	settings := config.GetSettings()

	router := config.NewApp(settings)
	router.Run(settings.ApiPort)
}
