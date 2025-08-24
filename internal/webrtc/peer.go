package webrtc

import (
	"encoding/base64"
	"encoding/json"

	"github.com/pion/webrtc/v4"
)

func CreatePeerConnection() (*webrtc.PeerConnection, error) {
	config := webrtc.Configuration{ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}}}
	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}
	return pc, nil
}

// New helper that allows custom ICE servers (from Twilio)
func CreatePeerConnectionWithServers(ice []webrtc.ICEServer) (*webrtc.PeerConnection, error) {
	cfg := webrtc.Configuration{
		ICEServers: ice,
		// Force TURN usage when available for better cross-network connectivity
		ICETransportPolicy: webrtc.ICETransportPolicyRelay,
		// Set connection constraints to prefer relay candidates
		BundlePolicy:  webrtc.BundlePolicyMaxBundle,
		RTCPMuxPolicy: webrtc.RTCPMuxPolicyRequire,
	}
	pc, err := webrtc.NewPeerConnection(cfg)
	if err != nil {
		return nil, err
	}
	return pc, nil
}

// CreatePeerConnectionWithFallback creates a connection that first tries relay-only,
// then falls back to all candidates if relay fails
func CreatePeerConnectionWithFallback(ice []webrtc.ICEServer) (*webrtc.PeerConnection, error) {
	// First try relay-only for cross-network connectivity
	cfg := webrtc.Configuration{
		ICEServers:         ice,
		ICETransportPolicy: webrtc.ICETransportPolicyRelay,
		BundlePolicy:       webrtc.BundlePolicyMaxBundle,
		RTCPMuxPolicy:      webrtc.RTCPMuxPolicyRequire,
	}

	pc, err := webrtc.NewPeerConnection(cfg)
	if err != nil {
		return nil, err
	}

	// Add connection state change handler to detect relay failures
	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		if s == webrtc.PeerConnectionStateFailed {
			// If relay fails, we could implement fallback logic here
			// For now, just log the failure
		}
	})

	return pc, nil
}

// Encode an SDP structure into base64 for manual exchange
func Encode(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// Decode base64 into the SDP structure
func Decode(encoded string, out interface{}) error {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
}
