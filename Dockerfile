# ===== Stage 1: Build Go Plugin (Linux) =====
FROM golang:1.21 AS build

# Install ZeroMQ dev libs + pkg-config for cgo builds
RUN apt-get update && apt-get install -y --no-install-recommends \
    libzmq3-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go.mod & go.sum first for dependency caching
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Linux binary with CGO enabled for ZeroMQ
RUN go build -o pluginengine main.go

# ===== Stage 2: Minimal Runtime with ZeroMQ runtime libs =====
FROM debian:bookworm-slim

# Install ZeroMQ runtime library for libzmq.so.5
RUN apt-get update && apt-get install -y --no-install-recommends \
    libzmq5 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy compiled binary from build stage
COPY --from=build /app/pluginengine .

# Ensure binary is executable and strip debug info to reduce size
RUN chmod +x pluginengine && strip pluginengine 2>/dev/null || true

# Default command to run the Go plugin
ENTRYPOINT ["./pluginengine"]
