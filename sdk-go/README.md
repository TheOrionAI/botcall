# BotCall Go SDK

Official Go SDK for BotCall — the voice layer for the agentic web.

## Installation

```bash
go get github.com/TheOrionAI/botcall-sdk-go
```

## Quick Start

```go
package main

import (
    "log"
    "github.com/TheOrionAI/botcall-sdk-go"
)

func main() {
    // Create bot client
    bot := botcall.NewClient("orion", "your-botauth-token")
    
    // Handle incoming calls
    bot.OnCall(func(call *botcall.Call) {
        log.Printf("📞 Call from %s", call.HumanID)
        // TODO: Stream audio, process STT, generate response
    })
    
    // Connect and start accepting calls
    if err := bot.Connect(); err != nil {
        log.Fatal(err)
    }
    
    bot.HandleIncoming(":9000", nil)
}
```

## Features

- ✅ Registration with discovery server
- ✅ WebSocket signaling
- ✅ HTTP call acceptance
- ✅ Keepalive / heartbeat
- ✅ Graceful shutdown
- 🚧 Opus streaming (coming)
- 🚧 STT integration (coming)

## Examples

See `examples/` directory for:

- `basic/` - Minimal bot
- `interactive/` - Bot with TTS responses

## Configuration

```go
bot := botcall.NewClient("your-agent-id", "your-botauth-token")
bot.SetDiscoveryURL("https://discover.botcall.io")
bot.Endpoint = "0.0.0.0:9000" // Must be public after port forward
```

## Architecture

Bot SDK sits between your AI and human callers:

```
Human (PWA) ◄──WebRTC/Opus──► Bot SDK ◄──API──► Your AI
                     │
                     ▼
              Discovery Server
```

## License

MIT — See LICENSE file
