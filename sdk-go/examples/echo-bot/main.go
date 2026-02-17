// Echo Bot - Simple BotCall bot for testing
// This bot registers with discovery and echoes back any messages
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	botcall "github.com/theorionai/botcall/sdk-go"
)

func main() {
	// Get configuration from environment or defaults
	agentID := os.Getenv("BOTCALL_AGENT_ID")
	if agentID == "" {
		agentID = "echo-bot"
	}
	
	discoveryURL := os.Getenv("BOTCALL_DISCOVERY_URL")
	if discoveryURL == "" {
		discoveryURL = "http://localhost:8080"
	}
	
	endpoint := os.Getenv("BOTCALL_ENDPOINT")
	if endpoint == "" {
		endpoint = "0.0.0.0:9001"
	}

	fmt.Printf("🤖 BotCall Echo Bot\n")
	fmt.Printf("   Agent ID: %s\n", agentID)
	fmt.Printf("   Discovery: %s\n", discoveryURL)
	fmt.Printf("   Endpoint: %s\n\n", endpoint)

	// Create client
	client := botcall.NewClient(agentID, "test-attestation")
	client.SetDiscoveryURL(discoveryURL)
	
	// Set up call handler
	client.OnCall(func(call *botcall.Call) {
		fmt.Printf("\n📞 Incoming call from: %s\n", call.HumanID)
		fmt.Printf("   Call ID: %s\n", call.CallID)
		fmt.Printf("   Started: %s\n\n", call.StartedAt.Format("15:04:05"))
		fmt.Println("   (Text mode active - type responses below)")
	})

	// Start HTTP server for incoming calls
	fmt.Println("🔌 Connecting to discovery server...")
	
	go func() {
		if err := client.HandleIncoming(endpoint, nil); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait a moment for registration
	fmt.Println("✅ Registered and ready!")
	fmt.Printf("\n🌐 PWA URL: https://theorionai.github.io/botcall/\n")
	fmt.Printf("   To call this bot, enter:\n")
	fmt.Printf("   - Discovery: %s\n", discoveryURL)
	fmt.Printf("   - Bot ID: %s\n\n", agentID)
	fmt.Println("Commands: /quit, /status, /help")
	fmt.Println("─────────────────────────────────")

	// Interactive console for sending messages
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("→ ")
		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		switch input {
		case "/quit", "/q":
			fmt.Println("👋 Goodbye!")
			client.Close()
			return
		case "/status":
			if client.IsRegistered() {
				fmt.Printf("✅ Registered as %s at %s\n", agentID, client.GetPublicEndpoint())
			} else {
				fmt.Println("❌ Not registered")
			}
		case "/help":
			fmt.Println("Commands:")
			fmt.Println("  /quit, /q    - Exit the bot")
			fmt.Println("  /status      - Show registration status")
			fmt.Println("  /help        - Show this help")
			fmt.Println("  <message>    - Echo back (simulates bot response)")
		default:
			// Echo the message
			fmt.Printf("📤 Echo: %s\n", input)
		}
	}
}
