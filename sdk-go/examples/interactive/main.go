// Interactive BotCall bot
// Receives calls and responds with TTS

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/TheOrionAI/botcall-sdk-go"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║     BotCall Interactive Bot (Orion)      ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	// Get config from env or use defaults
	agentID := getEnv("BOTCALL_AGENT_ID", "orion")
	attestationToken := getEnv("BOTCALL_TOKEN", "demo-token")
	discoveryURL := getEnv("BOTCALL_DISCOVERY", "http://localhost:8080")
	listenAddr := getEnv("BOTCALL_ADDR", ":9000")

	log.Printf("🤖 Agent: %s", agentID)
	log.Printf("📡 Discovery: %s", discoveryURL)
	log.Printf("🔊 Listening on: %s", listenAddr)
	fmt.Println()

	// Create client
	bot := botcall.NewClient(agentID, attestationToken)
	bot.SetDiscoveryURL(discoveryURL)
	bot.Endpoint = listenAddr

	// Call Tracking
	activeCalls := make(map[string]*botcall.Call)

	// Handle incoming calls
	bot.OnCall(func(call *botcall.Call) {
		log.Printf("\n📞 INCOMING CALL from %s", call.HumanID)
		log.Printf("   Call ID: %s", call.CallID)
		log.Printf("   Started: %s", call.StartedAt.Format("15:04:05"))
		log.Printf("\n💬 Type your response or press Enter to hangup:")
		
		activeCalls[call.CallID] = call

		// In a real implementation:
		// - Accept WebRTC offer
		// - Set up Opus stream
		// - Play TTS greeting
		// - Receive audio from human
		// - Stream to STT
		// - Generate AI response
		// - Stream TTS response
	})

	// Start keepalive in background
	go bot.StartKeepalive(4 * time.Minute)
	log.Println("💓 Keepalive started (4 min interval)")

	// Interactive console in another goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		log.Println("\n⚡ Bot is running. Commands:")
		log.Println("   • Type message + Enter: Send text response")
		log.Println("   • status: Show registration status")
		log.Println("   • quit: Shut down")
		log.Println()

		for scanner.Scan() {
			line := scanner.Text()
			
			switch line {
			case "quit", "exit", "q":
				log.Println("👋 Shutting down...")
				os.Exit(0)
				
			case "status":
				if bot.IsRegistered() {
					log.Printf("✅ Registered at %s", bot.GetPublicEndpoint())
				} else {
					log.Println("❌ Not registered with discovery")
				}
				
			case "":
				// Empty line - could hangup active call
				if len(activeCalls) > 0 {
					log.Println("📴 Hanging up active calls...")
					activeCalls = make(map[string]*botcall.Call)
				}
				
			default:
				// Send as response
				if len(activeCalls) > 0 {
					log.Printf("🗣️  Response: %s", line)
					// TODO: Actually send audio via WebRTC
				} else {
					log.Println("⚠️  No active calls to respond to")
				}
			}
		}
	}()

	// Start HTTP server
	log.Println()
	log.Println("═══════════════════════════════════════════")
	if err := bot.HandleIncoming(listenAddr, nil); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
