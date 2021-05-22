package metrics

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net/http"
	"strconv"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

var HttpHits = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "httpHits",
		Help: "Number of the requests with the same response code and path",
	},[]string{"status", "path"},
	)

var HttpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_response_time_seconds",
	Buckets: []float64{0.001, 0.002, 0.005, 0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 20},
	Help: "Duration of HTTP requests.",
}, []string{"status", "path"},
)

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()


		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)

		statusCode := rw.statusCode

		timer := prometheus.NewTimer(HttpDuration.WithLabelValues(strconv.Itoa(statusCode), path))
		HttpHits.WithLabelValues(strconv.Itoa(statusCode), path).Inc()

		defer timer.ObserveDuration()
	})
}
