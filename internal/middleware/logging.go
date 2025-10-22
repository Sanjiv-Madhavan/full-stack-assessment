package middleware

import (
	"log/slog"
	"net/http"
)

// responseWriter is a custom http.ResponseWriter that captures the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

// WriteHeader intercepts the call to capture the status code before calling the
// original WriteHeader method.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write intercepts the call to ensure that if a Write is called without a
// preceding WriteHeader, the status code is still captured.
func (rw *responseWriter) Write(data []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(data)
}

// LoggingMiddleware is a middleware that logs the request method and URL.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)
		slog.LogAttrs(r.Context(), slog.LevelInfo, "Request handled",
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
			slog.Int("status_code", rw.statusCode),
		)
	})
}
