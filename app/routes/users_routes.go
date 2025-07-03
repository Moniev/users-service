package routes

import (
	"users-service/app/controllers"
	"users-service/app/middlewares"
	"users-service/app/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// RegisterUserRoutes registers user-related endpoints with the provided router.
// It sets up authenticated and rate-limited endpoints for user management operations.
//
// Parameters:
//   - prefix: The base path prefix for all user endpoints (e.g., "/users")
//   - router: The Gin RouterGroup to register the routes with
//   - userController: Controller instance containing user handling logic
//   - tracker: RequestTracker instance for rate limiting
//   - middlewares: Middleware set for logging, authentication, and rate limiting
//   - logger: Zerolog logger instance for logging requests
//
// All endpoints are protected by:
//   - User authentication middleware (via UserAuthenticationMiddleware)
//   - Request rate limiting middleware (via LimitRequestsMiddleware)
//   - Logging middleware (via Log)
func RegisterUserRoutes(
	prefix string,
	router *gin.RouterGroup,
	usersController controllers.UsersControllerInterface,
	authService services.AuthServiceInterface,
	tracker *middlewares.RequestTracker,
	middlewares middlewares.MiddlewaresInterface,
	logger zerolog.Logger) {

	api := router.Group(prefix).
		Use(middlewares.Log(logger, "users-controller")).
		Use(middlewares.LimitRequestsMiddleware(tracker)).
		Use(middlewares.UserAuthenticationMiddleware(authService))

	api.PATCH("/update", usersController.UpdateUser)
	api.PATCH("/details/update", usersController.UpdateDetails)
	api.PATCH("/settings/update", usersController.UpdateSettings)
	api.DELETE("/remove", usersController.RemoveAccount)
}
