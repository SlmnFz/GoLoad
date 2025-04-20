package balancer

import (
	"testing"
	"time"
)

func TestRoundRobin_NextBackend(t *testing.T) {
	backends := []string{"http://localhost:8081", "http://localhost:8082"}
	rr, err := NewRoundRobin(backends, 3, 30*time.Second)
	if err != nil {
		t.Fatalf("Failed to create RoundRobin: %v", err)
	}

	// Expected order: 8081, 8082, 8081, 8082
	expected := []string{"http://localhost:8081", "http://localhost:8082"}
	for i := range 4 {
		backend, err := rr.NextBackend()
		if err != nil {
			t.Fatalf("NextBackend failed: %v", err)
		}
		if backend.URL.String() == expected[i%2] {
			t.Errorf("Expected backend %s, got %s", expected[i%2], backend.URL.String())
		}
	}
}

func TestRandom_NextBackend(t *testing.T) {
	backends := []string{"http://localhost:8081", "http://localhost:8082"}
	rand, err := NewRandom(backends, 3, 30*time.Second)
	if err != nil {
		t.Fatalf("Failed to create Random: %v", err)
	}

	// Check that we get valid backends
	for range 10 {
		backend, err := rand.NextBackend()
		if err != nil {
			t.Fatalf("NextBackend failed: %v", err)
		}
		url := backend.URL.String()
		if url != "http://localhost:8081" && url != "http://localhost:8082" {
			t.Errorf("Unexpected backend: %s", url)
		}
	}
}

func TestCircuitBreakerIntegration(t *testing.T) {
	backend, _ := NewRoundRobin([]string{"http://localhost:8081"}, 2, 1*time.Second)
	b, err := backend.NextBackend()
	if err != nil {
		t.Fatalf("NextBackend failed: %v", err)
	}

	// Simulate failures
	b.CircuitBreaker.RecordFailure()
	b.CircuitBreaker.RecordFailure()
	if b.CircuitBreaker.IsAvailable() {
		t.Error("Circuit breaker should be open after reaching failure threshold")
	}

	// Wait for cooldown
	time.Sleep(2 * time.Second)
	if !b.CircuitBreaker.IsAvailable() {
		t.Error("Circuit breaker should be half-open after cooldown")
	}

	// Test recovery
	b.CircuitBreaker.RecordSuccess()
	if !b.CircuitBreaker.IsAvailable() {
		t.Error("Circuit breaker should be closed after success in half-open state")
	}
}
