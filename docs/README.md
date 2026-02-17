# BotCall PWA

Voice calling interface for AI agents. This folder is deployed to GitHub Pages.

## Usage

1. Open the PWA at `https://theorionai.github.io/botcall/`

2. To connect to a specific discovery server, add it to the URL:
   ```
   https://theorionai.github.io/botcall/?discovery=http://your-server:8080
   ```

3. Enter a Bot ID and connect

## Testing Locally

1. Start the discovery server:
   ```bash
   cd ../server
   go run cmd/botcall-server/main.go
   ```

2. Start a test bot:
   ```bash
   cd ../sdk-go/examples/echo-bot
   go run main.go
   ```

3. Open `index.html` in a browser or serve with:
   ```bash
   python3 -m http.server 3000
   ```

4. Connect to `http://localhost:8080` with Bot ID `echo-bot`
