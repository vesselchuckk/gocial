package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
    HttpRequestsTotal *prometheus.CounterVec
    HttpRequestDuration *prometheus.HistogramVec
    DBQueriesTotal prometheus.Counter
    PostsCreatedTotal prometheus.Counter
    UsersRegisteredTotal prometheus.Counter
    JWTValidationErrorsTotal prometheus.Counter
    HTTP5xxTotal prometheus.Counter
)

func init() {
    HttpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    }, []string{"method", "path", "status"})

    HttpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request durations in seconds",
        Buckets: prometheus.DefBuckets,
    }, []string{"method", "path"})

    DBQueriesTotal = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "db_queries_total",
        Help: "Total number of DB queries executed",
    })

    PostsCreatedTotal = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "posts_created_total",
        Help: "Total number of posts created",
    })

    UsersRegisteredTotal = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "users_registered_total",
        Help: "Total number of users registered",
    })

    JWTValidationErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "jwt_validation_errors_total",
        Help: "Total number of JWT validation errors",
    })

    prometheus.MustRegister(HttpRequestsTotal)
    prometheus.MustRegister(HttpRequestDuration)
    prometheus.MustRegister(DBQueriesTotal)
    prometheus.MustRegister(PostsCreatedTotal)
    prometheus.MustRegister(UsersRegisteredTotal)
    prometheus.MustRegister(JWTValidationErrorsTotal)
}

func IncDBQuery() {
    DBQueriesTotal.Inc()
}

func IncPostCreated() {
    PostsCreatedTotal.Inc()
}

func IncUserRegistered() {
    UsersRegisteredTotal.Inc()
}

func IncJWTValidationError() {
    JWTValidationErrorsTotal.Inc()
}
