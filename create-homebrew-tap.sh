#!/bin/bash

# Script to help create a Homebrew tap repository
# Run this script to get instructions for creating the tap

echo "🍺 Creating Homebrew Tap for SnapShell"
echo "======================================="
echo ""
echo "To fix the Homebrew installation issue, you need to create a proper tap repository:"
echo ""
echo "1️⃣  Create a new GitHub repository named: 'homebrew-snapshell'"
echo ""
echo "2️⃣  Add this structure to the repository:"
echo "   homebrew-snapshell/"
echo "   ├── Formula/"
echo "   │   └── snapshell.rb"
echo "   └── README.md"
echo ""
echo "3️⃣  Copy the formula content:"

cat << 'EOF'

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

EOF

echo ""
echo "4️⃣  After creating the tap repository, users can install with:"
echo "   brew tap saswatsam786/snapshell"
echo "   brew install snapshell"
echo ""
echo "   Or in one command:"
echo "   brew install saswatsam786/snapshell/snapshell"
echo ""
echo "📋 The formula content above has been saved to Formula/snapshell.rb in this repository."
echo "   Just copy it to your new 'homebrew-snapshell' repository!"
