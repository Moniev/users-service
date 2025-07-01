package routes

import (
	"users-service/app/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// RegisterMonitoringRoutes registers monitoring-related endpoints with the provided router.
// It sets up a rate-limited GET endpoint for exposing application metrics.
//
// Parameters:
//   - prefix: The base path prefix for monitoring endpoints (e.g., "/monitoring")
//   - router: The Gin RouterGroup to register the routes with
//   - tracker: RequestTracker instance for rate limiting
//
// The function creates the following endpoint:
//   - GET {prefix}/metrics - Serves Prometheus metrics
//
// The endpoint is protected by:
//   - CORS middleware for metrics access
//   - Request rate limiting middleware
//
// Uses promhttp.Handler() from the Prometheus client library to serve metrics.
func RegisterMonitoringRoutes(prefix string, router *gin.RouterGroup, tracker *middlewares.RequestTracker, middlewares *middlewares.Middlewares, logger zerolog.Logger) {
	api := router.Group(prefix).
		Use(middlewares.Log(logger, "monitoring")).
		Use(middlewares.MetricsCORS()).
		Use(middlewares.LimitRequestsMiddleware(tracker))
	api.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
