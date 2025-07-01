package middlewares

import (
	"time"
	"users-service/app/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Log returns a Gin middleware that logs HTTP requests and responses.
// It captures request details such as method, path, and client IP, along with response status,
// duration, and size. Logs are written using the provided Logrus logger with fields
// specific to the request, including a unique request ID, controller name, API version,
// and environment. The middleware logs at different levels (Info, Warn, Error) based
// on the response status and any Gin context errors.
//
// Parameters:
//   - logger: The Logrus logger instance used for logging.
//   - controller: The name of the controller handling the request.
//
// Returns:
//   - gin.HandlerFunc: A middleware function compatible with Gin's routing.
func (m *Middlewares) Log(logger zerolog.Logger, controller string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()

		requestID := utils.GetRequestID(ctx)
		if requestID == "" {
			ctx.Next()
		}

		logEvent := logger.Info().
			Str("requestID", requestID).
			Str("method", ctx.Request.Method).
			Str("path", ctx.Request.URL.Path).
			Str("remoteAddress", ctx.ClientIP()).
			Str("controller", controller).
			Str("APIversion", m.ApiVersion).
			Str("environment", m.Environment)

		logEvent.Msg("Incoming request")

		ctx.Next()

		duration := time.Since(startTime)

		event := logger.With().
			Str("requestID", requestID).
			Str("method", ctx.Request.Method).
			Str("path", ctx.Request.URL.Path).
			Str("remoteAddress", ctx.ClientIP()).
			Str("controller", controller).
			Str("APIversion", m.ApiVersion).
			Str("environment", m.Environment).
			Int("status", ctx.Writer.Status()).
			Int64("durationMS", duration.Milliseconds()).
			Int("responseSize", ctx.Writer.Size()).
			Logger()

		switch {
		case ctx.Writer.Status() >= 500:
			event.Error().Msg("Request completed with error")
		case ctx.Writer.Status() >= 400:
			event.Warn().Msg("Request completed with client error")
		default:
			event.Info().Msg("Request completed successfully")
		}

		if len(ctx.Errors) > 0 {
			logEvent.Err(ctx.Errors[0]).Msgf("Request errors: %v", ctx.Errors)
		}
	}
}
