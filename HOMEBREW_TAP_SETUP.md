# Homebrew Tap Setup for SnapShell

## Steps to Create a Proper Homebrew Tap

### 1. Create a New Repository

Create a new GitHub repository named: `homebrew-snapshell`

**Repository name MUST follow the pattern**: `homebrew-<formula-name>`

### 2. Repository Structure

```
homebrew-snapshell/
├── Formula/
│   └── snapshell.rb
└── README.md
```

### 3. Upload the Formula

Copy the contents of `./Formula/snapshell.rb` to the new repository at `Formula/snapshell.rb`

### 4. Installation Commands

Once the tap repository is created, users can install with:

```bash
# Add the tap
brew tap saswatsam786/snapshell

# Install the formula
brew install snapshell
```

Or in one command:
```bash
brew install saswatsam786/snapshell/snapshell
```

## Alternative: Use GitHub Releases

If you prefer not to create a separate tap repository, we can:

1. Create GitHub releases with pre-built binaries
2. Update the install script to download from releases
3. Use the existing repository structure

## Current Formula Content

The formula that should be placed in `homebrew-snapshell/Formula/snapshell.rb`:

```ruby
class Snapshell < Formula
  desc "Real-time ASCII video sharing via WebRTC in your terminal"
  homepage "https://github.com/saswatsam786/snapshell"
  head "https://github.com/saswatsam786/snapshell.git", branch: "main"
  license "MIT"

  depends_on "go" => :build
  depends_on "pkg-config" => :build
  depends_on "opencv"

  def install
    ENV["CGO_ENABLED"] = "1"
    
    # Set OpenCV paths for gocv
    ENV["PKG_CONFIG_PATH"] = "#{Formula["opencv"].opt_lib}/pkgconfig"
    
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/main.go"
  end

  test do
    # Test that the binary runs and shows help
    output = shell_output("#{bin}/snapshell -h 2>&1", 2)
    assert_match "Usage", output
  end

  def caveats
    <<~EOS
      🎥 SnapShell is now installed!
      
      Quick Start:
        # Start video sharing session (offerer)
        snapshell -signaled-o --room demo123 --server https://snapshell.onrender.com
        
        # Join video session (answerer)  
        snapshell -signaled-a --room demo123 --server https://snapshell.onrender.com
      
      Note: SnapShell requires a webcam to capture video.
      The video will be converted to ASCII art and shared in real-time!
    EOS
  end
end
```
