package webrtc

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"snapshell/internal/capture"
	"snapshell/internal/render"

	"github.com/pion/webrtc/v4"
	"gocv.io/x/gocv"
)

const (
	signalDir  = "/tmp/webrtc-signals"
	offerFile  = "/tmp/webrtc-signals/offer.json"
	answerFile = "/tmp/webrtc-signals/answer.json"
)

type SignalData struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func ensureSignalDir() {
	os.MkdirAll(signalDir, 0o755)
}

func RunAutoOffer() {
	ensureSignalDir()

	peer_connection, err := CreatePeerConnection()
	if err != nil {
		log.Fatal("Failed to create peer connection:", err)
	}

	peer_connection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Connection state: %s\n", s.String())
	})

	peer_connection.OnDataChannel(func(dc *webrtc.DataChannel) {
		fmt.Println("Received data channel (fallback)")
		dc.OnOpen(func() {
			fmt.Println("Data channel opened (fallback)")
		})
	})

	data_channel, err := peer_connection.CreateDataChannel("ascii", nil)
	if err != nil {
		log.Fatal("Failed to create data channel:", err)
	}

	data_channel.OnOpen(func() {
		fmt.Println("✅ Data channel opened! Starting video transmission...")
		webcam, _ := capture.OpenWebCam()
		defer webcam.Close()

		// Set webcam properties for better performance
		webcam.SetProperty(gocv.VideoCaptureFPS, 10)
		webcam.SetProperty(gocv.VideoCaptureFrameWidth, 640)
		webcam.SetProperty(gocv.VideoCaptureFrameHeight, 480)

		for {
			frame, err := webcam.ReadFrame()
			if err != nil {
				log.Println("Failed to read frame:", err)
				continue
			}

			ascii := render.ConvertFrameToASCII(frame)
			_ = data_channel.SendText(ascii)
			// avoid leaking mats
			frame.Close()
			time.Sleep(100 * time.Millisecond) // 10 FPS
		}
	})

	// NEW: also render anything we receive from the peer
	data_channel.OnMessage(func(msg webrtc.DataChannelMessage) {
		render.ClearTerminal()
		fmt.Print(string(msg.Data))
	})

	data_channel.OnClose(func() {
		fmt.Println("Data channel closed")
	})

	var iceCandidates []string

	peer_connection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			enc, _ := Encode(candidate.ToJSON())
			iceCandidates = append(iceCandidates, enc)
			fmt.Println("ICE:", enc)
		}
	})

	offer, _ := peer_connection.CreateOffer(nil)
	peer_connection.SetLocalDescription(offer)

	// Wait for ICE candidates
	<-webrtc.GatheringCompletePromise(peer_connection)

	// Save offer to file
	offerData := SignalData{
		Type: "offer",
		Data: map[string]interface{}{
			"sdp": peer_connection.LocalDescription(),
			"ice": iceCandidates,
		},
	}

	offerBytes, _ := json.Marshal(offerData)
	os.WriteFile(offerFile, offerBytes, 0o644)
	fmt.Printf("✅ Offer saved to %s\n", offerFile)
	fmt.Println("Waiting for answer...")

	// Wait for answer file
	for {
		if _, err := os.Stat(answerFile); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Read answer
	answerBytes, _ := os.ReadFile(answerFile)
	var answerData SignalData
	json.Unmarshal(answerBytes, &answerData)

	answerMap := answerData.Data.(map[string]interface{})
	answerSDP := answerMap["sdp"].(map[string]interface{})

	var session_description webrtc.SessionDescription
	session_description.Type = webrtc.SDPTypeAnswer
	session_description.SDP = answerSDP["sdp"].(string)

	peer_connection.SetRemoteDescription(session_description)

	// Add ICE candidates from answer
	if answerICE, ok := answerMap["ice"].([]interface{}); ok {
		for _, iceStr := range answerICE {
			var iceCandidate webrtc.ICECandidateInit
			Decode(iceStr.(string), &iceCandidate)
			peer_connection.AddICECandidate(iceCandidate)
		}
	}

	// Clean up
	os.Remove(answerFile)

	fmt.Println("✅ Connection established!")
	select {}
}
