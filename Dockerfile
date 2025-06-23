FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with security flags
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags="-w -s" -o go-url-shortener .

# Final stage
FROM alpine:latest

# Install SQLite runtime dependencies and security updates
RUN apk --no-cache update && \
    apk --no-cache add ca-certificates libc6-compat sqlite-libs && \
    adduser -D -u 1000 appuser && \
    mkdir -p /data && \
    chown -R appuser:appuser /data

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/go-url-shortener .

# Create directories that need write access
RUN mkdir -p /tmp && chown -R appuser:appuser /tmp

# Use non-root user
USER 1000

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/metrics || exit 1

# Command to run the application
ENTRYPOINT ["./go-url-shortener"]
