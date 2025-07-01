package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestAuthenticationMiddleware generates a new UUID request ID for each incoming request
// and stores it in the request context. This helps with request tracking and logging.
//
// Usage:
//
//	router.Use(middlewares.RequestAuthenticationMiddleware())
//
// Context Key:
//   - "RequestID": A unique string identifier for the request.
//
// Returns:
//   - gin.HandlerFunc: Middleware function to be used in the Gin framework.
func (middlewares *Middlewares) RequestAuthenticationMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := uuid.New().String()
		ctx.Set("RequestID", requestID)
		ctx.Next()
	}
}
