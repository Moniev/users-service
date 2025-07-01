package middlewares

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// MetricsCORS returns a Gin middleware that sets CORS headers for metrics endpoints.
// It allows cross-origin requests from the URL specified in the PROMETHEUS_URL environment
// variable, restricting methods to GET and permitting specific headers (Origin, Content-Type,
// Authorization). If the request uses the CONNECT method, it aborts with an HTTP 400 Bad Request
// status; otherwise, it proceeds to the next handler.
//
// This middleware is intended for use with Prometheus metrics endpoints to ensure they are
// accessible only from the configured Prometheus server while adhering to CORS security policies.
// If PROMETHEUS_URL is not set, the Access-Control-Allow-Origin header will be an empty string,
// effectively disabling CORS unless explicitly configured.
//
// Returns:
//   - A gin.HandlerFunc that can be used as middleware in a Gin router.
func (m *Middlewares) MetricsCORS() gin.HandlerFunc {
	prometheusURL := os.Getenv("PROMETHEUS_URL")

	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", prometheusURL)
		ctx.Header("Access-Control-Allow-Methods", "GET")
		ctx.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if ctx.Request.Method == http.MethodConnect {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		ctx.Next()
	}
}
