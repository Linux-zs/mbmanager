# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o mbmanager ./cmd/server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    mysql-client \
    mariadb-connector-c \
    wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/mbmanager .

# Copy web assets
COPY --from=builder /app/web/dist ./web/dist
COPY --from=builder /app/web/templates ./web/templates

# Create necessary directories
RUN mkdir -p /data/backups /data/db /app/logs

# Set timezone
ENV TZ=Asia/Shanghai

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./mbmanager"]
