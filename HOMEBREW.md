# Homebrew Installation for SnapShell

## Quick Install (macOS/Linux)

```bash
# Add the tap
brew tap saswatsam786/snapshell https://github.com/saswatsam786/snapshell

# Install SnapShell
brew install snapshell
```

## Alternative: Install directly from formula

```bash
brew install https://raw.githubusercontent.com/saswatsam786/snapshell/main/Formula/snapshell.rb
```

## Usage

After installation, you can use SnapShell directly:

```bash
# Start video sharing session (offerer)
snapshell -signaled-o --room demo123 --server https://snapshell.onrender.com

# Join video session (answerer)
snapshell -signaled-a --room demo123 --server https://snapshell.onrender.com
```

## Requirements

- **macOS**: Homebrew will automatically install OpenCV and other dependencies
- **Webcam**: SnapShell requires a webcam to capture video
- **Go**: Automatically installed as a build dependency

## What it does

SnapShell converts your live webcam feed to ASCII art and shares it in real-time through WebRTC peer-to-peer connections. Perfect for terminal-based video calls!

## Troubleshooting

If you encounter issues:

1. **OpenCV not found**: Homebrew should handle this automatically
2. **Build errors**: Try `brew update && brew upgrade` first
3. **Webcam access**: Grant camera permissions when prompted

## Manual Build (if needed)

```bash
git clone https://github.com/saswatsam786/snapshell
cd snapshell
brew install opencv pkg-config
go build -o snapshell cmd/main.go
```
