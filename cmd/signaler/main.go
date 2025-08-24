// SnapShell Signaling Server
// Made with ❤️ by Saswat Samal (https://github.com/saswatsam786)
// License: MIT
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

var (
	ctx = context.Background()
	rdb *redis.Client
)

const ttl = 15 * time.Minute

// Keys / channels
func kRoles(id string) string         { return "room:" + id + ":roles" }       // hash: clientID -> offer|answer
func kOfferSDP(id string) string      { return "room:" + id + ":offer" }       // string
func kAnswerSDP(id string) string     { return "room:" + id + ":answer" }      // string
func kICEList(id, side string) string { return "room:" + id + ":ice:" + side } // list backlog for side (offer|answer)
func chICE(id, to string) string      { return "chan:" + id + ":ice:" + to }   // pubsub for to=offer|answer

type joinReq struct {
	ClientID string `json:"clientId"`
}
type sdpReq struct {
	SDP string `json:"sdp"`
}
type iceReq struct {
	Candidate string `json:"candidate"`
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func joinRoom(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req joinReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ClientID == "" {
		http.Error(w, "missing clientId", http.StatusBadRequest)
		return
	}

	rolesKey := kRoles(id)
	roles, _ := rdb.HGetAll(ctx, rolesKey).Result()

	// already joined?
	if role, ok := roles[req.ClientID]; ok {
		rdb.Expire(ctx, rolesKey, ttl)
		writeJSON(w, map[string]string{"role": role})
		return
	}

	// assign role: first -> offer, second -> answer, else full
	haveOffer, haveAnswer := false, false
	for _, role := range roles {
		if role == "offer" {
			haveOffer = true
		}
		if role == "answer" {
			haveAnswer = true
		}
	}
	if haveOffer && haveAnswer {
		http.Error(w, "room full", http.StatusConflict)
		return
	}
	role := "offer"
	if haveOffer {
		role = "answer"
	}

	if err := rdb.HSet(ctx, rolesKey, req.ClientID, role).Err(); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	rdb.Expire(ctx, rolesKey, ttl)
	writeJSON(w, map[string]string{"role": role})
}

func postOffer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	clientID := r.URL.Query().Get("clientId")
	var req sdpReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SDP == "" {
		http.Error(w, "bad sdp", 400)
		return
	}
	role, _ := rdb.HGet(ctx, kRoles(id), clientID).Result()
	if role != "offer" {
		http.Error(w, "forbidden", 403)
		return
	}
	if err := rdb.Set(ctx, kOfferSDP(id), req.SDP, ttl).Err(); err != nil {
		http.Error(w, "server", 500)
		return
	}
	// reset answer's backlog because it will consume fresh ICE from offer
	rdb.Del(ctx, kICEList(id, "offer"))
	writeJSON(w, map[string]string{"ok": "1"})
}

func getOffer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	val, err := rdb.Get(ctx, kOfferSDP(id)).Result()
	if err == redis.Nil {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "server", 500)
		return
	}
	writeJSON(w, map[string]string{"sdp": val})
}

func postAnswer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	clientID := r.URL.Query().Get("clientId")
	var req sdpReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SDP == "" {
		http.Error(w, "bad sdp", 400)
		return
	}
	role, _ := rdb.HGet(ctx, kRoles(id), clientID).Result()
	if role != "answer" {
		http.Error(w, "forbidden", 403)
		return
	}
	if err := rdb.Set(ctx, kAnswerSDP(id), req.SDP, ttl).Err(); err != nil {
		http.Error(w, "server", 500)
		return
	}
	// reset offer's backlog because it will consume fresh ICE from answer
	rdb.Del(ctx, kICEList(id, "answer"))
	writeJSON(w, map[string]string{"ok": "1"})
}

func getAnswer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	val, err := rdb.Get(ctx, kAnswerSDP(id)).Result()
	if err == redis.Nil {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "server", 500)
		return
	}
	writeJSON(w, map[string]string{"sdp": val})
}

