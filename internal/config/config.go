package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"load/internal/balancer"

	"github.com/joho/godotenv"
)

// Config holds the load balancer configuration.
type Config struct {
	Port                   string
	Backends               []string
	BalancerType           balancer.BalancerType
	CircuitBreakerFailures int32
	CircuitBreakerCooldown time.Duration
	HealthCheckPath        string 
}

// Load loads configuration from environment variables or .env file.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, errors.New("failed to load .env file: " + err.Error())
	}

	port := os.Getenv("LOAD_PORT")
	if port == "" {
		port = "8080" // Default port
	}

	backends := os.Getenv("LOAD_BACKENDS")
	if backends == "" {
		return nil, errors.New("LOAD_BACKENDS environment variable is required")
	}

	backendList := strings.Split(backends, ",")
	for i, b := range backendList {
		backendList[i] = strings.TrimSpace(b)
		if backendList[i] == "" {
			return nil, errors.New("empty backend address in LOAD_BACKENDS")
		}
	}

	balancerType, err := balancer.ParseBalancerType(os.Getenv("LOAD_BALANCER_TYPE"))
	if err != nil {
		return nil, err
	}

	failuresStr := os.Getenv("CIRCUIT_BREAKER_FAILURES")
	failures := int32(3) // Default: 3 failures
	if failuresStr != "" {
		f, err := strconv.ParseInt(failuresStr, 10, 32)
		if err != nil || f < 1 {
			return nil, errors.New("CIRCUIT_BREAKER_FAILURES must be a positive integer")
		}
		failures = int32(f)
	}

	cooldownStr := os.Getenv("CIRCUIT_BREAKER_COOLDOWN")
	cooldown := 30 * time.Second // Default: 30 seconds
	if cooldownStr != "" {
		c, err := time.ParseDuration(cooldownStr)
		if err != nil || c < time.Second {
			return nil, errors.New("CIRCUIT_BREAKER_COOLDOWN must be a valid duration (e.g., '30s')")
		}
		cooldown = c
	}

	healthCheckPath := os.Getenv("HEALTH_CHECK_PATH")
	if healthCheckPath == "" {
		healthCheckPath = "/health" // Default health check endpoint
	}

	return &Config{
		Port:                   port,
		Backends:               backendList,
		BalancerType:           balancerType,
		CircuitBreakerFailures: failures,
		CircuitBreakerCooldown: cooldown,
		HealthCheckPath:        healthCheckPath,
	}, nil
}
