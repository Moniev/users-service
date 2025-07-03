package routes

import (
	"users-service/app/controllers"
	"users-service/app/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// RegisterAuthRoutes registers authentication-related endpoints with the provided router.
// It sets up rate-limited endpoints for user authentication and registration flows.
//
// Parameters:
//   - prefix: The base path prefix for all auth endpoints (e.g., "/auth")
//   - router: The Gin RouterGroup to register the routes with
//   - userController: Controller instance containing authentication handling logic
//   - tracker: RequestTracker instance for rate limiting
//   - middlewares: Middleware set for logging, authentication, and rate limiting
//   - logger: Zerolog logger instance for logging requests
//
// All endpoints are protected by request rate limiting and logging middleware.
func RegisterAuthRoutes(
	prefix string,
	router *gin.RouterGroup,
	authController controllers.AuthControllerInterface,
	tracker *middlewares.RequestTracker,
	middlewares middlewares.MiddlewaresInterface,
	logger zerolog.Logger) {

	api := router.Group(prefix).
		Use(middlewares.Log(logger, "auth-controller")).
		Use(middlewares.LimitRequestsMiddleware(tracker))
	api.GET("/register", authController.Register)
	api.POST("/register/externally", authController.RegisterExternally)
	api.PATCH("/activation/account", authController.ActivateAccount)
	api.POST("/activation/resend/code", authController.ResendActivationCode)
	api.PATCH("/verification/account", authController.VerifyAccount)
	api.POST("/verification/resend/code", authController.ResendVerificationCode)
	api.GET("/login", authController.Login)
	api.POST("/login/externally", authController.LoginExternally)
	api.DELETE("/logout", authController.Login)
	api.POST("/password/reset/request", authController.RequestPasswordReset)
	api.DELETE("/password/reset/request/cancel", authController.RequestPasswordReset)
	api.PATCH("/password/reset/request/confirm", authController.RequestPasswordReset)
	api.POST("/password/reset/request/resend", authController.RequestPasswordReset)
}
