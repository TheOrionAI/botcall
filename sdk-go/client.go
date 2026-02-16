// BotCall Go SDK
// Lightweight client for AI agents to accept voice calls

package botcall

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client handles bot registration and call acceptance
type Client struct {
	AgentID         string
	DiscoveryURL    string
	AttestationToken string
	Endpoint        string
	
	// Internal state
	httpClient      *http.Client
	wsConn          *websocket.Conn
	registered      bool
	onCallHandler   func(*Call)
	onAudioHandler  func([]byte) []byte
	mu              sync.RWMutex
}

// Call represents an incoming call from a human
type Call struct {
	CallID     string
	HumanID    string
	StartedAt  time.Time
	client     *Client
}

// RegisterRequest sent to discovery server
type RegisterRequest struct {
	AgentID     string `json:"agent_id"`
	Endpoint    string `json:"endpoint"`
	Mode        string `json:"mode"`
	Attestation string `json:"attestation"`
}

// RegisterResponse from discovery server
type RegisterResponse struct {
	Confirmed bool   `json:"confirmed"`
	URL       string `json:"url,omitempty"`
	Status    string `json:"status"`
}

// NewClient creates a BotCall client
func NewClient(agentID, attestationToken string) *Client {
	return &Client{
		AgentID:          agentID,
		DiscoveryURL:     "http://localhost:8080", // Default
		AttestationToken: attestationToken,
		httpClient:       &http.Client{Timeout: 10 * time.Second},
	}
}

// SetDiscoveryURL customizes the discovery server URL
func (c *Client) SetDiscoveryURL(url string) *Client {
	c.DiscoveryURL = url
	return c
}

// Connect registers with discovery and starts listening
func (c *Client) Connect() error {
	// Determine our endpoint (public IP:port)
	// For now, use 0.0.0.0:9000 and let user configure
	if c.Endpoint == "" {
		c.Endpoint = "0.0.0.0:9000"
	}

	// Register with discovery server
	req := RegisterRequest{
		AgentID:     c.AgentID,
		Endpoint:    c.Endpoint,
		Mode:        "direct",
		Attestation: c.AttestationToken,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal register request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.DiscoveryURL+"/v1/register",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("register with discovery: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discovery returned %d", resp.StatusCode)
	}

	var result RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if !result.Confirmed {
		return fmt.Errorf("registration rejected")
	}

	c.mu.Lock()
	c.registered = true
	c.mu.Unlock()

	log.Printf("[BotCall] Registered as %s at %s", c.AgentID, c.Endpoint)
	return nil
}

// OnCall sets the handler for incoming calls
func (c *Client) OnCall(handler func(*Call)) {
	c.onCallHandler = handler
}

// HandleIncoming starts HTTP server for incoming calls
func (c *Client) HandleIncoming(addr string, handler func(http.ResponseWriter, *http.Request)) error {
	if addr != "" {
		c.Endpoint = addr
	}

	http.HandleFunc("/call", c.handleCall)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"agent":   c.AgentID,
			"version": "0.1.0",
		})
	})

	// Re-register with actual endpoint
	if err := c.Connect(); err != nil {
		return err
	}

	log.Printf("[BotCall] Listening for calls on %s", c.Endpoint)
	return http.ListenAndServe(c.Endpoint, nil)
}

// handleCall processes incoming call requests
func (c *Client) handleCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify this is a legitimate call (could check attestation here)
	var req struct {
		HumanID     string `json:"human_id"`
		Attestation string `json:"attestation"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	call := &Call{
		CallID:    fmt.Sprintf("call-%d", time.Now().Unix()),
		HumanID:   req.HumanID,
		StartedAt: time.Now(),
		client:    c,
	}

	log.Printf("[BotCall] Incoming call from %s", req.HumanID)

	// Respond immediately, handle call asynchronously
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "accepted",
		"call_id":   call.CallID,
		"webrtc":    true, // Signal to use WebRTC
	})

	// Trigger handler in goroutine
	if c.onCallHandler != nil {
		go c.onCallHandler(call)
	}
}

// StartKeepalive pings discovery server periodically
func (c *Client) StartKeepalive(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if err := c.Connect(); err != nil {
				log.Printf("[BotCall] Keepalive failed: %v", err)
			}
		}
	}()
}

// Close cleans up resources
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.registered = false
	if c.wsConn != nil {
		return c.wsConn.Close()
	}
	return nil
}

// IsRegistered returns registration status
func (c *Client) IsRegistered() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.registered
}

// GetPublicEndpoint returns the current endpoint
func (c *Client) GetPublicEndpoint() string {
	return c.Endpoint
}
