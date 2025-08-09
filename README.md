# SnapShell 📸

A **real-time WebRTC-based terminal video sharing application** built with Go. SnapShell enables peer-to-peer webcam streaming with unique ASCII art rendering directly in your terminal - no GUI required!

## 🎯 What SnapShell Does

SnapShell creates a **bidirectional WebRTC connection** between two terminals, allowing real-time webcam feed sharing with ASCII art conversion. Think of it as "video calling for terminals" - perfect for remote pair programming, terminal demos, or just having fun with ASCII video art.

### Current Features ✅

- **🎥 Real-time Webcam Streaming**: Live video capture using OpenCV with configurable resolution (640x480 @ 10 FPS)
- **🎨 ASCII Art Conversion**: Advanced real-time video-to-ASCII conversion with dynamic terminal sizing
- **📡 Multiple Connection Modes**:
  - **Signaling Server Mode**: Production-ready with Redis backend (recommended)
  - **File-based Signaling**: Local testing using `/tmp/webrtc-signals/`
  - **Manual Mode**: Copy-paste SDP for development/debugging
- **🔄 Bidirectional Communication**: Both peers send AND receive video simultaneously
- **⚡ WebRTC Performance**: Direct peer-to-peer connection with ICE candidate optimization
- **🖥️ Smart Terminal Adaptation**: Automatic scaling based on terminal dimensions
- **🔧 Production Ready**: Docker, Heroku, and Redis deployment configurations

### Technical Implementation

```
┌─────────────────┐    WebRTC Data Channel    ┌─────────────────┐
│   Terminal A    │◄──────────────────────────►│   Terminal B    │
│                 │                            │                 │
│ Webcam → ASCII  │     Signaling Server       │ ASCII ← Webcam  │
│ ASCII ← Remote  │   (Redis + HTTP/SSE)       │ Remote → ASCII  │
└─────────────────┘                            └─────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.24+**
- **OpenCV 4.x** (for video capture)

  ```bash
  # macOS
  brew install opencv

  # Ubuntu/Debian
  sudo apt-get install libopencv-dev

  # Arch Linux
  sudo pacman -S opencv
  ```

- **Redis** (for signaling server)

  ```bash
  # macOS
  brew install redis

  # Ubuntu/Debian
  sudo apt-get install redis-server
  ```

### Installation

#### 🍺 **Homebrew (macOS)** ⭐ **Recommended**

**Official Homebrew Tap:**

```bash
# Add the tap and install
brew tap saswatsam786/snapshell
brew install --HEAD snapshell

# Ready to use!
snapshell --help
```

**Alternative - Local Formula:**

```bash
# Clone repository and install locally
git clone https://github.com/saswatsam786/snapshell.git
cd snapshell
brew install --build-from-source --HEAD --formula ./Formula/snapshell.rb
```

**Manual Dependencies + Build:**

```bash
# Install dependencies via Homebrew
brew install opencv pkg-config go

# Clone and build
git clone https://github.com/saswatsam786/snapshell.git
cd snapshell
go build -o snapshell cmd/main.go

# Optional: Install to system
sudo mv snapshell /usr/local/bin/
```

#### 🚀 **Quick Install Script (Linux)**

```bash
# One-line installation for Linux
curl -sSL https://raw.githubusercontent.com/saswatsam786/snapshell/main/install.sh | bash
```

#### 🔨 **Build from Source**

```bash
git clone https://github.com/saswatsam786/snapshell.git
cd snapshell

# Install dependencies
# macOS
brew install opencv pkg-config go

# Ubuntu/Debian
sudo apt-get install libopencv-dev libopencv-contrib-dev pkg-config golang-go

# Build the client
go build -o snapshell cmd/main.go

# Build the signaling server (optional - already deployed)
go build -o signaler cmd/signaler/main.go

# Ready to use!
./snapshell --help
```

## 🔧 Troubleshooting

### macOS Command Line Tools Issue

If you get an error about outdated Command Line Tools during Homebrew installation:

```
Error: Your Command Line Tools are too outdated.
Update them from Software Update in System Settings.
```

**Solution:**

```bash
# Option 1: Update via System Settings
# Go to System Settings > General > Software Update

# Option 2: Manual update (if System Settings doesn't show updates)
sudo rm -rf /Library/Developer/CommandLineTools
sudo xcode-select --install

