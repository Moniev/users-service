package routes

import (
	_ "users-service/app/docs"
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
//   - middlewares: Interface for custom middlewares (logging, rate limiting)
//   - logger: Zerolog logger instance for logging
//
// The function creates the following endpoint:
//   - GET {prefix}/*any - Serves Swagger UI and API documentation.
//     This single handler will serve all Swagger UI assets (HTML, CSS, JS)
//     and the generated 'doc.json' file, thanks to the '_ "users-service/app/docs"' import.
//
// The endpoint is protected by request rate limiting middleware and logging middleware.
func RegisterSwaggerRoutes(
	prefix string,
	router *gin.RouterGroup,
	tracker *middlewares.RequestTracker,
	middlewares middlewares.MiddlewaresInterface,
	logger zerolog.Logger) {

	logger.Info().Str("prefix", prefix).Msg("Registering Swagger routes")

	swaggerGroup := router.Group(prefix).
		Use(middlewares.Log(logger, "swagger")).
		Use(middlewares.LimitRequestsMiddleware(tracker))
	swaggerGroup.GET("/*any", func(c *gin.Context) {
		ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/docs/swagger.json"))(c)

		if c.Writer.Status() >= 400 {
			logger.Error().Int("status", c.Writer.Status()).Str("path", c.Request.URL.Path).Msg("Failed to serve Swagger resource")
		}
	})
}
