# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Use USTC mirror for faster and more reliable package downloads in China
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk update --no-cache

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o mbmanager ./cmd/server

# Frontend build stage
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web

# Use USTC mirror for faster and more reliable package downloads in China
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk update --no-cache

# Copy frontend files
COPY web/package*.json ./
RUN npm install

COPY web/ ./
RUN npm run build

# Runtime stage
FROM alpine:3.20

# Use USTC mirror for faster and more reliable package downloads in China
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

# Install runtime dependencies
RUN apk update && apk add --no-cache \
    ca-certificates \
    tzdata \
    mariadb-client \
    mariadb-connector-c \
    wget \
    bash

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/mbmanager .

# Copy web assets from frontend builder
COPY --from=frontend-builder /app/web/dist ./web/dist
COPY web/templates ./web/templates

# Create necessary directories
RUN mkdir -p /app/data/backups /app/data/db /app/logs

# Set timezone
ENV TZ=Asia/Shanghai

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./mbmanager"]
