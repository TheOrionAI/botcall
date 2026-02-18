# BotCall 🎙️

> **The voice layer for the agentic web**

BotCall is an ultra-lightweight calling platform designed specifically for AI-human communication. Bots become servers. Humans join via PWA. No browser hacks, no pretending.

[![AGPL-3.0](https://img.shields.io/badge/License-AGPL%203.0-blue.svg)](LICENSE)

## Why BotCall?

| Platform | Bot Support | Identity | Latency | Setup |
|----------|------------|----------|---------|-------|
| Discord | 🟡 (libs) | ❌ | Low | Complex |
| Jitsi | ❌ | ❌ | Low | Complex |
| Twilio | 🟡 | 🟡 | Medium | Phone-centric |
| **BotCall** | ✅ **Native** | ✅ **BotAuth** | **Low** | **Zero friction** |

## Architecture

```
┌─────────────┐      HTTP/WSS      ┌──────────┐     P2P     ┌─────────┐
│   Orion     │◄───────────────────►│ Discovery│◄──────────►│  Gopi   │
│   (Bot)     │                      │  Server  │   WebRTC    │ (Human) │
│ :9000       │                      │          │             │  PWA    │
└─────────────┘                      └──────────┘             └─────────┘
  Direct mode                          $5/mo
  No middleman                         JSON only
```

## Localtunnel Quick Test

Test BotCall end-to-end without deploying anything:

```bash
# 1. Start discovery server
cd server
./botcall-server

# 2. Expose your bot via localtunnel (in another terminal)
npx localtunnel --port 9000
# Copy the URL: https://xxx.loca.lt

# 3. Run a bot with the localtunnel URL
cd cmd/bot-cli
./bot-cli --agent-id=orion \
  --endpoint=https://xxx.loca.lt \
  --discovery=http://localhost:8080

# 4. Open PWA and call
# https://theorionai.github.io/botcall/pwa/?discovery=http://localhost:8080
# Enter bot ID: orion, click Connect
```

## Quick Start

### 1. Discovery Server (Go)

```bash
# Clone
git clone https://github.com/botcall/botcall.git
cd botcall

# Build
make build

# Run
./bin/botcall-server

# Test
curl http://localhost:8080/health
```

### 2. Bot SDK (Go)

```go
import "github.com/TheOrionAI/botcall-sdk-go"

client := botcall.NewClient("orion", "your-botauth-token")
client.Serve(":9000")  // Accept direct calls
```

### 3. Human Client (PWA)

Open `pwa/` in browser or visit hosted instance:
```
https://theorionai.github.io/botcall/join/orion
```

## API

### Register Bot
```bash
curl -X POST http://localhost:8080/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "orion",
    "endpoint": "203.0.113.45:9000",
    "mode": "direct",
    "attestation": "eyJ..."
  }'
```

### Lookup Bot
```bash
curl http://localhost:8080/v1/lookup/orion
```

## Repositories

This is a monorepo containing:

- `server/` - Go discovery server
- `pwa/` - Human client (PWA)
- `sdk-go/` - Go bot SDK
- `sdk-python/` - Python bot SDK
- `docs/` - Documentation

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md)

## License

AGPL-3.0 for server/core, MIT for SDKs. See [LICENSE](LICENSE)

## Roadmap

- [x] Discovery server prototype
- [ ] WebSocket signaling
- [ ] PWA human client
- [ ] Bot SDKs (Go, Python)
- [ ] BotAuth integration
- [ ] TURN relay (fallback)
- [ ] v1.0 release

---

**🌌 The voice layer for the agentic web**
