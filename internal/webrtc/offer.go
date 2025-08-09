package webrtc

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"snapshell/internal/capture"
	"snapshell/internal/render"

	"github.com/pion/webrtc/v4"
)

func RunOffer() {
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
			time.Sleep(100 * time.Millisecond)
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

	// ✅ Wait for ICE candidates
	<-webrtc.GatheringCompletePromise(peer_connection)

	encOffer, _ := Encode(*peer_connection.LocalDescription())
	fmt.Println("OFFER:", encOffer)

	fmt.Println("\nPaste answer:")
	reader := bufio.NewReader(os.Stdin)
	answerEnc, _ := reader.ReadString('\n')
	answerEnc = strings.TrimSpace(answerEnc)

	var session_description webrtc.SessionDescription
	Decode(answerEnc, &session_description)
	peer_connection.SetRemoteDescription(session_description)

	fmt.Println("Paste ICE candidates from answerer (one per line, empty line to finish):")
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
