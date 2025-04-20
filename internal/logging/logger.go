package logging

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Middleware wraps an http.Handler with logging functionality.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter to capture status code and backend
		lrw := &LoggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)

		// Calculate duration
		duration := time.Since(start)

		// Color codes
		reset := "\033[0m"
		green := "\033[32m"
		yellow := "\033[33m"
		red := "\033[31m"
		cyan := "\033[36m"
		purple := "\033[35m"

		// Color method
		methodColor := cyan
		switch r.Method {
		case "GET":
			methodColor = green
		default:
			methodColor = purple
		}

		// Color status
		statusColor := green
		switch {
		case lrw.statusCode >= 500:
			statusColor = red
		case lrw.statusCode >= 400:
			statusColor = yellow
		case lrw.statusCode >= 300:
			statusColor = cyan
		}

		backendInfo := ""
		if lrw.backendURL != "" {
			backendInfo = fmt.Sprintf(" -> %s", lrw.backendURL)
		}
		log.Printf(
			"%s[%s]%s %s %s %s{%d}%s%s ------------------- %s%v%s",
			methodColor, r.Method, reset,
			r.URL.Path,
			r.URL.RawQuery,
			statusColor, lrw.statusCode, reset,
			backendInfo,
			purple, duration, reset,
		)
	})
}

// LoggingResponseWriter wraps an http.ResponseWriter to capture the status code and backend URL.
type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	backendURL string
}

// WriteHeader captures the status code and calls the underlying WriteHeader.
func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// SetBackendURL sets the backend URL for logging.
func (lrw *LoggingResponseWriter) SetBackendURL(url string) {
	lrw.backendURL = url
}
