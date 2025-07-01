package middlewares

import (
	"users-service/app/infrastructure"
	utilsModels "users-service/app/models/utils"
	"users-service/app/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Middlewares holds the dependencies required for the middleware operations.
// It contains a Redis client for interacting with Redis and a Utils struct
// for accessing utility functions that can be used throughout different
// middleware logic.
type Middlewares struct {
	ApiVersion  string
	Environment string
	CacheStore  infrastructure.CacheStoreInterface // Redis client used for caching, querying, or other Redis-related tasks.
	Logger      zerolog.Logger
}

type MiddlewaresInterface interface {
	Log(logger zerolog.Logger, controller string) gin.HandlerFunc

	LimitRequestsMiddleware(tracker *RequestTracker) gin.HandlerFunc

	UserAuthenticationMiddleware(service services.AuthServiceInterface) gin.HandlerFunc
	extractBearerToken(ctx *gin.Context) (string, error)
	handleCachedAuth(ctx *gin.Context, tokenString string) bool
	handleJWTValidation(ctx *gin.Context, service services.AuthServiceInterface, tokenString string)
	cacheValidClaims(claims *utilsModels.Claims, tokenString string)

	MetricsCORS() gin.HandlerFunc
}

var _ MiddlewaresInterface = (*Middlewares)(nil)

// NewMiddlewares creates a new instance of Middlewares with the provided
// Redis client and utility functions. It initializes the Redis client and
// utility struct that can be used across various middleware operations.
//
// Parameters:
//   - redisClient: A pointer to an existing Redis client to interact with Redis.
//   - utils: A pointer to a utility struct containing helper functions.
//
// Returns:
//   - A pointer to a Middlewares struct, ready for use in middleware operations.
func NewMiddlewares(
	cacheStore infrastructure.CacheStoreInterface,
	logger zerolog.Logger,
	apiVersion, env string) *Middlewares {

	return &Middlewares{
		ApiVersion:  apiVersion,
		Environment: env,
		CacheStore:  cacheStore,
		Logger:      logger,
	}
}
