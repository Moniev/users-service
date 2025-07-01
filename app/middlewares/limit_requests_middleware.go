package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestTracker manages request rate limiting by tracking request timestamps per client IP.
// It uses a sliding window approach to enforce a maximum number of requests within a specified
// time duration. The tracker is thread-safe, using a mutex to protect concurrent access to its IP map.
type RequestTracker struct {
	Window     time.Duration          // The time window for rate limiting (e.g., 1 minute).
	MaxRequest int                    // The maximum number of requests allowed within the window.
	IPMap      map[string][]time.Time // Maps client IPs to their request timestamps.
	Mutex      sync.Mutex             // Ensures thread-safe access to IPMap.
}

// NewRequestTracker creates a new RequestTracker instance with the specified maximum requests.
// If maxRequests is 0, it defaults to 20. The tracker uses a 1-minute sliding window and initializes
// an empty IP map for tracking requests.
//
// Parameters:
//   - maxRequests: The maximum number of requests allowed per IP within the window.
//
// Returns:
//   - A pointer to a newly initialized RequestTracker.
func NewRequestTracker(maxRequests int) *RequestTracker {
	if maxRequests == 0 {
		maxRequests = 20
	}

	return &RequestTracker{
		Window:     1 * time.Minute,
		MaxRequest: maxRequests,
		IPMap:      make(map[string][]time.Time),
	}
}

// LimitRequestsMiddleware returns a Gin middleware that enforces request rate limiting.
// It tracks requests per client IP using the provided RequestTracker. If the number of requests
// from an IP exceeds the tracker's MaxRequest within the Window duration, it responds with an
// HTTP 429 Too Many Requests status and a JSON body containing a message, evidence, User-Agent,
// and client IP, then aborts the request. Otherwise, it allows the request to proceed.
//
// The middleware uses a sliding window approach, retaining only timestamps within the Window
// duration and appending the current request time. It is thread-safe due to the tracker's mutex.
//
// Parameters:
//   - tracker: A pointer to a RequestTracker instance that manages rate limiting state.
//
// Returns:
//   - A gin.HandlerFunc that can be used as middleware in a Gin router.
func (m *Middlewares) LimitRequestsMiddleware(tracker *RequestTracker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientIP := ctx.ClientIP()
		userAgent := ctx.Request.Header.Get("User-Agent")

		now := time.Now()

		tracker.Mutex.Lock()
		defer tracker.Mutex.Unlock()

		requests, exists := tracker.IPMap[clientIP]
		if !exists {
			requests = []time.Time{}
		}

		validRequests := make([]time.Time, 0, len(requests))
		for _, time := range requests {
			if now.Sub(time) <= tracker.Window {
				validRequests = append(validRequests, time)
			}
		}

		validRequests = append(validRequests, now)
		tracker.IPMap[clientIP] = validRequests

		if len(validRequests) > tracker.MaxRequest {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message":    "The Federation of Super Earth keeps managed Democracy safe with the lives of our heroes!",
				"evidence":   "Rate limit exceeded: to much requests!",
				"user_agent": userAgent,
				"client_ip":  ctx.ClientIP(),
			})
			return
		}

		ctx.Next()
	}
}

// Cleanup starts a background goroutine that periodically cleans up expired request timestamps.
// It runs at the specified interval, removing timestamps outside the Window duration from each IP's
// request list in the tracker's IPMap. If an IP has no valid requests remaining, it is removed from
// the map. The cleanup process is thread-safe due to the tracker's mutex.
//
// This method should be called once after creating a RequestTracker to ensure the IPMap doesn’t
// grow indefinitely. The goroutine stops when the ticker is garbage collected or the program exits.
//
// Parameters:
//   - interval: The duration between cleanup operations (e.g., 5 * time.Minute).
func (t *RequestTracker) Cleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			t.Mutex.Lock()
			now := time.Now()
			for ip, requests := range t.IPMap {
				validRequests := make([]time.Time, 0, len(requests))
				for _, time := range requests {
					if now.Sub(time) <= t.Window {
						validRequests = append(validRequests, time)
					}
				}
				if len(validRequests) > 0 {
					t.IPMap[ip] = validRequests
				} else {
					delete(t.IPMap, ip)
				}
			}
			t.Mutex.Unlock()
		}
	}()
}
