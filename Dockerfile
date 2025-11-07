# syntax=docker/dockerfile:1.7

# Build stage
FROM golang:1.24-alpine AS builder

# Labels
LABEL org.opencontainers.image.title="DCGM Mapper"
LABEL org.opencontainers.image.description="GPU process to PID mapper for NVIDIA GPUs"
LABEL org.opencontainers.image.source="https://github.com/yourusername/dcgm-mapper"
LABEL org.opencontainers.image.licenses="MIT"


# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod ./

# Download dependencies (will create go.sum if it doesn't exist)
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/dcgm-mapper .

FROM alpine:latest as runner

# Copy binary from builder
COPY --from=builder /build/bin/dcgm-mapper /usr/local/bin/dcgm-mapper

# Create directory for mapping files
RUN mkdir -p /var/lib/dcgm-mapper

# Set default command
ENTRYPOINT ["/usr/local/bin/dcgm-mapper"]
CMD ["-daemon", "-interval", "30s", "-dir", "/var/lib/dcgm-mapper"]

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD test -d /var/lib/dcgm-mapper || exit 1


