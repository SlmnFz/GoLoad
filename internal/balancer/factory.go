package balancer

import (
	"fmt"
	"time"
)

// NewBalancer creates a balancer based on the specified type.
func NewBalancer(balancerType BalancerType, backendAddrs []string, failureThreshold int32, cooldown time.Duration) (Balancer, error) {
	switch balancerType {
	case RandomBalancerType:
		return NewRandom(backendAddrs, failureThreshold, cooldown)
	case RoundRobinBalancerType:
		return NewRoundRobin(backendAddrs, failureThreshold, cooldown)
	default:
		return nil, fmt.Errorf("unknown balancer type: %s", balancerType)
	}
}
