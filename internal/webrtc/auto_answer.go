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

func RunAutoAnswer() {
	ensureSignalDir()

	peer_connection, err := CreatePeerConnection()
	if err != nil {
		log.Fatal("Failed to create peer connection:", err)
	}

	peer_connection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Connection state: %s\n", s.String())
	})

	var iceCandidates []string

	peer_connection.OnDataChannel(func(dc *webrtc.DataChannel) {
		fmt.Println("✅ Data channel received")

		// receive & render remote frames
		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			render.MoveCursorToTop()
			fmt.Print(string(msg.Data))
		})

		// NEW: also send our local webcam frames
		dc.OnOpen(func() {
			fmt.Println("✅ Data channel opened (callee). Starting local send...")
			render.HideCursor()
			render.ClearTerminal()

			webcam, _ := capture.OpenWebCam()
			defer webcam.Close()

			// Set webcam properties for better performance
			webcam.SetProperty(gocv.VideoCaptureFPS, 10)
			webcam.SetProperty(gocv.VideoCaptureFrameWidth, 640)
			webcam.SetProperty(gocv.VideoCaptureFrameHeight, 480)

			ticker := time.NewTicker(100 * time.Millisecond) // ~10 FPS
			defer ticker.Stop()
			for range ticker.C {
				frame, err := webcam.ReadFrame()
				if err != nil {
					continue
				}
				ascii := render.ConvertFrameToASCII(frame)
				frame.Close()
				_ = dc.SendText(ascii)
			}
		})
	})

	peer_connection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			enc, _ := Encode(candidate.ToJSON())
			iceCandidates = append(iceCandidates, enc)
			fmt.Println("ICE:", enc)
		}
	})

	fmt.Println("Waiting for offer...")

	// Wait for offer file
	for {
		if _, err := os.Stat(offerFile); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Read offer
	offerBytes, _ := os.ReadFile(offerFile)
	var offerData SignalData
	json.Unmarshal(offerBytes, &offerData)

	offerMap := offerData.Data.(map[string]interface{})
	offerSDP := offerMap["sdp"].(map[string]interface{})

	var offer webrtc.SessionDescription
	offer.Type = webrtc.SDPTypeOffer
	offer.SDP = offerSDP["sdp"].(string)

	peer_connection.SetRemoteDescription(offer)

	// Add ICE candidates from offer
	if offerICE, ok := offerMap["ice"].([]interface{}); ok {
		for _, iceStr := range offerICE {
			var iceCandidate webrtc.ICECandidateInit
			Decode(iceStr.(string), &iceCandidate)
			peer_connection.AddICECandidate(iceCandidate)
		}
	}

	answer, _ := peer_connection.CreateAnswer(nil)
	peer_connection.SetLocalDescription(answer)

	// Wait for ICE candidates
	<-webrtc.GatheringCompletePromise(peer_connection)

	// Save answer to file
	answerData := SignalData{
		Type: "answer",
		Data: map[string]interface{}{
			"sdp": peer_connection.LocalDescription(),
			"ice": iceCandidates,
		},
	}

	answerBytes, _ := json.Marshal(answerData)
	os.WriteFile(answerFile, answerBytes, 0o644)
	fmt.Printf("✅ Answer saved to %s\n", answerFile)

	// Clean up
	os.Remove(offerFile)

	fmt.Println("✅ Connection established!")
	select {}
}
