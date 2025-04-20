package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"load/internal/balancer"
	"load/internal/config"
	"load/internal/logging"
	"load/internal/proxy"
)

// healthCheck tests a backend's health by sending a GET request.
func healthCheck(backend *balancer.Backend, healthCheckPath string) bool {
	client := &http.Client{}
	resp, err := client.Get(backend.URL.String() + healthCheckPath)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// startHealthChecks runs periodic health checks for all backends.
func startHealthChecks(backends []*balancer.Backend, interval time.Duration, healthCheckPath string) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			for _, backend := range backends {
				if !backend.CircuitBreaker.IsAvailable() {
					if healthCheck(backend, healthCheckPath) {
						log.Printf("Backend %s recovered", backend.URL.String())
						backend.CircuitBreaker.RecordSuccess()
					}
				}
			}
		}
	}()
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize load balancer
	lb, err := balancer.NewBalancer(cfg.BalancerType, cfg.Backends, cfg.CircuitBreakerFailures, cfg.CircuitBreakerCooldown)
	if err != nil {
		log.Fatalf("Failed to initialize balancer: %v", err)
	}

	// Extract backends for health checks
	var backends []*balancer.Backend
	switch b := lb.(type) {
	case *balancer.RoundRobin:
		backends = b.Backends
	case *balancer.Random:
		backends = b.Backends
	}

	// Start health checks
	startHealthChecks(backends, 2*time.Second, cfg.HealthCheckPath)

	// Create proxy handler with logging middleware
	proxyHandler := proxy.New(lb)
	handler := logging.Middleware(proxyHandler)

	// Start server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}
	go func() {
		log.Printf("Starting load balancer on via mode %s on port %s [%s]", cfg.BalancerType, cfg.Port, strings.Join(cfg.Backends, ","))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown failed: %v", err)
	}
	log.Println("Server stopped")
}
