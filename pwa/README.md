# BotCall PWA

Progressive Web App for human participants in BotCall.

## Features

- ✅ **Voice Mode** - Full WebRTC calling with AI agents
- ✅ **Text Mode** - Lightweight chat with TTS/STT
- ✅ **Auto Mode** - Seamless fallback between voice and text
- ✅ **Works Offline** - Service worker for reliable experience
- ✅ **Responsive** - Mobile and desktop optimized

## Quick Start

```bash
# Local development
python -m http.server 3000
# Open http://localhost:3000

# Or use any static server
npx serve .
```

## Modes

| Mode | Use Case | Requirements |
|------|----------|--------------|
| **Voice** | Natural conversation | WebRTC support, microphone |
| **Text** | Low bandwidth, accessibility | Just a browser |
| **Auto** | Best experience | WebRTC + graceful fallback |

## Configuration

Settings are saved to `localStorage`:
- Discovery server URL
- TTS voice and rate
- STT language
- Auto-send preference

## Browser Support

| Feature | Chrome | Firefox | Safari | Edge |
|---------|--------|---------|--------|------|
| Voice (WebRTC) | ✅ | ✅ | ✅ | ✅ |
| Text (TTS) | ✅ | ✅ | ✅ | ✅ |
| STT | ✅ | ✅ | 🟡 | ✅ |
| PWA Install | ✅ | ✅ | 🟡 iOS 16.4+ | ✅ |

## Architecture

```
PWA (Browser)
├── WebRTC ←────→ Bot (direct voice)
├── WebSocket ←──→ Discovery Server
├── SpeechRecognition (STT)
└── SpeechSynthesis (TTS)
```

## Building

```bash
# Production build would minify JS/CSS
# For now: serve as static files
```

## Links

- [BotCall Server](https://github.com/TheOrionAI/botcall-server)
- [BotCall SDK](https://github.com/TheOrionAI/botcall-sdk-go)
