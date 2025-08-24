// SnapShell CLI
// Made with ❤️ by Saswat Samal (https://github.com/saswatsam786)
// License: MIT
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/saswatsam786/snapshell/internal/webrtc"
)

func getDefaultServer() string {
	// Check environment variable first
	if env := os.Getenv("SNAPSHELL_SERVER"); env != "" {
		return env
	}
	// Default to localhost for development
	return "http://localhost:8080"
}

func main() {
	autoOfferSignaled := flag.Bool("signaled-o", false, "Start as offerer (caller) - signaling server mode")
	autoAnswerSignaled := flag.Bool("signaled-a", false, "Start as answerer (callee) - signaling server mode")
	server := flag.String("server", getDefaultServer(), "Signaling server base URL (default: SNAPSHELL_SERVER env var or http://localhost:8080)")
	room := flag.String("room", "", "Meeting ID (room)")
	clientID := flag.String("id", "", "Client ID (optional; random if empty)")
	flag.Parse()

	if (*autoOfferSignaled || *autoAnswerSignaled) && *room == "" {
		fmt.Println("For signaled auto mode, provide --room (and optionally --id)")
		fmt.Printf("Using signaling server: %s\n", *server)
		os.Exit(1)
	}

	if *autoOfferSignaled {
		fmt.Println("Running as auto caller (signaling server)...")
		webrtc.RunAutoOfferSignaled(*server, *room, *clientID)
	} else if *autoAnswerSignaled {
		fmt.Println("Running as auto callee (signaling server)...")
		webrtc.RunAutoAnswerSignaled(*server, *room, *clientID)
	} else {
		fmt.Println("Usage:")
		fmt.Println("  Signaling server mode (recommended):")
		fmt.Println("    snapshell -signaled-o --room <id> [--id <client>]    # Start as caller")
		fmt.Println("    snapshell -signaled-a --room <id> [--id <client>]    # Join as answerer")
		fmt.Println("    # Server auto-detected from SNAPSHELL_SERVER env var or defaults to localhost:8080")
		fmt.Println("")
		fmt.Println("  Other modes:")
		fmt.Println("    snapshell -auto-o     # Auto caller (file signaling)")
		fmt.Println("    snapshell -auto-a     # Auto answerer (file signaling)")
		fmt.Println("    snapshell -o          # Manual caller")
		fmt.Println("    snapshell -a          # Manual answerer")
		fmt.Println("")
		fmt.Printf("  Current signaling server: %s\n", getDefaultServer())
		fmt.Println("")
		fmt.Println("  Made with ❤️ by Saswat Samal (https://github.com/saswatsam786)")
		os.Exit(1)
	}
}
