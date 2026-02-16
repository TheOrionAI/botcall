package discovery

import (
	"testing"
	"time"
)

func TestDiscoveryStore(t *testing.T) {
	store := NewDiscoveryStore()

	// Test Register
	agent := &Agent{
		ID:       "test-agent",
		Endpoint: "192.168.1.100:9000",
		Mode:     "direct",
		Online:   true,
		LastSeen: time.Now(),
	}
	store.Register(agent)

	// Test Lookup
	found, ok := store.Lookup("test-agent")
	if !ok {
		t.Fatal("Expected to find agent")
	}
	if found.ID != "test-agent" {
		t.Errorf("Expected ID test-agent, got %s", found.ID)
	}

	// Test non-existent
	_, ok = store.Lookup("non-existent")
	if ok {
		t.Error("Expected not to find non-existent agent")
	}

	// Test Touch
	oldTime := found.LastSeen
	time.Sleep(1 * time.Millisecond)
	store.Touch("test-agent")
	
	updated, _ := store.Lookup("test-agent")
	if !updated.LastSeen.After(oldTime) {
		t.Error("Expected LastSeen to be updated")
	}

	// Test ListOnline
	online := store.ListOnline()
	if len(online) != 1 {
		t.Errorf("Expected 1 online agent, got %d", len(online))
	}
}
