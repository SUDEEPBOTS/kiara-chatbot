# ==============================================================================
# Copyright (c) 2026 SUDEEPBOTS. All rights reserved.
#
# Project: Kiara Chatbot - High-Performance Telegram AI Assistant
# Repository: https://github.com/SUDEEPBOTS/kiara-chatbot
# License: MIT License
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.
# ==============================================================================

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