// POST /room/{id}/ice?from=offer|answer&clientId=...
func postICE(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	from := r.URL.Query().Get("from") // role of sender
	clientID := r.URL.Query().Get("clientId")
	if from != "offer" && from != "answer" {
		http.Error(w, "from must be offer|answer", 400)
		return
	}

	role, _ := rdb.HGet(ctx, kRoles(id), clientID).Result()
	if role != from {
		http.Error(w, "forbidden", 403)
		return
	}

	var req iceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Candidate == "" {
		http.Error(w, "bad candidate", 400)
		return
	}
	// store in sender backlog (useful for history) and publish to the opposite
	if err := rdb.RPush(ctx, kICEList(id, from), req.Candidate).Err(); err != nil {
		http.Error(w, "server", 500)
		return
	}
	rdb.Expire(ctx, kICEList(id, from), ttl)

	to := "offer"
	if from == "offer" {
		to = "answer"
	}
	if err := rdb.Publish(ctx, chICE(id, to), req.Candidate).Err(); err != nil {
		http.Error(w, "server", 500)
		return
	}
	writeJSON(w, map[string]string{"ok": "1"})
}

// GET /room/{id}/ice?to=offer|answer   (SSE stream)
func streamICE(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	to := r.URL.Query().Get("to")
	if to != "offer" && to != "answer" {
		http.Error(w, "to must be offer|answer", 400)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "no flush", 500)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// backlog first
	vals, _ := rdb.LRange(ctx, kICEList(id, to), 0, -1).Result()
	for _, c := range vals {
		fmt.Fprintf(w, "data: %s\n\n", c)
	}
	flusher.Flush()

	// live
	sub := rdb.Subscribe(ctx, chICE(id, to))
	defer sub.Close()

	ch := sub.Channel()
	tick := time.NewTicker(15 * time.Second)
	defer tick.Stop()

	for {
		select {
		case m := <-ch:
			if m == nil {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", m.Payload)
			flusher.Flush()
		case <-r.Context().Done():
			return
		case <-tick.C:
			// keep-alive
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

// GET /ice -> returns { "ice_servers": [ {urls, username, credential}, ... ] }
func getICEServers(w http.ResponseWriter, r *http.Request) {
	type iceOut struct {
		ICEServers []struct {
			URLs       []string `json:"urls"`
			Username   string   `json:"username,omitempty"`
			Credential string   `json:"credential,omitempty"`
		} `json:"ice_servers"`
	}

	type twilioToken struct {
		ICEServers []struct {
			URL        string      `json:"url"`  // sometimes present
			URLs       interface{} `json:"urls"` // string or []string
			Username   string      `json:"username,omitempty"`
			Credential string      `json:"credential,omitempty"`
		} `json:"ice_servers"`
	}

	w.Header().Set("Content-Type", "application/json")

	sid := os.Getenv("TWILIO_SID")
	auth := os.Getenv("TWILIO_AUTH")

	// Try Twilio first (if creds are set)
	if sid != "" && auth != "" {
		req, err := http.NewRequest(
			"POST",
			fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Tokens.json", sid),
			strings.NewReader(""), // no body required
		)
		if err == nil {
			req.SetBasicAuth(sid, auth)
			// Twilio expects form content-type even with empty body
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode/100 == 2 {
					var tw twilioToken
					if err := json.NewDecoder(resp.Body).Decode(&tw); err == nil {
						out := iceOut{}
						for _, s := range tw.ICEServers {
							// normalize urls -> []string
							var urls []string
							switch v := s.URLs.(type) {
							case string:
								if v != "" {
									urls = []string{v}
								}
							case []interface{}:
								for _, anyu := range v {
									if us, ok := anyu.(string); ok && us != "" {
										urls = append(urls, us)
									}
								}
							case []string:
								if len(v) > 0 {
									urls = append(urls, v...)
								}
							default:
								// fallback to "url" if "urls" not usable
								if s.URL != "" {
									urls = []string{s.URL}
								}
							}
							if len(urls) == 0 && s.URL != "" {
								urls = []string{s.URL}
							}
							if len(urls) == 0 {
								continue
							}

							out.ICEServers = append(out.ICEServers, struct {
								URLs       []string `json:"urls"`
								Username   string   `json:"username,omitempty"`
								Credential string   `json:"credential,omitempty"`
							}{
								URLs:       urls,
								Username:   s.Username,
								Credential: s.Credential,
							})
						}

						// ✅ Return exactly one JSON object
						_ = json.NewEncoder(w).Encode(out)
						return
					}
				} else {
					// Non-2xx from Twilio — fall through to fallback below
				}
			}
		}
	}

	// Fallback: public STUN + optional static TURN from env
	out := iceOut{
		ICEServers: []struct {
			URLs       []string `json:"urls"`
			Username   string   `json:"username,omitempty"`
			Credential string   `json:"credential,omitempty"`
		}{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	// Optional static TURN (if you set envs)
	if u := os.Getenv("TURN_URLS"); u != "" {
		out.ICEServers = append(out.ICEServers, struct {
			URLs       []string `json:"urls"`
			Username   string   `json:"username,omitempty"`
			Credential string   `json:"credential,omitempty"`
		}{
			URLs:       strings.Split(u, ","),
			Username:   os.Getenv("TURN_USERNAME"),
			Credential: os.Getenv("TURN_PASSWORD"),
		})
	}

	_ = json.NewEncoder(w).Encode(out)
}

func main() {
	// Load environment variables from .env.development
	if err := godotenv.Load(".env.development"); err != nil {
		log.Printf("Warning: Could not load .env.development: %v", err)
	}

	// REDIS_URL like: redis://:password@host:6379/0  (or leave empty for localhost)
	redisAddr := "127.0.0.1:6379"
	if env := os.Getenv("REDIS_ADDR"); env != "" {
		redisAddr = env
	}

	redisOpts := &redis.Options{Addr: redisAddr}

	// Support full Redis URL format
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		log.Printf("Using REDIS_URL: %s", redisURL)
		opts, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Fatal("Invalid REDIS_URL:", err)
		}
		log.Printf("Parsed Redis config - Addr: %s, Username: %s", opts.Addr, opts.Username)
		redisOpts = opts
	} else {
		log.Printf("No REDIS_URL found, using default: %s", redisAddr)
	}

	rdb = redis.NewClient(redisOpts)

	// Test Redis connection immediately
	log.Println("Testing Redis connection...")
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("❌ Redis connection failed: %v", err)
		log.Printf("Redis config: Addr=%s, Username=%s", redisOpts.Addr, redisOpts.Username)
	} else {
		log.Println("✅ Redis connection successful!")
	}

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		// Check Redis connection
		if err := rdb.Ping(ctx).Err(); err != nil {
			http.Error(w, "Redis unavailable", http.StatusServiceUnavailable)
			return
		}
		writeJSON(w, map[string]string{"status": "healthy", "service": "snapshell-signaler"})
	})

	// Root endpoint with service info
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{
			"service":   "SnapShell WebRTC Signaling Server",
			"version":   "1.0.0",
			"endpoints": "/room/{id}/join, /room/{id}/offer, /room/{id}/answer, /room/{id}/ice",
		})
	})

	mux.HandleFunc("POST /room/{id}/join", joinRoom)
	mux.HandleFunc("POST /room/{id}/offer", postOffer)
	mux.HandleFunc("GET /room/{id}/offer", getOffer)
	mux.HandleFunc("POST /room/{id}/answer", postAnswer)
	mux.HandleFunc("GET /room/{id}/answer", getAnswer)
	mux.HandleFunc("POST /room/{id}/ice", postICE)
	mux.HandleFunc("GET /room/{id}/ice", streamICE)
	mux.HandleFunc("GET /ice", getICEServers)

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	log.Println("signaler (redis) listening on", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
