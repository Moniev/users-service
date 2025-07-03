package routes

import (
	"users-service/app/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RegisterSwaggerRoutes registers Swagger UI endpoints with the provided router.
// It sets up a rate-limited endpoint to serve Swagger API documentation.
//
// Parameters:
//   - prefix: The base path prefix for Swagger endpoints (e.g., "/swagger")
//   - router: The Gin RouterGroup to register the routes with
//   - tracker: RequestTracker instance for rate limiting
//
// The function creates the following endpoint:
//   - GET {prefix}/*any - Serves Swagger UI and API documentation
//
// The endpoint is protected by request rate limiting middleware.
// Uses ginSwagger and swaggerFiles to serve embedded Swagger documentation.
func RegisterSwaggerRoutes(
	prefix string,
	router *gin.RouterGroup,
	tracker *middlewares.RequestTracker,
	middlewares middlewares.MiddlewaresInterface,
	logger zerolog.Logger) {

	api := router.Group(prefix).
		Use(middlewares.Log(logger, "swagger")).
		Use(middlewares.LimitRequestsMiddleware(tracker))
	api.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
