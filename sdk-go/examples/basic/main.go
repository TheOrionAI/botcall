// Basic BotCall bot example
// This bot registers with discovery and accepts voice calls

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/TheOrionAI/botcall-sdk-go"
)

func main() {
	log.Println("Starting BotCall bot...")

	// Create client
	bot := botcall.NewClient("orion", "your-botauth-token-here")
	
	// Optional: Use custom discovery server
	// bot.SetDiscoveryURL("https://discover.botcall.io")

	// Set public endpoint (must be reachable from internet after port forward)
	bot.Endpoint = "0.0.0.0:9000" // Will update after Connect()

	// Handle incoming calls
	bot.OnCall(func(call *botcall.Call) {
		log.Printf("📞 Call received from %s at %s", 
			call.HumanID, 
			call.StartedAt.Format("15:04:05"))
		
		// TODO: Send TTS greeting
		// TODO: Receive audio from human
		// TODO: Process with STT
		// TODO: Respond with AI response
		
		log.Printf("Call %s handled", call.CallID)
	})

	// Start HTTP server and accept calls
	log.Println("🤖 Orion bot listening on :9000")
	log.Println("📡 Registering with discovery server...")
	
	// Optionally start keepalive to stay registered
	go bot.StartKeepalive(4 * time.Minute)

	// This blocks forever, handling HTTP requests
	if err := bot.HandleIncoming(":9000", nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
