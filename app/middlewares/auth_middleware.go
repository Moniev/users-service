package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"users-service/app/models/utils"
	"users-service/app/services"

	"github.com/gin-gonic/gin"
)

func (m *Middlewares) UserAuthenticationMiddleware(service services.AuthServiceInterface) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr, err := m.extractBearerToken(ctx)
		if err != nil {
			m.Logger.Warn().Err(err).Msg("Failed to extract bearer token")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "failure", "error": err.Error()})
			return
		}

		if m.handleCachedAuth(ctx, tokenStr) {
			return
		}

		m.handleJWTValidation(ctx, service, tokenStr)
	}
}

func (m *Middlewares) extractBearerToken(ctx *gin.Context) (string, error) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is missing")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("invalid authorization header format, 'Bearer ' prefix is missing")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == "" {
		return "", errors.New("token is empty")
	}

	return tokenStr, nil
}

func (m *Middlewares) handleCachedAuth(ctx *gin.Context, tokenStr string) bool {
	cacheKey := "auth:token:" + tokenStr
	cachedData, err := m.CacheStore.Get(ctx, cacheKey)
	if err != nil {
		if err.Error() != "redis: nil" {
			m.Logger.Warn().Err(err).Str("cache_key", cacheKey).Msg("Failed to get token from cache")
		}
		return false
	}

	claims, err := m.CacheStore.DecacheClaims(cachedData)
	if err != nil {
		m.Logger.Warn().Err(err).Str("cache_key", cacheKey).Msg("Failed to decode cached claims")
		return false
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Unix() <= time.Now().UTC().Unix() {
		m.Logger.Info().Int("user_id", claims.UserID).Msg("Cached token has expired")
		go func() {
			if err := m.CacheStore.Del(context.Background(), cacheKey); err != nil {
				m.Logger.Warn().Err(err).Str("cache_key", cacheKey).Msg("Failed to delete expired token from cache")
			}
		}()
		return false
	}

	m.Logger.Info().Int("user_id", claims.UserID).Msg("Authentication successful from cache")
	ctx.Set("UserID", claims.UserID)
	ctx.Set("UserRoles", claims.UserRoles)
	ctx.Set("Token", tokenStr)
	ctx.Next()
	return true
}

func (m *Middlewares) handleJWTValidation(ctx *gin.Context, service services.AuthServiceInterface, tokenStr string) {
	claims, err := service.ValidateJWT(ctx.Request.Context(), tokenStr)
	if err != nil {
		m.Logger.Warn().Err(err).Str("token", tokenStr).Msg("JWT validation failed")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "failure", "error": "invalid token"})
		return
	}

	go m.cacheValidClaims(claims, tokenStr)

	m.Logger.Info().Int("user_id", claims.UserID).Msg("Authentication successful via JWT validation")
	ctx.Set("UserID", claims.UserID)
	ctx.Set("UserRoles", claims.UserRoles)
	ctx.Set("Token", tokenStr)
	ctx.Next()
}

func (m *Middlewares) cacheValidClaims(claims *utils.Claims, tokenStr string) {
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		m.Logger.Error().Err(err).Int("user_id", claims.UserID).Msg("Failed to marshal claims for caching")
		return
	}

	cacheKey := "auth:token:" + tokenStr
	if err := m.CacheStore.Set(context.Background(), cacheKey, claimsBytes, 5*time.Minute); err != nil {
		m.Logger.Error().Err(err).Str("cache_key", cacheKey).Int("user_id", claims.UserID).Msg("Failed to cache token claims")
		return
	}

	m.Logger.Debug().Int("user_id", claims.UserID).Str("cache_key", cacheKey).Msg("Successfully cached token claims")
}
