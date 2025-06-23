# URL Shortener for Single-Node GitOps

A lightweight, fast Go-based URL shortener service tailored for deployment on [single-node-gitops](https://github.com/arnabdey73/single-node-gitops) platform.

## Features

- Shorten long URLs with a random 6-character code
- Redirect from short code to original URL
- Basic metrics (total requests, redirects by URL, errors)
- SQLite storage with persistence
- Prometheus-compatible `/metrics` endpoint
- Structured logs for Loki/Grafana
- Kubernetes deployment files with Kustomize
- ArgoCD application manifest for GitOps deployment

## Project Structure

```plaintext
go-url-shortener/
├── main.go                # Application entry point
├── handler/
│   └── url.go             # URL shortening and redirect handlers
├── storage/
│   ├── storage.go         # Storage interface
│   ├── memory.go          # In-memory storage implementation
│   └── sqlite.go          # SQLite storage implementation
├── k8s/                   # Kubernetes manifests
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── pvc.yaml
│   ├── ingress.yaml
│   ├── kustomization.yaml
│   └── argocd-application.yaml
├── Dockerfile             # Docker container definition
├── go.mod                 # Go module definition
├── Makefile               # Common development tasks
└── README.md              # This file
```

## Getting Started

### Prerequisites

- Go 1.21 or later
- SQLite (optional, for SQLite storage)
- Docker (for containerization)
- Kubernetes (for deployment)
- ArgoCD (for GitOps deployment)

### Running Locally

1. Clone the repository:

```bash
git clone https://github.com/your-username/go-url-shortener.git
cd go-url-shortener
```

1. Run with in-memory storage:

```bash
go run main.go --db memory
```

1. Or with SQLite storage:

```bash
go run main.go --db sqlite --db-path urls.db
```

1. Alternatively, use the Makefile:

```bash
# Run with in-memory storage
make run-memory

# Run with SQLite storage
make run-sqlite
```

## Deployment Instructions

### 1. Clone and Configure

Clone this repository to your GitOps server:

```bash
git clone https://github.com/your-username/go-url-shortener.git
cd go-url-shortener
```

Update the configuration with your server IP and Git repository URL:

```bash
chmod +x scripts/*.sh
./scripts/update-config.sh "your-server-ip" "https://github.com/your-username/go-url-shortener.git"
```

### 2. Build and Push Docker Image

```bash
# Build and push the Docker image to your local registry
./scripts/build-push.sh "your-server-ip"
```

### 3. Deploy with ArgoCD

```bash
# Apply the ArgoCD application
kubectl apply -f k8s/application.yaml
```

### 4. Monitor the Deployment

Use the included monitoring script:

```bash
# Check deployment status
./scripts/monitor.sh status

# View logs
./scripts/monitor.sh logs

# Check ArgoCD status
./scripts/monitor.sh argocd

# Port-forward to metrics endpoint
./scripts/monitor.sh metrics
```

### 5. Access the Application

The URL shortener will be available at:

- `http://url.your-server-ip.nip.io`

## API Usage

### Shorten a URL

```bash
curl -X POST http://url.your-server-ip.nip.io/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/very-long-url-that-needs-shortening"}'
```

Response:

```json
{
  "id": "abc123",
  "original": "https://example.com/very-long-url-that-needs-shortening",
  "created_at": "2025-06-23T12:34:56Z",
  "hits": 0
}
```

### Get Statistics

```bash
curl http://url.your-server-ip.nip.io/api/stats
```

Response:

```json
{
  "urls": [
    {
      "id": "abc123",
      "original": "https://example.com/very-long-url-that-needs-shortening",
      "created_at": "2025-06-23T12:34:56Z",
      "hits": 5
    }
  ]
}
```

### Access Prometheus Metrics

```bash
curl http://url.your-server-ip.nip.io/metrics
```

## Integration with Single-Node GitOps

This application is designed to integrate with the Single-Node GitOps platform:

- **Prometheus Integration**: ServiceMonitor automatically configures monitoring
- **Logging**: Structured logs work with the platform's Loki setup
- **Storage**: Uses the K3s Local-Path storage class
- **Ingress**: Compatible with the platform's ingress controller
- **Security**: Follows the platform's security practices

## Troubleshooting

If you encounter issues with the deployment, check:

1. Image availability in the registry: `./scripts/docker-registry.sh info`
2. ArgoCD synchronization status: `kubectl get applications -n argocd`
3. Application logs: `kubectl logs -f -l app=url-shortener -n url-shortener`
4. Platform health: `./scripts/health-check.sh`

## Docker

For local development, you can build and run the Docker container:

```bash
# Build
docker build -t url-shortener:latest .

# Run
docker run -p 8080:8080 url-shortener:latest
```

## Kubernetes Deployment

### Deployment Prerequisites

- Kubernetes cluster
- kubectl configured
- kustomize installed
- ArgoCD installed (for GitOps)

### Deploy with Kustomize

```bash
kubectl apply -k k8s/
```

### Deploy with ArgoCD

1. Edit `k8s/argocd-application.yaml` to point to your Git repository.
2. Apply the ArgoCD Application:

```bash
kubectl apply -f k8s/argocd-application.yaml
```

## Testing

The application includes comprehensive unit tests for all components and an integration test for verifying the application startup.

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage report
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./handler
go test ./storage
```

Using the Makefile:

```bash
# Run all tests with verbose output
make test

# Run tests with coverage report
make test-cover

# Run tests with HTML coverage report
make test-cover-html

# Run tests with race detector
make test-race
```

### Test Structure

- **Unit Tests**: Test individual packages in isolation
  - `storage_test.go`: Tests both memory and SQLite storage implementations
  - `handler/url_test.go`: Tests HTTP handler functionality
  - `handler/error_test.go`: Tests error handling scenarios
- **Integration Tests**: Test the application as a whole
  - `main_test.go`: Tests application startup, metrics endpoint, and full user flow
- **Benchmarks**: Performance tests
  - `storage/benchmark_test.go`: Benchmarks for storage operations

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run benchmarks for a specific package
go test -bench=. ./storage

# Run benchmarks with memory allocation stats
go test -bench=. -benchmem ./...
```

Using the Makefile:

```bash
# Run all benchmarks
make benchmark
```

### Adding New Tests

When adding new features, make sure to:

1. Add unit tests for any new functions or methods
2. Update existing tests if changing functionality
3. Ensure all tests pass before deploying

## Monitoring & Observability

### Prometheus Metrics

The application exposes the following metrics at `/metrics`:

- `url_shortener_redirects_total{url_id="abc123"}` - Total redirects by URL ID
- `url_shortener_shorten_requests_total` - Total shorten requests
- `url_shortener_errors_total` - Total errors
- Standard Go metrics (`go_*`)
- Process metrics (`process_*`)

### Structured Logging

The application uses structured logging with zap, compatible with Loki. Logs include:

- HTTP request details (method, path, status code)
- Latency
- Client IP
- User agent

## Enhancement Ideas

- Rate limit per IP (middleware)
- Basic auth (to restrict shortening)
- Swagger/OpenAPI docs
- Sentry or Grafana Tempo for tracing
- Custom URL aliases
- URL expiration
- QR code generation
- Analytics dashboard

## License

MIT
