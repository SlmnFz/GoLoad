package balancer

import (
	"errors"
	"math/rand"
	"net/url"
	"time"

	"load/internal/pattern"

	"golang.org/x/sync/errgroup"
)

// Random implements a random load balancer.
type Random struct {
	Backends []*Backend
}

// NewRandom initializes a random balancer with the given backend addresses.
func NewRandom(backendAddrs []string, failureThreshold int32, cooldown time.Duration) (*Random, error) {
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

	return &Random{
		Backends: backends,
	}, nil
}

// NextBackend returns a random available backend from the list of backends.
func (rr *Random) NextBackend() (*Backend, error) {
	if len(rr.Backends) == 0 {
		return nil, errors.New("no available backends")
	}

	attempts := len(rr.Backends)
	for attempts > 0 {
		index := rand.Intn(len(rr.Backends))
		backend := rr.Backends[index]
		if backend.CircuitBreaker.IsAvailable() {
			return backend, nil
		}
		attempts--
	}

	return nil, errors.New("no available backends")
}
