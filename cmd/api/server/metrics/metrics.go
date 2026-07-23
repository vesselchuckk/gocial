package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
    httpRequestsTotal *prometheus.CounterVec
    httpRequestDuration *prometheus.HistogramVec
    
    dbQueriesTotal prometheus.Counter
    postsCreatedTotal prometheus.Counter
    usersRegisteredTotal prometheus.Counter
    jwtValidationErrorsTotal prometheus.Counter
)

func init() {
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        }, 
        []string{"method", "route", "status"},
    )

    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request durations in seconds",
            Buckets: prometheus.DefBuckets,
        }, 
        []string{"method", "route"},
    )

    dbQueriesTotal = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "db_queries_total",
        Help: "Total number of DB queries executed",
    })

    postsCreatedTotal = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "posts_created_total",
        Help: "Total number of posts created",
    })

    usersRegisteredTotal = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "users_registered_total",
        Help: "Total number of users registered",
    })

    jwtValidationErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "jwt_validation_errors_total",
        Help: "Total number of JWT validation errors",
    })

    prometheus.MustRegister(
        httpRequestsTotal,
        httpRequestDuration,
        dbQueriesTotal,
        postsCreatedTotal,
        usersRegisteredTotal,
        jwtValidationErrorsTotal,
    )
}

func ObserveHTTPRequest(method, route string, status int, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(
		method,
		route,
		strconv.Itoa(status),
	).Inc()

	httpRequestDuration.WithLabelValues(
		method,
		route,
	).Observe(duration.Seconds())
}

func IncDBQuery() {
	dbQueriesTotal.Inc()
}

func IncPostCreated() {
	postsCreatedTotal.Inc()
}

func IncUserRegistered() {
	usersRegisteredTotal.Inc()
}

func IncJWTValidationError() {
	jwtValidationErrorsTotal.Inc()
}