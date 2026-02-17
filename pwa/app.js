/**
 * BotCall PWA - Human Client
 * Handles WebRTC/WS calls with voice/text mode support
 */
class BotCallPWA {
  constructor() {
    const urlParams = new URLSearchParams(window.location.search);
    const urlDiscovery = urlParams.get('discovery');
    
    this.discoveryUrl = urlDiscovery || localStorage.getItem('discoveryUrl') || 'http://localhost:8080';
    this.botId = '';
    this.callActive = false;
    this.currentMode = 'voice';
    this.websocket = null;
    this.peerConnection = null;
    this.localStream = null;
    this.callStartTime = null;
    this.muted = false;
    this.speechSynthesis = window.speechSynthesis;
    this.speechRecognition = null;
    this.elements = {};
    this.init();
  }

  init() {
    this.cacheElements();
    this.bindEvents();
    this.initSpeechRecognition();
    
    localStorage.setItem('discoveryUrl', this.discoveryUrl);
  }

  cacheElements() {
    this.elements = {
      connectCard: document.getElementById('connectCard'),
      callCard: document.getElementById('callCard'),
      discoveryUrl: document.getElementById('discoveryUrl'),
      botId: document.getElementById('botId'),
      connectBtn: document.getElementById('connectBtn'),
      connectionStatus: document.getElementById('connectionStatus'),
      activeBotId: document.getElementById('activeBotId'),
      callDuration: document.getElementById('callDuration'),
      voiceMode: document.getElementById('voiceMode'),
      textMode: document.getElementById('textMode'),
      textInputArea: document.getElementById('textInputArea'),
      messageInput: document.getElementById('messageInput'),
      sendBtn: document.getElementById('sendBtn'),
      muteBtn: document.getElementById('muteBtn'),
      hangupBtn: document.getElementById('hangupBtn'),
      voiceWaves: document.getElementById('voiceWaves'),
      voiceStatus: document.getElementById('voiceStatus'),
      chatArea: document.getElementById('textMode'),
      modeOptions: document.querySelectorAll('.mode-option')
    };

    if (this.elements.discoveryUrl) {
      this.elements.discoveryUrl.value = this.discoveryUrl;
    }
  }

