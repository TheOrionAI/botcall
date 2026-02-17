// BotCall PWA Configuration
// Edit this to point to your discovery server
const BOTCALL_CONFIG = {
  // Default discovery server
  // For local testing: 'http://localhost:8080'
  // For production: 'https://your-discovery-server.com'
  discoveryUrl: localStorage.getItem('discoveryUrl') || 'http://localhost:8080',
  
  // Version
  version: '0.1.0'
};

// Allow override via URL params (e.g., ?discovery=http://example.com:8080)
const urlParams = new URLSearchParams(window.location.search);
if (urlParams.has('discovery')) {
  BOTCALL_CONFIG.discoveryUrl = urlParams.get('discovery');
  localStorage.setItem('discoveryUrl', BOTCALL_CONFIG.discoveryUrl);
}
