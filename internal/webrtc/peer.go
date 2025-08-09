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
