# -----------------------------------------------------------------------------
# Stage 1: Build binary
# -----------------------------------------------------------------------------
FROM golang:alpine AS builder

WORKDIR /src

# Install git and build essentials
RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Compile optimized static Go binary (stripped symbols for minimal footprint)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /src/bin/kiara cmd/bot/main.go

# -----------------------------------------------------------------------------
# Stage 2: Minimal runtime image
# -----------------------------------------------------------------------------
FROM alpine:latest

WORKDIR /app

# Install FFmpeg (required for video sticker frame extraction & webp conversion),
# CA certificates (for HTTPS/MTProto TLS connections), and timezone database
RUN apk add --no-cache \
    ffmpeg \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create dedicated non-root user and persistent storage directories
RUN addgroup -S appgroup && adduser -S appuser -G appgroup \
    && mkdir -p /app/data \
    && chown -R appuser:appgroup /app

# Copy compiled binary and persona file from builder stage
COPY --from=builder /src/bin/kiara /app/kiara
COPY --from=builder /src/system_prompt.txt /app/system_prompt.txt

# Set permissions
RUN chmod +x /app/kiara && chown appuser:appgroup /app/system_prompt.txt

# Run as non-root user
USER appuser

VOLUME ["/app/data"]

ENTRYPOINT ["/app/kiara"]
