package webrtc

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"snapshell/internal/capture"
	"snapshell/internal/render"
	sig "snapshell/internal/signal"

	"github.com/pion/webrtc/v4"
	"gocv.io/x/gocv"
)

func randID() string {
	const alpha = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 10)
	for i := range b {
		b[i] = alpha[rand.Intn(len(alpha))]
	}
	return string(b)
}

func RunAutoOfferSignaled(server, room, clientID string) {
	if clientID == "" {
		clientID = "offer-" + randID()
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sg := sig.New(strings.TrimRight(server, "/"), room, clientID)
	role, err := sg.Join()
	if err != nil {
		log.Fatal("join:", err)
	}
	if role != "offer" {
		log.Fatalf("room %s already assigned offer to someone else (you got %q). Start as -auto-a instead.", room, role)
	}

	pc, err := CreatePeerConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Connection state: %s\n", s.String())
		if s == webrtc.PeerConnectionStateFailed ||
			s == webrtc.PeerConnectionStateDisconnected ||
			s == webrtc.PeerConnectionStateClosed {
			stop()
		}
	})

	// Receive remote ASCII (peer's video)
	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			render.ClearTerminal()
			fmt.Print(string(msg.Data))
		})
	})

	// Local data channel: send our ASCII frames and also receive remote (if peer uses this DC)
	dc, err := pc.CreateDataChannel("ascii", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer dc.Close()

	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		render.ClearTerminal()
		fmt.Print(string(msg.Data))
	})

	// Send local webcam frames after open
	dc.OnOpen(func() {
		fmt.Println("✅ Data channel opened (offer). Sending...")
		webcam, _ := capture.OpenWebCam()
		defer webcam.Close()

		webcam.SetProperty(gocv.VideoCaptureFPS, 10)
		webcam.SetProperty(gocv.VideoCaptureFrameWidth, 640)
		webcam.SetProperty(gocv.VideoCaptureFrameHeight, 480)

		t := time.NewTicker(100 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				frame, err := webcam.ReadFrame()
				if err != nil {
					continue
				}
				ascii := render.ConvertFrameToASCII(frame)
				frame.Close()
				_ = dc.SendText(ascii)
			}
		}
	})

	// Send local ICE to server
	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		b64, _ := Encode(c.ToJSON())
		_ = sg.PostICE("offer", b64)
	})

	// Offer
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		log.Fatal(err)
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		log.Fatal(err)
	}
	<-webrtc.GatheringCompletePromise(pc)
	if err := sg.PostOffer(pc.LocalDescription().SDP); err != nil {
		log.Fatal(err)
	}

	// Subscribe to ICE destined to offer
	sseResp, err := sg.SubscribeICE("offer", func(cand string) {
		var init webrtc.ICECandidateInit
		if err := Decode(cand, &init); err == nil {
			_ = pc.AddICECandidate(init)
		}
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sseResp.Body.Close()

	// Wait for answer
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(700 * time.Millisecond):
			if sdp, ok, _ := sg.GetAnswer(); ok {
				ans := webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: sdp}
				if err := pc.SetRemoteDescription(ans); err != nil {
					log.Fatal(err)
				}
				fmt.Println("✅ Remote answer set")
				goto RUN
			}
		}
	}

RUN:
	<-ctx.Done()
}

func RunAutoAnswerSignaled(server, room, clientID string) {
	if clientID == "" {
		clientID = "answer-" + randID()
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sg := sig.New(strings.TrimRight(server, "/"), room, clientID)
	role, err := sg.Join()
	if err != nil {
		log.Fatal("join:", err)
	}
	if role != "answer" {
		log.Fatalf("room %s expects you as answer; server gave role %q. Start as -auto-o if you're first.", room, role)
	}

	pc, err := CreatePeerConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Connection state: %s\n", s.String())
		if s == webrtc.PeerConnectionStateFailed ||
			s == webrtc.PeerConnectionStateDisconnected ||
			s == webrtc.PeerConnectionStateClosed {
			stop()
		}
	})

	// When the caller's DC arrives, render and also send our video
	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		fmt.Println("✅ Data channel received")
		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			render.MoveCursorToTop()
			fmt.Print(string(msg.Data))
		})
		dc.OnOpen(func() {
			fmt.Println("✅ DC opened (answer). Sending...")
			render.HideCursor()
			render.ClearTerminal()

			webcam, _ := capture.OpenWebCam()
			defer webcam.Close()

			webcam.SetProperty(gocv.VideoCaptureFPS, 10)
			webcam.SetProperty(gocv.VideoCaptureFrameWidth, 640)
			webcam.SetProperty(gocv.VideoCaptureFrameHeight, 480)

			t := time.NewTicker(100 * time.Millisecond)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					frame, err := webcam.ReadFrame()
					if err != nil {
						continue
					}
					ascii := render.ConvertFrameToASCII(frame)
					frame.Close()
					_ = dc.SendText(ascii)
				}
			}
		})
	})

	// Send local ICE to server
	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		b64, _ := Encode(c.ToJSON())
		_ = sg.PostICE("answer", b64)
	})

	// Subscribe to ICE destined to answer (from offerer)
	sseResp, err := sg.SubscribeICE("answer", func(cand string) {
		var init webrtc.ICECandidateInit
		if err := Decode(cand, &init); err == nil {
			_ = pc.AddICECandidate(init)
		}
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sseResp.Body.Close()

	// Wait for offer, then answer
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(700 * time.Millisecond):
			if sdp, ok, _ := sg.GetOffer(); ok {
				off := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: sdp}
				if err := pc.SetRemoteDescription(off); err != nil {
					log.Fatal(err)
				}

				answer, err := pc.CreateAnswer(nil)
				if err != nil {
					log.Fatal(err)
				}
				if err := pc.SetLocalDescription(answer); err != nil {
					log.Fatal(err)
				}
				<-webrtc.GatheringCompletePromise(pc)
				if err := sg.PostAnswer(pc.LocalDescription().SDP); err != nil {
					log.Fatal(err)
				}
				fmt.Println("✅ Posted answer")
				goto RUN
			}
		}
	}

RUN:
	<-ctx.Done()
}
