package balancer

import (
	"net/url"
	"strings"

	"load/internal/pattern"
)

// Backend represents a backend server with its circuit breaker.
type Backend struct {
	URL            *url.URL
	CircuitBreaker *pattern.CircuitBreaker
}

// Balancer defines the interface for load balancing strategies.
type Balancer interface {
	NextBackend() (*Backend, error)
}

// BalancerType defines the type of load balancing strategy.
type BalancerType int

const (
	// RandomBalancerType represents the random load balancing strategy.
	RandomBalancerType BalancerType = iota
	// RoundRobinBalancerType represents the round-robin load balancing strategy.
	RoundRobinBalancerType
)

// String returns the string representation of the BalancerType.
func (bt BalancerType) String() string {
	switch bt {
	case RandomBalancerType:
		return "Random"
	case RoundRobinBalancerType:
		return "RoundRobin"
	default:
		return "Unknown"
	}
}

// ParseBalancerType converts a string to a BalancerType.
func ParseBalancerType(s string) (BalancerType, error) {
	switch strings.ToLower(s) {
	case "random":
		return RandomBalancerType, nil
	case "roundrobin":
		return RoundRobinBalancerType, nil
	default:
		return RandomBalancerType, nil
	}
}
