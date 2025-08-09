# SnapShell 📸

A real-time screen sharing application using WebRTC technology, built with Go. SnapShell allows you to share your screen/webcam feed across devices with low latency through WebRTC peer-to-peer connections.

## Features

- **Real-time Screen Sharing**: Share your screen or webcam feed with minimal latency
- **Multiple Connection Modes**: 
  - Signaling server mode (recommended for production)
  - File-based signaling (for local testing)
  - Manual mode (for development)
- **Cross-platform**: Works on Linux, macOS, and Windows
- **ASCII Preview**: Terminal-based preview of video feed
- **Redis-based Signaling**: Scalable signaling server with Redis backend

## Quick Start

### Prerequisites

- Go 1.24+
- OpenCV (for video capture)
- Redis (for signaling server)

### Installation

```bash
git clone https://github.com/yourusername/snapshell.git
cd snapshell
go mod download
```

### Building

```bash
# Build client
go build -o snapshell cmd/main.go

# Build signaling server
go build -o signaler cmd/signaler/main.go
```

### Usage

#### 1. Start the Signaling Server

```bash
# Make sure Redis is running on localhost:6379
redis-server

# Start the signaling server
./signaler
```

#### 2. Screen Sharing Session

**Caller (screen sharer):**
```bash
./snapshell -signaled-o --room myroom
```

**Viewer:**
```bash
./snapshell -signaled-a --room myroom
```

#### Alternative Modes

**File-based signaling (local testing):**
```bash
# Terminal 1 (caller)
./snapshell -auto-o

# Terminal 2 (answerer)
./snapshell -auto-a
```

**Manual mode (development):**
```bash
# Terminal 1
./snapshell -o

# Terminal 2
./snapshell -a
```

## Environment Variables

- `SNAPSHELL_SERVER`: Signaling server URL (default: `http://localhost:8080`)

## Architecture

- **Client**: Go application with WebRTC peer connections
- **Signaling Server**: HTTP server with Redis backend for WebRTC signaling
- **Video Capture**: OpenCV integration for screen/webcam capture
- **Rendering**: Terminal ASCII art and preview modes

## Development

```bash
# Run tests
go test ./...

# Format code
go fmt ./...

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o snapshell-linux cmd/main.go
GOOS=windows GOARCH=amd64 go build -o snapshell-windows.exe cmd/main.go
```

## License

MIT License - see LICENSE file for details.
