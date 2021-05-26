package metrics

import (
	"bufio"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type responseWriterProxy struct {
	responseWriter http.ResponseWriter
	statusCode     int
}

func NewResponseWriter(w http.ResponseWriter) *responseWriterProxy {
	return &responseWriterProxy{w, http.StatusOK}
}

func (rw *responseWriterProxy) WriteHeader(code int) {
	rw.statusCode = code
	rw.responseWriter.WriteHeader(code)
}

func (rw *responseWriterProxy) Header() http.Header {
	return rw.responseWriter.Header()
}

func (rw *responseWriterProxy) Write(bytes []byte) (int, error) {
	return rw.responseWriter.Write(bytes)
}

func (rw *responseWriterProxy) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rw.responseWriter.(http.Hijacker).Hijack()
}

var HttpHits = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "httpHits",
		Help: "Number of the requests with the same response code and path",
	}, []string{"status", "path"},
)

var HttpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "http_response_time_seconds",
	Buckets: []float64{0.001, 0.002, 0.005, 0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 20},
	Help:    "Duration of HTTP requests.",
}, []string{"path"},
)

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		rw := NewResponseWriter(w)

		timer := prometheus.NewTimer(HttpDuration.WithLabelValues(path))

		next.ServeHTTP(rw, r)

		statusCode := rw.statusCode

		HttpHits.WithLabelValues(strconv.Itoa(statusCode), path).Inc()

		defer timer.ObserveDuration()
	})
}
