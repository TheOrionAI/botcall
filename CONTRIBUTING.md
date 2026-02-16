# Contributing to BotCall

Thank you for your interest! This project is in early alpha and we're figuring things out together.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/TheOrionAI/botcall.git`
3. Install Go 1.21+
4. Run: `make dev`

## Development Workflow

```bash
# Run tests
make test

# Check formatting
make check

# Build locally
make build

# Run dev server
make run
```

## Project Structure

```
botcall/
├── server/       # Go discovery server
├── pwa/          # Human client (SvelteKit)
├── sdk-go/       # Go bot SDK
├── sdk-python/   # Python bot SDK
└── docs/         # Documentation
```

## Code Style

- Go: Standard formatting (`go fmt`)
- JavaScript: Prettier
- Commit messages: Conventional commits

## Areas We Need Help

- [ ] WebSocket implementation
- [ ] PWA WebRTC integration
- [ ] Opus codec streaming
- [ ] Docker deployment
- [ ] Documentation
- [ ] Testing

## Questions?

Open an issue or reach out on Discord (link coming soon).

## License

By contributing, you agree that your contributions will be licensed under the project's AGPL-3.0 license.
