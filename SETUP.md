# BotCall Setup

## Prerequisites

- Go 1.21+ (https://go.dev/dl/)
- Docker (optional)
- Git

## Install Go (Ubuntu/Debian)

```bash
# Download and install
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify
go version
```

## Build

```bash
# Clone
git clone https://github.com/botcall/botcall.git
cd botcall

# Download deps
cd server
go mod tidy

# Build
cd ..
make build

# Run
./bin/botcall-server
```

## API Test

```bash
# Health check
curl http://localhost:8080/health

# Register a bot
curl -X POST http://localhost:8080/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "orion",
    "endpoint": "203.0.113.45:9000",
    "mode": "direct",
    "attestation": "demo-token"
  }'

# Lookup
ccurl http://localhost:8080/v1/lookup/orion

# List online agents
curl http://localhost:8080/v1/agents
```

## Docker

```bash
# Build image
make docker-build

# Run
make docker-run
```

## Production

```bash
# Set environment
export PORT=8080

# Run with systemd
sudo cp systemd/botcall-server.service /etc/systemd/system/
sudo systemctl enable botcall-server
sudo systemctl start botcall-server
```
