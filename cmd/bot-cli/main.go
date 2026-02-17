// BotCall CLI Bot - Reference implementation for AI agents
// Usage: go run main.go --agent-id=orion --endpoint=localhost:9000
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
)

var (
	agentID        = flag.String("agent-id", "orion", "Your bot's unique ID")
	discoveryURL   = flag.String("discovery", "http://localhost:8080", "BotCall discovery server URL")
	endpointAddr   = flag.String("endpoint", "localhost:9000", "Local HTTP endpoint (ip:port)")
	useLocaltunnel = flag.Bool("lt", false, "Use localtunnel to expose endpoint publicly")
)

func main() {
	flag.Parse()

	log.Printf("🤖 BotCall Bot starting...")
	log.Printf("   Agent ID: %s", *agentID)
	log.Printf("   Discovery: %s", *discoveryURL)

	publicEndpoint := *endpointAddr

	// Start localtunnel if requested
	if *useLocaltunnel {
		publicEndpoint = startLocaltunnelSimple(*endpointAddr)
		log.Printf("🌐 Public URL: %s", publicEndpoint)
	}

	// Register with discovery
	if err := registerWithDiscovery(publicEndpoint); err != nil {
		log.Fatalf("Failed to register: %v", err)
	}

	log.Printf("📞 Listening for calls on %s", *endpointAddr)
	log.Printf("   Humans can dial: %s/v1/lookup/%s", *discoveryURL, *agentID)

	// Keepalive ticker
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			registerWithDiscovery(publicEndpoint)
		}
	}()

	// HTTP server for incoming calls
	http.HandleFunc("/call", handleIncomingCall)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/ws", handleWebSocket)

	log.Fatal(http.ListenAndServe(*endpointAddr, nil))
}

func startLocaltunnelSimple(endpoint string) string {
	port := "9000"
	// Extract port from endpoint
	for i := len(endpoint) - 1; i >= 0; i-- {
		if endpoint[i] == ':' {
			port = endpoint[i+1:]
			break
		}
	}

	log.Printf("🔧 Start localtunnel: npx localtunnel --port %s", port)
	log.Printf("   (run in another terminal, then update --endpoint URL)")

	// Return placeholder - user will run lt manually
	return "https://YOUR-LT-URL.loca.lt"
}

func registerWithDiscovery(publicEndpoint string) error {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"agent_id":    *agentID,
		"endpoint":    publicEndpoint,
		"mode":        "direct",
		"attestation": "test-attestation",
	})

	resp, err := http.Post(*discoveryURL+"/v1/register", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discovery returned %d", resp.StatusCode)
	}

	var result struct {
		Confirmed bool   `json:"confirmed"`
		URL       string `json:"url"`
		Status    string `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Confirmed {
		log.Printf("✅ Registered: %s", result.Status)
	}
	return nil
}

type CallRequest struct {
	HumanID     string `json:"human_id"`
	Attestation string `json:"attestation"`
}

func handleIncomingCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("📲 Incoming call from: %s", req.HumanID)

	response := map[string]interface{}{
		"status":   "accepted",
		"call_id":  fmt.Sprintf("call-%d", time.Now().Unix()),
		"webrtc":   true,
		"agent_id": *agentID,
		"message":  fmt.Sprintf("Hello %s! I'm %s. How can I help you today?", req.HumanID, *agentID),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("   ✓ Call accepted: %s", response["call_id"])
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "ok",
		"agent":    *agentID,
		"version":  "0.1.0",
		"endpoint": *endpointAddr,
	})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("🔄 WebSocket connected")

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var incoming struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(msg, &incoming); err == nil && incoming.Type == "text" {
			log.Printf("💬 Received: %s", incoming.Text)

			response := map[string]interface{}{
				"type": "text",
				"text": fmt.Sprintf("Echo: %s", incoming.Text),
				"from": *agentID,
			}
			respJSON, _ := json.Marshal(response)
			conn.WriteMessage(msgType, respJSON)
		}
	}
}
