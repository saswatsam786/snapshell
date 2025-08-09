#!/bin/bash

echo "=== WebRTC Process Status ==="
echo

# Check for auto caller
echo "🔍 Checking for auto caller (-auto-o):"
if pgrep -f "main -auto-o" > /dev/null; then
    echo "✅ Auto caller is RUNNING"
    ps aux | grep "main -auto-o" | grep -v grep
else
    echo "❌ Auto caller is NOT running"
fi
echo

# Check for auto answerer
echo "🔍 Checking for auto answerer (-auto-a):"
if pgrep -f "main -auto-a" > /dev/null; then
    echo "✅ Auto answerer is RUNNING"
    ps aux | grep "main -auto-a" | grep -v grep
else
    echo "❌ Auto answerer is NOT running"
fi
echo

# Check for manual caller
echo "🔍 Checking for manual caller (-o):"
if pgrep -f "main -o" > /dev/null; then
    echo "✅ Manual caller is RUNNING"
    ps aux | grep "main -o" | grep -v grep
else
    echo "❌ Manual caller is NOT running"
fi
echo

# Check for manual answerer
echo "🔍 Checking for manual answerer (-a):"
if pgrep -f "main -a" > /dev/null; then
    echo "✅ Manual answerer is RUNNING"
    ps aux | grep "main -a" | grep -v grep
else
    echo "❌ Manual answerer is NOT running"
fi
echo

# Check signal files
echo "📁 Checking signal files:"
if [ -f "/tmp/webrtc-signals/offer.json" ]; then
    echo "✅ Offer file exists"
    ls -la /tmp/webrtc-signals/
else
    echo "❌ No offer file found"
fi
echo

# Check network connections
echo "🌐 Checking network connections:"
echo "Active UDP connections (WebRTC uses UDP):"
lsof -i -P | grep UDP | grep -E "(main|go)" | head -5
echo

echo "=== Summary ==="
TOTAL_PROCESSES=$(pgrep -f "main -" | wc -l)
echo "Total WebRTC processes running: $TOTAL_PROCESSES"

if [ $TOTAL_PROCESSES -gt 0 ]; then
    echo "🎥 WebRTC video streaming is ACTIVE"
else
    echo "⏸️  No WebRTC processes running"
fi 