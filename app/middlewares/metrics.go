// Copyright © 2025 Robert Moń. All rights reserved.
// This source code is proprietary and confidential. Unauthorized copying,
// distribution, or commercial use is prohibited without prior written consent.
package middlewares

import "github.com/prometheus/client_golang/prometheus"

// HttpRequestCountWithPath is a CounterVec that tracks the total number of HTTP requests by URL path.
// It increments for each request, labeled with the "url" of the request path.
// Use this metric to monitor request volume per endpoint.
var HttpRequestCountWithPath = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total_by_path",
		Help: "Number of HTTP requests by path.",
	},
	[]string{"url"},
)

// HttpRequestDuration is a HistogramVec that measures the response time of HTTP requests in seconds.
// It records durations labeled with the "path" of the request.
// Use this metric to analyze latency distribution across different endpoints.
var HttpRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Response time of HTTP request.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"path"},
)

// HttpRequestStatusCode is a CounterVec that tracks the total number of HTTP requests by status code.
// It increments for each request, labeled with the "url" and "status_code" (e.g., "200", "404").
// Use this metric to monitor the frequency of different response statuses per endpoint.
var HttpRequestStatusCode = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total_by_status",
		Help: "Number of HTTP requests by status code.",
	},
	[]string{"url", "status_code"},
)

// HttpRequestSize is a HistogramVec that measures the size of HTTP requests in bytes.
// It records request sizes labeled with the "url" of the request path.
// Use this metric to analyze the distribution of incoming request payloads.
var HttpRequestSize = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "http_request_size_bytes",
		Help: "Size of HTTP requests in bytes.",
	},
	[]string{"url"},
)

// HttpResponseSize is a HistogramVec that measures the size of HTTP responses in bytes.
// It records response sizes labeled with the "url" of the request path.
// Use this metric to analyze the distribution of outgoing response payloads.
var HttpResponseSize = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "http_response_size_bytes",
		Help: "Size of HTTP responses in bytes.",
	},
	[]string{"url"},
)

// ActiveRequests is a Gauge that tracks the number of currently active HTTP requests.
// It increases when a request starts and decreases when it completes.
// Use this metric to monitor concurrency and load on the server.
var ActiveRequests = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "active_http_requests",
		Help: "Number of currently active HTTP requests.",
	},
)

// CPUUsage is a Gauge that tracks the current CPU usage as a percentage.
// It should be updated periodically to reflect system resource utilization.
// Use this metric to monitor CPU load on the server.
var CPUUsage = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "cpu_usage_percent",
		Help: "Current CPU usage as a percentage.",
	},
)

// MemoryUsage is a Gauge that tracks the current memory usage in megabytes.
// It should be updated periodically to reflect system resource utilization.
// Use this metric to monitor memory consumption on the server.
var MemoryUsage = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "memory_usage_megabytes",
		Help: "Current memory usage in bytes.",
	},
)

// ErrorRate is a CounterVec that tracks the total number of HTTP errors.
// It increments for each error, labeled with the "url" and "error_type" (e.g., "timeout", "validation").
// Use this metric to monitor error frequency and types per endpoint.
var ErrorRate = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_errors_total",
		Help: "Total number of HTTP errors.",
	},
	[]string{"url", "error_type"},
)

// SuccessRate is a CounterVec that tracks the total number of HTTP successes.
// It increments for each successful request, labeled with the "url" and "success_type" (e.g., "created", "ok").
// Use this metric to monitor success frequency and types per endpoint.
var SuccessRate = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_successes_total",
		Help: "Total number of HTTP successes.",
	},
	[]string{"url", "success_type"},
)

// DiskUsage is a Gauge that tracks the current disk usage in bytes.
// It should be updated periodically to reflect system storage utilization.
// Use this metric to monitor disk space consumption on the server.
var DiskUsage = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "disk_usage_bytes",
		Help: "Current disk usage in bytes.",
	},
)

// ThreadCount is a Gauge that tracks the number of active threads.
// It should be updated to reflect the current number of goroutines or threads in use.
// Use this metric to monitor concurrency and resource usage in the application.
var ThreadCount = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "thread_count",
		Help: "Number of active threads.",
	},
)

// DBQueryDuration is a HistogramVec that measures the duration of database queries in seconds.
// It records query times labeled with the "operation" (e.g., "select", "insert") and "table" name.
// Use this metric to analyze the performance of database operations executed by Ent.
var DBQueryDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "db_query_duration_seconds",
		Help: "Duration of database queries executed by Ent in seconds.",
	},
	[]string{"operation", "table"},
)

// CacheMisses is a Counter that tracks the total number of cache misses.
// It increments each time a cache lookup fails to find a value.
// Use this metric to monitor cache effectiveness.
var CacheMisses = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total number of cache misses.",
	},
)

// CacheHits is a Counter that tracks the total number of cache hits.
// It increments each time a cache lookup successfully retrieves a value.
// Use this metric to monitor cache effectiveness.
var CacheHits = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total number of cache hits.",
	},
)

// AuthFailures is a Counter that tracks the total number of authentication failures.
// It increments for each failed authentication attempt.
// Use this metric to monitor security-related issues or brute-force attempts.
var AuthFailures = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "auth_failures_total",
		Help: "Total number of authentication failures.",
	},
)

// BlockedIPs is a Gauge that tracks the number of currently blocked IP addresses.
// It should be updated as IPs are blocked or unblocked (e.g., by a rate-limiting or security mechanism).
// Use this metric to monitor the effectiveness of IP-based security measures.
var BlockedIPs = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "blocked_ips",
		Help: "Number of currently blocked IP addresses.",
	},
)
