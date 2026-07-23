package metrics

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type metricsResponseWriter struct {
    http.ResponseWriter
    status int
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	if rw.status != 0 {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *metricsResponseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}

	return rw.ResponseWriter.Write(b)
}

func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        mrw := &metricsResponseWriter{
            ResponseWriter: w,
            status: 0,
        }

        next.ServeHTTP(mrw, r)

        status := mrw.status
        if status == 0 {
            status = http.StatusOK
        }

        route := chi.RouteContext(r.Context()).RoutePattern()
        if route == "" {
            route = "unmatched"
        }
        if route == "/metrics" {
            return
        }

        ObserveHTTPRequest(
            r.Method,
            route,
            status,
            time.Since(start),
        )

    })
}
