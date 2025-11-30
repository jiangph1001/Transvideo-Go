# Build stage
FROM golang:1.21-alpine AS builder

# Install git for fetching dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o video-transcoder main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates docker-cli

# Create non-root user
RUN addgroup -g 1000 transcoder && \
    adduser -D -u 1000 -G transcoder transcoder

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/video-transcoder .
COPY --from=builder /app/static ./static

# Create directories for input/output (用于本机测试)
RUN mkdir -p /input /output && \
    chown -R transcoder:transcoder /app /input /output

# Switch to non-root user
USER transcoder

# Expose port
EXPOSE 8080

# Run the application
CMD ["./video-transcoder"]
