# Go Load Balancer

A simple, modular HTTP load balancer written in Go, designed for high availability and extensibility.

## Features
- **Load Balancing Strategies**: Supports round-robin and random balancing.
- **Circuit Breaker**: Automatically isolates failing backends with configurable failure thresholds and cooldown periods.
- **Health Checks**: Periodic health checks to detect and recover backend servers.
- **Request Logging**: Detailed, color-coded logs for method, path, status code, response time, and backend selection.
- **Configurable**: Environment variable-based configuration with sensible defaults.
- **Graceful Shutdown**: Handles server termination gracefully.

## Prerequisites
- Go 1.23.4 or later
- Backend servers with a health check endpoint (e.g., `/health`)

## Setup
1. Clone the repository:
   ```bash
   git clone https://github.com/SlmnFz/GoLoad.git
   cd GoLoad
   ```

2. Set environment variables (or create a `.env` file):
   ```bash
   export LOAD_PORT=8080
   export LOAD_BACKENDS="http://localhost:8081,http://localhost:8082"
   export LOAD_BALANCER_TYPE="roundrobin" # or "random"
   export CIRCUIT_BREAKER_COOLDOWN="30s"
   export CIRCUIT_BREAKER_FAILURES="3"
   export HEALTH_CHECK_PATH="/health" # Optional, defaults to /health
   ```

3. Build and run:
   ```bash
   go build -o load ./cmd/
   ./load
   ```

4. Test the load balancer:
   ```bash
   curl http://localhost:8080
   ```

## Example Logs
Logs include request details and backend selection:
```
2025/04/20 10:00:00 Starting load balancer on via mode RoundRobin on port 8080 [http://localhost:8081,http://localhost:8082]
2025/04/20 10:00:01 [GET] /test ? -> http://localhost:8081 {200} ------------------- 1.234ms
2025/04/20 10:00:02 [POST] /api ? -> http://localhost:8082 {201} ------------------- 2.567ms
2025/04/20 10:00:03 Backend http://localhost:8081 recovered
```

## Testing
Run unit tests to verify balancer and circuit breaker functionality:
```bash
go test ./internal/balancer
```

## Development
- **Add Balancers**: Implement new strategies in `internal/balancer/`.
- **Extend Config**: Modify `internal/config/` for additional settings.
- **Customize Proxy**: Update `internal/proxy/` for proxy behavior.
- **Enhance Logging**: Adjust `internal/logging/` for custom log formats.

## Contributing
Contributions are welcome! Please:
1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/your-feature`).
3. Commit changes (`git commit -m 'Add your feature'`).
4. Push to the branch (`git push origin feature/your-feature`).
5. Open a pull request.

## License
MIT License. See [LICENSE](LICENSE) for details.