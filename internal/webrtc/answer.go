package webrtc

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/saswatsam786/snapshell/internal/capture"
	"github.com/saswatsam786/snapshell/internal/render"

	"github.com/pion/webrtc/v4"
)

func RunAnswer() {
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
			render.ClearTerminal()
			fmt.Print(string(msg.Data))
		})

		// NEW: also send our local webcam frames
		dc.OnOpen(func() {
			fmt.Println("✅ Data channel opened (callee). Starting local send...")
			webcam, _ := capture.OpenWebCam()
			defer webcam.Close()

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

	fmt.Println("Paste offer:")
	reader := bufio.NewReader(os.Stdin)
	offerEnc, _ := reader.ReadString('\n')
	offerEnc = strings.TrimSpace(offerEnc)

	var offer webrtc.SessionDescription
	Decode(offerEnc, &offer)
	peer_connection.SetRemoteDescription(offer)

	answer, _ := peer_connection.CreateAnswer(nil)
	peer_connection.SetLocalDescription(answer)

	// ✅ Wait for ICE candidates
	<-webrtc.GatheringCompletePromise(peer_connection)

	ansEnc, _ := Encode(*peer_connection.LocalDescription())
	fmt.Println("ANSWER:", ansEnc)

	fmt.Println("\nPaste ICE candidates from offerer (one per line, empty line to finish):")
	for {
		iceLine, _ := reader.ReadString('\n')
		iceLine = strings.TrimSpace(iceLine)
		if iceLine == "" {
			break
		}
		if strings.HasPrefix(iceLine, "ICE: ") {
			iceLine = strings.TrimPrefix(iceLine, "ICE: ")
		}
		var iceCandidate webrtc.ICECandidateInit
		Decode(iceLine, &iceCandidate)
		peer_connection.AddICECandidate(iceCandidate)
	}

	select {}
}
