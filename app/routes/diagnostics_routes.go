package routes

import (
	"users-service/app/controllers"
	"users-service/app/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// RegisterDiagnosticsRoutes registers diagnostics-related endpoints with the provided router.
// It sets up rate-limited endpoints for monitoring and accessing API diagnostics information.
//
// Parameters:
//   - prefix: The base path prefix for all diagnostics endpoints (e.g., "/api/diagnostics")
//   - router: The Gin RouterGroup to register the routes with
//   - diagnosticsController: Controller instance containing diagnostics handling logic
//   - tracker: RequestTracker instance for rate limiting
//   - middlewares: Middleware set for logging and rate limiting
//   - logger: Zer beginnersolog logger instance for logging requests
//
// The function creates the following endpoints:
//   - GET {prefix}/readiness - Checks the readiness status of the application
//   - GET {prefix}/health - Checks the health status of the application
//   - GET {prefix}/api-docs - Retrieves API documentation
//
// All endpoints are protected by request rate limiting middleware.
func RegisterDiagnosticsRoutes(prefix string,
	router *gin.RouterGroup,
	diagnosticsController controllers.DiagnosticsControllerInterface,
	tracker *middlewares.RequestTracker,
	middlewares middlewares.MiddlewaresInterface,
	logger zerolog.Logger) {
	api := router.Group(prefix).
		Use(middlewares.Log(logger, "diagnostics-controller")).
		Use(middlewares.LimitRequestsMiddleware(tracker))
	api.GET("/readiness", diagnosticsController.ReadinessProbe)
	api.GET("/health", diagnosticsController.HealthProbe)
}
