# Build stage
FROM golang:1.21-alpine AS builder

# git: 拉依赖; gcc/musl-dev: go-sqlite3 需要 cgo
RUN apk add --no-cache git gcc musl-dev

# 国内构建走代理（可用 --build-arg 覆盖）
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY} \
    GOSUMDB=sum.golang.google.cn \
    CGO_ENABLED=1

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -trimpath -ldflags="-s -w" -o video-transcoder main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates docker-cli tzdata

# Create non-root user
RUN addgroup -g 1000 transcoder && \
    adduser -D -u 1000 -G transcoder transcoder

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/video-transcoder .
COPY --from=builder /app/static ./static

# Create directories for input/output and database persistence
RUN mkdir -p /input /output /app/data && \
    chown -R transcoder:transcoder /app /input /output

# 默认把 SQLite 放在持久化目录（compose 会挂 ./data）
ENV DB_PATH=/app/data/tasks.db

# Switch to non-root user
USER transcoder

# Expose port
EXPOSE 8080

# Run the application
CMD ["./video-transcoder"]
