package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"

	"load/internal/balancer"
	"load/internal/logging"
)

// Handler is an HTTP handler for the load balancer.
type Handler struct {
	balancer balancer.Balancer
}

// New creates a new proxy handler with the given balancer.
func New(b balancer.Balancer) *Handler {
	return &Handler{balancer: b}
}

// ServeHTTP handles incoming HTTP requests by proxying them to a backend.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	attempts := 0
	maxAttempts := 3         // Limit retries to avoid infinite loops
	clientCtx := r.Context() // Capture original client context

	for attempts < maxAttempts {
		backend, err := h.balancer.NextBackend()
		if err != nil {
			fmt.Printf("\nNo available backends after %d attempts: %v", attempts, err)
			http.Error(w, "No available backends", http.StatusServiceUnavailable)
			return
		}

		fmt.Printf("\nAttempt %d: Forwarding to backend: %s", attempts+1, backend.URL.String())
		proxy := httputil.NewSingleHostReverseProxy(backend.URL)
		failed := false
		clientCanceled := false

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			fmt.Printf("\nBackend %s failed for %s: %v", backend.URL.String(), r.URL.String(), err)
			if clientCtx.Err() == context.Canceled {
				fmt.Printf("\nClient canceled request for %s, not penalizing backend %s", r.URL.String(), backend.URL.String())
				clientCanceled = true
				failed = true
			} else {
				fmt.Printf("\nRecording failure for backend %s due to error: %v", backend.URL.String(), err)
				backend.CircuitBreaker.RecordFailure()
				failed = true
			}
		}
		proxy.ModifyResponse = func(resp *http.Response) error {
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				backend.CircuitBreaker.RecordSuccess()
			} else {
				fmt.Printf("\nBackend %s returned non-2xx status %d for %s", backend.URL.String(), resp.StatusCode, r.URL.String())
				backend.CircuitBreaker.RecordFailure()
				failed = true
			}
			return nil
		}

		// Use a custom ResponseWriter to prevent writing until we succeed
		rw := &responseWriter{ResponseWriter: w, written: false}
		if lrw, ok := w.(*logging.LoggingResponseWriter); ok {
			lrw.SetBackendURL(backend.URL.String())
		}
		proxy.ServeHTTP(rw, r)

		if clientCanceled {
			fmt.Printf("\nAborting retries for %s due to client cancellation", r.URL.String())
			http.Error(w, "Client closed request", 499) // Nginx-style 499 for client disconnect
			return
		}

		if !failed && !rw.written {
			fmt.Printf("\nBackend %s did not write response for %s", backend.URL.String(), r.URL.String())
			backend.CircuitBreaker.RecordFailure()
			failed = true
		}

		if !failed {
			return // Success, response sent to client
		}

		attempts++
	}

	fmt.Printf("\nAll %d attempts failed for %s", attempts, r.URL.String())
	http.Error(w, "All backends unavailable", http.StatusServiceUnavailable)
}

// responseWriter wraps http.ResponseWriter to track if a response was written.
type responseWriter struct {
	http.ResponseWriter
	written bool
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.written = true
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.written = true
	return rw.ResponseWriter.Write(b)
}
