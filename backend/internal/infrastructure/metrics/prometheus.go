package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Search Metrics
	SearchQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "search_queries_total",
			Help: "Total number of search queries",
		},
		[]string{"content_type", "sort_by"},
	)

	SearchResultsTotal = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "search_results_total",
			Help:    "Number of search results returned",
			Buckets: []float64{0, 1, 5, 10, 20, 50, 100, 200, 500},
		},
		[]string{"content_type"},
	)

	// Cache Metrics
	CacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	CacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
	)

	// Provider Sync Metrics
	ProviderSyncDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "provider_sync_duration_seconds",
			Help:    "Provider synchronization duration in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
		[]string{"provider_name"},
	)

	ProviderSyncItemsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "provider_sync_items_total",
			Help: "Total number of items synced from provider",
		},
		[]string{"provider_name", "status"},
	)

	ProviderSyncErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "provider_sync_errors_total",
			Help: "Total number of provider sync errors",
		},
		[]string{"provider_name", "error_type"},
	)

	// Database Metrics
	DatabaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
		},
		[]string{"operation", "table"},
	)

	// Rate Limiting Metrics
	RateLimitExceededTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_exceeded_total",
			Help: "Total number of rate limit exceeded events",
		},
		[]string{"endpoint"},
	)
)

// RecordHTTPRequest records an HTTP request metric
func RecordHTTPRequest(method, path string, status int, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, path, string(rune(status))).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// RecordSearchQuery records a search query metric
func RecordSearchQuery(contentType, sortBy string, resultCount int) {
	SearchQueriesTotal.WithLabelValues(contentType, sortBy).Inc()
	SearchResultsTotal.WithLabelValues(contentType).Observe(float64(resultCount))
}

// RecordCacheHit records a cache hit
func RecordCacheHit() {
	CacheHitsTotal.Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss() {
	CacheMissesTotal.Inc()
}

// RecordProviderSync records provider sync metrics
func RecordProviderSync(providerName string, duration float64, itemCount int, status string) {
	ProviderSyncDuration.WithLabelValues(providerName).Observe(duration)
	ProviderSyncItemsTotal.WithLabelValues(providerName, status).Add(float64(itemCount))
}

// RecordProviderSyncError records a provider sync error
func RecordProviderSyncError(providerName, errorType string) {
	ProviderSyncErrorsTotal.WithLabelValues(providerName, errorType).Inc()
}

// RecordDatabaseQuery records a database query metric
func RecordDatabaseQuery(operation, table string, duration float64) {
	DatabaseQueriesTotal.WithLabelValues(operation, table).Inc()
	DatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// RecordRateLimitExceeded records a rate limit exceeded event
func RecordRateLimitExceeded(endpoint string) {
	RateLimitExceededTotal.WithLabelValues(endpoint).Inc()
}
