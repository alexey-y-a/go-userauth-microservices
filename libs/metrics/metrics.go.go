package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var httpRequestsTotal = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests processed.",
    },
    []string{"service", "path", "method", "status"},
)

type statusRecorder struct {
    http.ResponseWriter
    status int
}

func (r *statusRecorder) WriteHeader (statusCode int) {
    r.status = statusCode
    r.ResponseWriter.WriteHeader(statusCode)
}

func InstrumentHandler(serviceName string, handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rec := &statusRecorder{
            ResponseWriter: w,
            status: http.StatusOK,
        }

        handler.ServeHTTP(rec, r)

        statusStr := strconv.Itoa(rec.status)

        httpRequestsTotal.WithLabelValues(
            serviceName,
            r.URL.Path,
            r.Method,
            statusStr,
        ).Inc()
    })
}