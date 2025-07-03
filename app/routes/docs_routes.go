package routes

import (
	"users-service/app/middlewares"
	"users-service/app/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func RegisterDocumentationRoutes(
	prefix string,
	router *gin.RouterGroup,

	authService services.AuthServiceInterface,
	tracker *middlewares.RequestTracker,
	middlewares middlewares.MiddlewaresInterface,
	logger zerolog.Logger) {

	api := router.Group(prefix).
		Use(middlewares.Log(logger, "users-controller")).
		Use(middlewares.LimitRequestsMiddleware(tracker))
	api.StaticFile("/swagger.json", "./swagger.json")
	api.StaticFile("/swagger.yaml", "./swagger.yaml")
}