  bindEvents() {
    this.elements.connectBtn?.addEventListener('click', () => this.connect());
    this.elements.hangupBtn?.addEventListener('click', () => this.hangup());
    this.elements.muteBtn?.addEventListener('click', () => this.toggleMute());
    this.elements.sendBtn?.addEventListener('click', () => this.sendTextMessage());
    this.elements.messageInput?.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') this.sendTextMessage();
    });
    this.elements.modeOptions?.forEach(opt => {
      opt.addEventListener('click', () => this.switchMode(opt.dataset.mode));
    });
  }

  initSpeechRecognition() {
    const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
    if (!SpeechRecognition) return;

    this.speechRecognition = new SpeechRecognition();
    this.speechRecognition.continuous = true;
    this.speechRecognition.interimResults = true;
    this.speechRecognition.lang = 'en-US';

    this.speechRecognition.onresult = (event) => {
      for (let i = event.resultIndex; i < event.results.length; i++) {
        if (event.results[i].isFinal) {
          this.sendTextMessage(event.results[i][0].transcript);
        }
      }
    };
  }

  async connect() {
    this.botId = this.elements.botId?.value.trim();
    if (!this.botId) {
      this.setConnectionStatus('offline', 'Enter Bot ID');
      return;
    }

    if (this.elements.discoveryUrl?.value) {
      this.discoveryUrl = this.elements.discoveryUrl.value.trim();
      localStorage.setItem('discoveryUrl', this.discoveryUrl);
    }

    this.setConnectionStatus('connecting', 'Connecting...');
    if (this.elements.connectBtn) this.elements.connectBtn.disabled = true;

    try {
      const response = await fetch(`${this.discoveryUrl}/v1/lookup/${this.botId}`);
      if (!response.ok) throw new Error(`Bot not found: ${response.status}`);

      const botInfo = await response.json();

      if (botInfo.status === 'online') {
        this.elements.connectCard?.classList.add('hidden');
        this.elements.callCard?.classList.remove('hidden');
        if (this.elements.activeBotId) this.elements.activeBotId.textContent = this.botId;
        this.setConnectionStatus('online', 'Connected');

        await this.startCall(botInfo);
        this.startCallTimer();
      } else {
        throw new Error('Bot is offline');
      }
    } catch (error) {
      this.setConnectionStatus('offline', error.message);
      if (this.elements.connectBtn) this.elements.connectBtn.disabled = false;
    }
  }

  async startCall(botInfo) {
    this.callActive = true;
    
    // Try to POST to the bot's /call endpoint
    try {
      const humanId = 'human-' + Math.random().toString(36).substr(2, 8);
      const endpoint = botInfo.endpoint || `${this.discoveryUrl.replace(/\/+$/, '')}/call`;
      
      const callResp = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ human_id: humanId, attestation: '' })
      });

      if (callResp.ok) {
        const callResult = await callResp.json();
        if (callResult.message) {
          this.addMessage('bot', callResult.message);
          this.speak(callResult.message);
        }
        this.connectWebSocket();
      } else {
        this.switchMode('text');
      }
    } catch (e) {
      console.log('Direct call failed, using text mode:', e);
      this.switchMode('text');
    }
  }

  connectWebSocket() {
    try {
      const wsUrl = this.discoveryUrl.replace('http', 'ws') + '?agent=' + this.botId;
      this.websocket = new WebSocket(wsUrl);
      
      this.websocket.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.type === 'text') {
          this.addMessage('bot', data.text);
          this.speak(data.text);
        }
      };

      this.websocket.onerror = (e) => console.log('WS error:', e);
    } catch (e) {
      console.log('WebSocket not available:', e);
    }
  }

  switchMode(mode) {
    this.currentMode = mode;
    
    document.querySelectorAll('.mode-option').forEach(el => el.classList.remove('active'));
    document.querySelector(`.mode-option[data-mode="${mode}"]`)?.classList.add('active');

    if (mode === 'voice') {
      this.elements.voiceMode?.classList.remove('hidden');
      this.elements.textMode?.classList.add('hidden');
      this.elements.textInputArea?.classList.add('hidden');
      this.speechRecognition?.start();
    } else {
      this.elements.voiceMode?.classList.add('hidden');
      this.elements.textMode?.classList.remove('hidden');
      this.elements.textInputArea?.classList.remove('hidden');
      this.speechRecognition?.stop();
    }
  }

  sendTextMessage(text) {
    const message = text || this.elements.messageInput?.value?.trim();
    if (!message) return;

    this.addMessage('human', message);
    
    if (this.websocket?.readyState === WebSocket.OPEN) {
      this.websocket.send(JSON.stringify({ type: 'text', text: message }));
    }

    if (this.elements.messageInput) {
      this.elements.messageInput.value = '';
    }
  }

  addMessage(sender, text) {
    const div = document.createElement('div');
    div.className = `message ${sender}`;
    div.textContent = text;
    this.elements.chatArea?.appendChild(div);
    this.elements.chatArea?.scrollTo(0, this.elements.chatArea.scrollHeight);
  }

  speak(text) {
    if (!this.speechSynthesis) return;
    const utterance = new SpeechSynthesisUtterance(text);
    this.speechSynthesis.speak(utterance);
  }

  setConnectionStatus(state, text) {
    const el = this.elements.connectionStatus;
    if (!el) return;
    
    el.className = 'status ' + state;
    el.textContent = text;
  }

  startCallTimer() {
    this.callStartTime = Date.now();
    setInterval(() => {
      if (!this.callActive || !this.elements.callDuration) return;
      const secs = Math.floor((Date.now() - this.callStartTime) / 1000);
      const m = Math.floor(secs / 60).toString().padStart(2, '0');
      const s = (secs % 60).toString().padStart(2, '0');
      this.elements.callDuration.textContent = `${m}:${s}`;
    }, 1000);
  }

  toggleMute() {
    this.muted = !this.muted;
    if (this.localStream) {
      this.localStream.getAudioTracks().forEach(t => t.enabled = !this.muted);
    }
    if (this.elements.muteBtn) {
      this.elements.muteBtn.textContent = this.muted ? '🔇 Unmute' : '🔇 Mute';
    }
  }

  hangup() {
    this.callActive = false;
    this.websocket?.close();
    this.peerConnection?.close();
    this.localStream?.getTracks().forEach(t => t.stop());
    this.speechRecognition?.stop();
    
    this.elements.callCard?.classList.add('hidden');
    this.elements.connectCard?.classList.remove('hidden');
    this.elements.connectBtn?.removeAttribute('disabled');
    this.setConnectionStatus('offline', 'Offline');
  }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
  window.botcall = new BotCallPWA();
});