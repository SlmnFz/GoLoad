package balancer

import (
	"errors"
	"net/url"
	"sync/atomic"
	"time"

	"load/internal/pattern"

	"golang.org/x/sync/errgroup"
)

// RoundRobin implements a round-robin load balancer using atomic operations.
type RoundRobin struct {
	Backends []*Backend
	Current  uint64
}

// NewRoundRobin initializes a round-robin balancer with the given backend addresses.
func NewRoundRobin(backendAddrs []string, failureThreshold int32, cooldown time.Duration) (*RoundRobin, error) {
	if len(backendAddrs) == 0 {
		return nil, errors.New("no backends provided")
	}

	backends := make([]*Backend, 0, len(backendAddrs))
	var g errgroup.Group
	for _, addr := range backendAddrs {
		addr := addr // Capture range variable
		g.Go(func() error {
			u, err := url.Parse(addr)
			if err != nil {
				return err
			}
			backends = append(backends, &Backend{
				URL:            u,
				CircuitBreaker: pattern.New(failureThreshold, cooldown),
			})
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &RoundRobin{
		Backends: backends,
		Current:  0,
	}, nil
}

// NextBackend returns the next available backend in the round-robin sequence.
func (rr *RoundRobin) NextBackend() (*Backend, error) {
	if len(rr.Backends) == 0 {
		return nil, errors.New("no available backends")
	}

	attempts := len(rr.Backends)
	for attempts > 0 {
		current := atomic.AddUint64(&rr.Current, 1) - 1
		index := int(current % uint64(len(rr.Backends)))
		backend := rr.Backends[index]
		if backend.CircuitBreaker.IsAvailable() {
			return backend, nil
		}
		attempts--
	}

	return nil, errors.New("no available backends")
}
