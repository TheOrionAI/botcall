// BotCall Discovery Server
// Handles bot registration and human lookup for direct calling

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// Agent represents a registered bot
type Agent struct {
	ID          string    `json:"agent_id"`
	Endpoint    string    `json:"endpoint"`
	Mode        string    `json:"mode"` // direct, relay, nat-pending
	Attestation string    `json:"attestation"`
	Online      bool      `json:"online"`
	LastSeen    time.Time `json:"last_seen"`
}

// DiscoveryStore holds registered agents
type DiscoveryStore struct {
	mu     sync.RWMutex
	agents map[string]*Agent
}

func NewDiscoveryStore() *DiscoveryStore {
	return &DiscoveryStore{
		agents: make(map[string]*Agent),
	}
}

func (s *DiscoveryStore) Register(agent *Agent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agents[agent.ID] = agent
	log.Printf("Registered agent: %s at %s", agent.ID, agent.Endpoint)
}

func (s *DiscoveryStore) Lookup(agentID string) (*Agent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	agent, ok := s.agents[agentID]
	return agent, ok
}

func (s *DiscoveryStore) ListOnline() []*Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var online []*Agent
	for _, agent := range s.agents {
		if agent.Online && time.Since(agent.LastSeen) < 6*time.Minute {
			online = append(online, agent)
		}
	}
	return online
}

func (s *DiscoveryStore) Touch(agentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if agent, ok := s.agents[agentID]; ok {
		agent.LastSeen = time.Now()
		agent.Online = true
	}
}

// Server handles HTTP and WebSocket
type Server struct {
	store *DiscoveryStore
	upgrader websocket.Upgrader
}

func NewServer() *Server {
	return &Server{
		store: NewDiscoveryStore(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true }, // TODO: restrict in production
		},
	}
}

// RegisterRequest from bots
type RegisterRequest struct {
	AgentID     string `json:"agent_id"`
	Endpoint    string `json:"endpoint"`
	Mode        string `json:"mode"` // direct, relay
	Attestation string `json:"attestation"`
}

type RegisterResponse struct {
	Confirmed bool   `json:"confirmed"`
	URL       string `json:"url,omitempty"`
	Status    string `json:"status"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.AgentID == "" || req.Endpoint == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// TODO: Verify BotAuth attestation
	// For now, accept all

	agent := &Agent{
		ID:          req.AgentID,
		Endpoint:    req.Endpoint,
		Mode:        req.Mode,
		Attestation: req.Attestation,
		Online:      true,
		LastSeen:    time.Now(),
	}

	s.store.Register(agent)

	resp := RegisterResponse{
		Confirmed: true,
		URL:       fmt.Sprintf("wss://%s/v1/call/%s", r.Host, agent.ID),
		Status:    "online",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// LookupResponse for humans
type LookupResponse struct {
	Status            string `json:"status"`
	Endpoint          string `json:"endpoint,omitempty"`
	Mode              string `json:"mode,omitempty"`
	AttestationValid  bool   `json:"attestation_valid"`
	LastSeen          string `json:"last_seen,omitempty"`
	Error             string `json:"error,omitempty"`
}

func (s *Server) handleLookup(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Path[len("/v1/lookup/"):]
	if agentID == "" {
		http.Error(w, "Missing agent ID", http.StatusBadRequest)
		return
	}

	agent, ok := s.store.Lookup(agentID)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LookupResponse{
			Status: "offline",
			Error:  "Agent not found",
		})
		return
	}

	// Check if still online (5 min timeout)
	isOnline := agent.Online && time.Since(agent.LastSeen) < 5*time.Minute

	resp := LookupResponse{
		Status:           "online",
		Endpoint:         agent.Endpoint,
		Mode:             agent.Mode,
		AttestationValid: true, // TODO: verify
		LastSeen:         agent.LastSeen.Format(time.RFC3339),
	}
	
	if !isOnline {
		resp.Status = "offline"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"version": "0.1.0",
	})
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	agentID := r.URL.Query().Get("agent")
	if agentID == "" {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "missing agent ID"}`))
		return
	}

	log.Printf("WebSocket connected for agent: %s", agentID)

	// Keep connection alive and forward messages
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send heartbeat
			if err := conn.WriteMessage(websocket.TextMessage, []byte(`{"type": "ping"}`)); err != nil {
				log.Printf("Ping failed: %v", err)
				return
			}
			// Touch the agent
			s.store.Touch(agentID)
		}
	}
}

func (s *Server) listAgents(w http.ResponseWriter, r *http.Request) {
	agents := s.store.ListOnline()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agents": agents,
		"count":  len(agents),
	})
}

func main() {
	server := NewServer()

	// Routes
	http.HandleFunc("/v1/register", server.handleRegister)
	http.HandleFunc("/v1/lookup/", server.handleLookup)
	http.HandleFunc("/v1/ws", server.handleWebSocket)
	http.HandleFunc("/v1/agents", server.listAgents)
	http.HandleFunc("/health", server.handleHealth)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: nil,
	}

	go func() {
		log.Printf("BotCall Discovery Server starting on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	
	log.Println("Server stopped")
}
