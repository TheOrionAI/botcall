/**
 * BotCall PWA - Human Client
 * Handles WebRTC calls with voice/text mode support
 */

class BotCallPWA {
    constructor() {
        this.discoveryUrl = (typeof BOTCALL_CONFIG !== 'undefined' ? BOTCALL_CONFIG.discoveryUrl : null) || localStorage.getItem('discoveryUrl') || 'http://localhost:8080';
        this.botId = '';
        this.callActive = false;
        this.currentMode = 'voice';
        this.websocket = null;
        this.peerConnection = null;
        this.localStream = null;
        this.callStartTime = null;
        this.muted = false;
        this.audioContext = null;
        this.analyser = null;
        this.speechSynthesis = window.speechSynthesis;
        this.speechRecognition = null;
        this.elements = {};
        this.messages = [];
        this.init();
    }

    init() {
        this.cacheElements();
        this.bindEvents();
        this.initSpeechSynthesis();
        this.initSpeechRecognition();
        this.checkWebRTCSupport();
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

    checkWebRTCSupport() {
        if (!window.RTCPeerConnection) {
            console.warn('WebRTC not available');
        }
    }

    initSpeechSynthesis() {
        if (!this.speechSynthesis) {
            console.warn('Speech synthesis not available');
        }
    }

    initSpeechRecognition() {
        const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
        if (!SpeechRecognition) {
            console.warn('Speech recognition not available');
            return;
        }

        this.speechRecognition = new SpeechRecognition();
        this.speechRecognition.continuous = true;
        this.speechRecognition.interimResults = true;
        this.speechRecognition.lang = 'en-US';

        let finalTranscript = '';

        this.speechRecognition.onresult = (event) => {
            let interimTranscript = '';
            for (let i = event.resultIndex; i < event.results.length; i++) {
                const transcript = event.results[i][0].transcript;
                if (event.results[i].isFinal) {
                    finalTranscript = transcript;
                    this.addMessage('human', finalTranscript);
                    this.sendTextMessage(finalTranscript);
                    finalTranscript = '';
                } else {
                    interimTranscript += transcript;
                }
            }
        };

        this.speechRecognition.onerror = (e) => {
            console.log('STT error:', e.error);
        };
    }

    async connect() {
        this.botId = this.elements.botId?.value.trim();
        if (!this.botId) {
            this.showNotification('Please enter a Bot ID', 'error');
            return;
        }

        this.discoveryUrl = this.elements.discoveryUrl?.value.trim() || this.discoveryUrl;
        localStorage.setItem('discoveryUrl', this.discoveryUrl);

        this.setConnectionStatus('connecting', 'Connecting...');
        if (this.elements.connectBtn) this.elements.connectBtn.disabled = true;

        try {
            const response = await fetch(`${this.discoveryUrl}/v1/lookup/${this.botId}`);
            if (!response.ok) {
                throw new Error(`Bot not found: ${response.status}`);
            }

            const botInfo = await response.json();
            
            if (botInfo.status === 'online') {
                this.elements.connectCard?.classList.add('hidden');
                this.elements.callCard?.classList.remove('hidden');
                if (this.elements.activeBotId) this.elements.activeBotId.textContent = this.botId;
                this.setConnectionStatus('online', 'Connected');
                
                if (this.currentMode === 'voice' && botInfo.mode === 'direct') {
                    await this.startVoiceCall(botInfo.endpoint);
                } else {
                    this.switchMode('text');
                }
                this.startCallTimer();
            } else {
                throw new Error('Bot is offline');
            }
        } catch (error) {
            console.error('Connection error:', error);
            this.showNotification('Connection failed: ' + error.message, 'error');
            this.setConnectionStatus('offline', 'Offline');
        } finally {
            if (this.elements.connectBtn) this.elements.connectBtn.disabled = false;
        }
    }

    async startVoiceCall(endpoint) {
        try {
            this.localStream = await navigator.mediaDevices.getUserMedia({ audio: true });
            
            this.peerConnection = new RTCPeerConnection({
                iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
            });

            this.localStream.getTracks().forEach(track => {
                this.peerConnection.addTrack(track, this.localStream);
            });

            this.peerConnection.ontrack = (event) => {
                const remoteAudio = new Audio();
                remoteAudio.srcObject = event.streams[0];
                remoteAudio.play();
            };

            const offer = await this.peerConnection.createOffer();
            await this.peerConnection.setLocalDescription(offer);
            
            this.callActive = true;
            this.setVoiceStatus('In voice call');
            
            this.connectWebSocket();

        } catch (error) {
            console.error('Voice call error:', error);
            this.switchMode('text');
        }
    }

    connectWebSocket() {
        const wsUrl = this.discoveryUrl.replace('http', 'ws');
        this.websocket = new WebSocket(`${wsUrl}?agent=${this.botId}`);
        
        this.websocket.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (data.type === 'text') {
                this.addMessage('bot', data.text);
                if (this.currentMode === 'voice