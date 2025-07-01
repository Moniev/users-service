// Copyright © 2025 Robert Moń. All rights reserved.
// This source code is proprietary and confidential. Unauthorized copying,
// distribution, or commercial use is prohibited without prior written consent.
package middlewares

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"users-service/app/models/ent"
	"users-service/app/utils"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/cpu"
)

// UpdateCacheHitRatioMetrics returns a Gin middleware that tracks cache hit and miss ratios using Redis.
// It checks if the request URL path exists as a key in Redis. If the key is not found (redis.Nil), it
// increments the CacheMisses counter and proceeds with the request. If found, it increments the
// CacheHits counter and aborts the request with a 200 OK response containing an empty JSON object.
// Errors other than redis.Nil are ignored, allowing the request to proceed without metric updates.
//
// This middleware assumes the Redis client is properly configured and the cache key matches the
// request URL path. It’s intended for monitoring cache effectiveness in a caching layer.
//
// Parameters:
//   - redisClient: A pointer to a Redis client instance used to query the cache.
//
// Returns:
//   - A gin.HandlerFunc that can be used as middleware in a Gin router.
func (m *Middlewares) UpdateCacheHitRatioMetrics(redisClient *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cacheKey := ctx.Request.URL.Path
		_, err := redisClient.Get(ctx, cacheKey).Result()

		if err == redis.Nil {
			CacheMisses.Inc()
			ctx.Next()
		} else if err != nil {
			ctx.Next()
		} else {
			CacheHits.Inc()
			ctx.AbortWithStatusJSON(200, gin.H{})
		}
	}
}

// UpdateDBMetrics returns an Ent hook that tracks the duration of database queries.
// It wraps an ent.Mutator to measure the time taken for each mutation (e.g., create, update, delete)
// using a Prometheus timer. The duration is recorded in the DBQueryDuration histogram, labeled with
// the operation type (e.g., "create", "update") and the entity type (table name).
//
// This hook should be registered with an Ent client to monitor database performance.
//
// Returns:
//   - An ent.Hook that wraps the next mutator in the chain with timing metrics.
func UpdateDBMetrics() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
			timer := prometheus.NewTimer(DBQueryDuration.WithLabelValues(mutation.Op().String(), mutation.Type()))
			defer timer.ObserveDuration()

			return next.Mutate(ctx, mutation)
		})
	}
}

// UpdateServerMetrics is a Gin middleware that tracks HTTP request metrics.
// It increments ActiveRequests when a request begins and decrements it when it ends.
// After the request is processed, it records:
//   - Request duration in HttpRequestDuration (labeled by method and path).
//   - Request count in HttpRequestCountWithPath (labeled by path).
//   - Status code in HttpRequestStatusCode (labeled by path and status).
//   - Request size in HttpRequestSize (header + body, labeled by path).
//   - Response size in HttpResponseSize (labeled by path).
//   - Success or error counts in SuccessRate or ErrorRate (labeled by path and status/error type).
//
// Success is determined by IsSuccessStatusCode. This middleware should be used early in the handler
// chain to capture all metrics accurately.
//
// Parameters:
//   - ctx: The Gin context for the current request.
func (m *Middlewares) UpdateServerMetrics() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ActiveRequests.Inc()
		defer ActiveRequests.Dec()

		ctx.Next()

		reqMethodAndPath := fmt.Sprintf("[%s] %s", ctx.Request.Method, ctx.FullPath())
		timer := prometheus.NewTimer(HttpRequestDuration.WithLabelValues(reqMethodAndPath))
		defer timer.ObserveDuration()

		var responseSize float64 = float64(ctx.Writer.Size())

		var headerSize int = 0
		for key, values := range ctx.Request.Header {
			headerSize += len(key)
			for _, value := range values {
				headerSize += len(value)
			}
		}

		bodySize := ctx.Request.ContentLength
		if bodySize < 0 {
			bodySize = 0
		}

		requestSize := float64(headerSize) + float64(bodySize)

		var statusCode string = strconv.Itoa(ctx.Writer.Status())

		if !utils.IsSuccessStatusCode(ctx.Writer.Status()) {
			ErrorRate.WithLabelValues(ctx.FullPath(), statusCode).Inc()
		} else {
			SuccessRate.WithLabelValues(ctx.FullPath(), statusCode).Inc()
		}

		HttpRequestCountWithPath.WithLabelValues(ctx.FullPath()).Inc()
		HttpRequestStatusCode.WithLabelValues(ctx.FullPath(), statusCode).Inc()
		HttpRequestSize.WithLabelValues(ctx.FullPath()).Observe(requestSize)
		HttpResponseSize.WithLabelValues(ctx.FullPath()).Observe(responseSize)
	}
}

// UpdateSystemMetrics updates system-level Prometheus metrics for CPU, memory, and thread usage.
// It sets:
//   - CPUUsage to the number of CPU cores (as a rough proxy for usage).
//   - MemoryUsage to the total system memory allocated (in bytes).
//   - ThreadCount to the number of active goroutines.
//
// This function should be called periodically (e.g., in a middleware or background task) to keep
// metrics current. Note that CPUUsage currently uses NumCPU as a static value; for true usage, consider
// integrating with system-level monitoring (e.g., via runtime or external libraries).
//
// Parameters:
//   - ctx: The Gin context for the current request (unused in current implementation).
func (m *Middlewares) UpdateSystemMetrics() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		percent, err := cpu.Percent(0, false)
		if err != nil {
			CPUUsage.Set(0)
		} else {
			CPUUsage.Set(percent[0])
		}

		var memory runtime.MemStats
		runtime.ReadMemStats(&memory)
		MemoryUsage.Set(float64(memory.HeapAlloc) / 1024 / 1024)

		ThreadCount.Set(float64(runtime.NumGoroutine()))
		ctx.Next()
	}
}