# Option 3: Download manually from Apple
# Visit: https://developer.apple.com/download/all/
# Download "Command Line Tools for Xcode 16.2" or latest version
```

After updating Command Line Tools, try the installation again:

```bash
brew install --HEAD snapshell
```

### OpenCV Issues

If you encounter OpenCV-related build errors:

```bash
# Reinstall OpenCV
brew uninstall opencv
brew install opencv

# Verify pkg-config can find OpenCV
pkg-config --modversion opencv4

# Try installation again
brew install --HEAD snapshell
```

### General Build Issues

If the build fails:

```bash
# Clean Homebrew cache
brew cleanup

# Update Homebrew
brew update

# Try installation with verbose output
brew install --HEAD snapshell --verbose
```

## 🎮 Usage Modes

### 1. 🌐 Production Server Mode (Recommended)

**Use our deployed signaling server - no setup required!**

**Terminal 1 - First User (Caller):**

```bash
snapshell -signaled-o --room demo123 --server https://snapshell.onrender.com
# Webcam will start, ASCII video begins streaming
```

**Terminal 2 - Second User (Answerer):**

```bash
snapshell -signaled-a --room demo123 --server https://snapshell.onrender.com
# Connects to same room, bidirectional video starts
```

### 2. 🏠 Local Signaling Server (Development)

**If you want to run your own signaling server:**

**Terminal 1 - Start Redis & Signaling Server:**

```bash
redis-server &                    # Start Redis in background
./signaler                        # Start signaling server on :8080
```

**Terminal 2 - First User (Caller):**

```bash
snapshell -signaled-o --room demo123 --server http://localhost:8080
```

**Terminal 3 - Second User (Answerer):**

```bash
snapshell -signaled-a --room demo123 --server http://localhost:8080
```

### 3. 📁 File-based Signaling (Local Testing)

**Terminal 1 (Auto Caller):**

```bash
./snapshell -auto-o
# Creates offer file in /tmp/webrtc-signals/
```

**Terminal 2 (Auto Answerer):**

```bash
./snapshell -auto-a
# Reads offer, creates answer, establishes connection
```

### 4. 🔧 Manual Mode (Development)

**Terminal 1 (Manual Caller):**

```bash
./snapshell -o
# Displays offer SDP - copy and paste to answerer
```

**Terminal 2 (Manual Answerer):**

```bash
./snapshell -a
# Paste offer, displays answer SDP - copy back to caller
```

## 🏗️ Architecture & Design Philosophy

### Core Components

1. **WebRTC Client (`cmd/main.go`)**

   - Handles peer connection lifecycle
   - Manages ICE candidate exchange
   - Routes between different signaling modes

2. **Video Pipeline (`internal/capture/webcam.go` → `internal/render/ascii.go`)**

   - OpenCV webcam capture with configurable properties
   - Real-time ASCII conversion with intelligent scaling
   - Terminal-aware rendering (respects COLUMNS/LINES)

3. **Signaling Server (`cmd/signaler/main.go`)**

   - Redis-backed HTTP server for WebRTC signaling
   - Server-Sent Events (SSE) for real-time ICE delivery
   - Room-based session management

4. **Rendering Engine (`internal/render/`)**
   - Dynamic ASCII character mapping (10 intensity levels)
   - Terminal size detection and adaptation
   - Cross-platform terminal control

### Why ASCII? The Philosophy

- **Universal Compatibility**: Works in any terminal, SSH session, or console
- **Bandwidth Efficiency**: ASCII is incredibly lightweight vs. raw video
- **Retro Aesthetic**: Brings back the charm of terminal-based computing
- **Educational Value**: Demonstrates WebRTC concepts without video complexity
- **Remote Development**: Perfect for pair programming over low-bandwidth connections

## 🔧 Configuration

### Environment Variables

```bash
export SNAPSHELL_SERVER="https://your-signaler.herokuapp.com"  # Production signaler
export REDIS_URL="redis://user:pass@host:port/db"              # Redis connection
export REDIS_ADDR="localhost:6379"                             # Simple Redis address
export PORT="8080"                                              # Signaler port
```

### Runtime Options

```bash
# Connection modes
./snapshell -signaled-o --room <room> [--id <client>] [--server <url>]
./snapshell -signaled-a --room <room> [--id <client>] [--server <url>]
./snapshell -auto-o     # File signaling caller
./snapshell -auto-a     # File signaling answerer
./snapshell -o          # Manual caller
./snapshell -a          # Manual answerer

