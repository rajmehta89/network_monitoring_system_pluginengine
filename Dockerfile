# Use a Debian-based Go image
FROM golang:1.21

# Install libzmq (ZeroMQ)
RUN apt-get update && apt-get install -y libzmq3-dev pkg-config

# Set work directory
WORKDIR /app

# Copy Go plugin source
COPY . .

# Download dependencies
RUN go mod tidy

# Build binary
RUN go build -o pluginengine main.go
