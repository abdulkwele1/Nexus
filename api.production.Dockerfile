# Multi-stage build for production Go API
# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY nexus-api/go.mod nexus-api/go.sum ./
RUN go mod download

# Copy source code
COPY nexus-api/ ./

# Build the binary
# -ldflags="-w -s" strips debug info and reduces binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o nexus-api .

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install ca-certificates and wget for health checks
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user for security
RUN addgroup -g 1000 nexus && \
    adduser -D -u 1000 -G nexus nexus

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/nexus-api .

# Change ownership to non-root user
RUN chown nexus:nexus /app/nexus-api

# Switch to non-root user
USER nexus

# Expose the API port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthcheck || exit 1

# Run the application
CMD ["./nexus-api"]