# Debug process status
./check_webrtc.sh       # Shows running processes and signal files
```

## 🚢 Deployment

### Docker Deployment

```bash
# Development with docker-compose
docker-compose up --build

# Production Docker build
docker build -t snapshell .
docker run -p 8080:8080 -e REDIS_URL=redis://redis:6379 snapshell
```

### Heroku Deployment

```bash
# Using Heroku Container Registry
heroku create your-snapshell-app --stack container
heroku addons:create heroku-redis:mini
git push heroku main
```

The signaling server will be available at `https://your-snapshell-app.herokuapp.com`

## 🛠️ Development & Debugging

### Project Structure

```
snapshell/
├── cmd/
│   ├── main.go          # Main client application
│   └── signaler/        # Redis-backed signaling server
├── internal/
│   ├── capture/         # OpenCV webcam integration
│   ├── render/          # ASCII conversion & terminal control
│   ├── signal/          # HTTP signaling client
│   └── webrtc/          # WebRTC peer management
├── pkg/utils/           # Shared utilities
├── Dockerfile           # Container build
├── docker-compose.yml   # Dev environment
└── check_webrtc.sh      # Process monitoring script
```

### Development Commands

```bash
# Format code
go fmt ./...

# Run tests
go test ./...

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o snapshell-linux cmd/main.go
GOOS=windows GOARCH=amd64 go build -o snapshell-windows.exe cmd/main.go
GOOS=darwin GOARCH=amd64 go build -o snapshell-macos cmd/main.go

# Live reload development
go install github.com/cosmtrek/air@latest
air  # Uses .air.toml configuration
```

### Debugging Tips

```bash
# Monitor WebRTC processes
./check_webrtc.sh

# Check signal files (file-based mode)
ls -la /tmp/webrtc-signals/
cat /tmp/webrtc-signals/offer.json

# Network debugging
lsof -i -P | grep UDP | grep -E "(snapshell|main)"

# Redis debugging (signaling server mode)
redis-cli monitor
redis-cli keys "room:*"
```

## 🎯 Future Roadmap

### Phase 2: Enhanced Features

- **Screen Capture**: Desktop/window sharing alongside webcam
- **Audio Support**: WebRTC audio channels with terminal visualizer
- **Multi-party**: Support for 3+ participants in a room
- **Recording**: Save ASCII sessions to files
- **Custom ASCII**: User-defined character sets and color palettes

### Phase 3: Advanced Capabilities

- **Bandwidth Adaptation**: Dynamic quality scaling based on connection
- **Mobile Support**: iOS/Android terminal apps
- **Web Interface**: Browser-based viewer for non-terminal users
- **Plugin System**: Extensible filters and effects
- **Authentication**: User accounts and private rooms

### Phase 4: Platform Features

- **Cloud Deployment**: One-click cloud instance deployment
- **CDN Integration**: Global signaling server distribution
- **Analytics**: Connection quality and usage metrics
- **API**: RESTful API for integration with other tools

## 🤝 Contributing

We welcome contributions! Areas of focus:

- **Performance**: Optimize ASCII conversion algorithms
- **Features**: Implement roadmap items
- **Platforms**: Windows/Linux compatibility testing
- **Documentation**: Usage examples and tutorials
- **Testing**: Unit tests and integration tests

## 📝 Technical Notes

### WebRTC Implementation Details

- Uses **Pion WebRTC v4** for Go-native implementation
- **Data Channels** for ASCII transmission (not video tracks)
- **STUN servers** for NAT traversal
- **ICE candidates** managed through Redis pub/sub

### Performance Characteristics

- **Latency**: ~100-200ms including ASCII conversion
- **Bandwidth**: ~1-5 KB/s per stream (vs. MB/s for raw video)
- **CPU Usage**: Moderate due to OpenCV processing
- **Memory**: Minimal frame buffering

### ASCII Conversion Algorithm

- **Grayscale conversion** using OpenCV color space transformation
- **Intelligent scaling** maintaining aspect ratio with terminal constraints
- **Character mapping** using 10-level intensity scale: ` .:=-+*#%@`
- **Dynamic sizing** based on terminal dimensions (COLUMNS/LINES)

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

**SnapShell** - Where retro meets real-time. Happy ASCII streaming! 🎥📺
