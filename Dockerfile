# Multi-stage build for Go application
FROM golang:1.24-alpine AS builder

# Install system dependencies for OpenCV
RUN apk add --no-cache \
    build-base \
    cmake \
    opencv-dev \
    pkgconfig

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the applications
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o snapshell cmd/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o signaler cmd/signaler/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    opencv \
    ca-certificates

WORKDIR /root/

# Copy the built binaries
COPY --from=builder /app/snapshell .
COPY --from=builder /app/signaler .

# Expose signaling server port
EXPOSE 8080

# Default to running the signaling server
CMD ["./signaler"]
