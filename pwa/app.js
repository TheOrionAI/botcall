/**
 * BotCall PWA - Human Client
 * Handles WebRTC calls with voice/text mode support
 */

class BotCallPWA {
    constructor() {
        this.discoveryUrl = localStorage.getItem('discoveryUrl') || 'http://localhost:8080';
        this.botId = '';
        this.callActive = false;
        this.currentMode = 'voice'; // 'voice' or 'text'
        this.websocket = null;
        this.peerConnection = null;
        this.localStream = null;
        this.callStartTime = null;
        this.muted = false;
        
        // Audio context for visualization
        this.audioContext = null;
        this.analyser = null;
        
        // TTS/STT
        this.speechSynthesis = window.speechSynthesis;
        this.speechRecognition = null;
        this.ttsQueue = [];
        
        // DOM elements
        this.elements = {};
        
        this.init();
    }

    init() {
        this.cacheElements();
        this.bindEvents();
        this.initSpeechSynthesis();
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
            settingsBtn: document.getElementById('settingsBtn'),
            settingsDrawer: document.getElementById('settingsDrawer'),
            settingsOverlay: document.getElementById('settingsOverlay'),
            closeSettings: document.getElementById('closeSettings'),
            ttsVoice: document.getElementById('ttsVoice'),
            ttsRate: document.getElementById('ttsRate'),
            sttLang: document.getElementById('sttLang'),
            autoSendStt: document.getElementById('autoSendStt'),
            chatArea: document.getElementById('textMode'),
            voiceStatus: document.getElementById('voiceStatus'),
            voiceWaves: document.getElementById('voiceWaves'),
            modeOptions: document.querySelectorAll('.mode-option')
        };
        
        // Restore saved URL
        this.elements.discoveryUrl.value = this.discoveryUrl;
    }

    bindEvents() {
        // Main controls
        this.elements.connectBtn.addEventListener('click', () => this.connect());
        this.elements.hangupBtn.addEventListener('click', () => this.hangup());
        this.elements.muteBtn.addEventListener('click', () => this.toggleMute());
        
        // Text mode
        this.elements.sendBtn.addEventListener('click', () => this.sendTextMessage());
        this.elements.messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.sendTextMessage();
        });
        
        // Mode toggle
        this.elements.modeOptions.forEach(opt => {
            opt.addEventListener('click', () => this.switchMode(opt.dataset.mode));
        });
        
        // Settings
        this.elements.settingsBtn.addEventListener('click', () => this.openSettings());
        this.elements.closeSettings.addEventListener('click', () => this.closeSettings());
        this.elements.settingsOverlay.addEventListener('click', () => this.closeSettings());
        
        // Settings changes
        this.elements.discoveryUrl.addEventListener('change', (e) => {
            this.discoveryUrl = e.target.value;
            localStorage.setItem('discoveryUrl', this.discoveryUrl);
        });
        
        this.elements.ttsRate.addEventListener('input', (e) => {
            localStorage.setItem('ttsRate', e.target.value);
        });
        
        this.elements.sttLang.addEventListener('change', (e) => {
            localStorage.setItem('sttLang', e.target.value);
            this.initSpeechRecognition();
        });
    }

    checkWebRTCSupport() {
        if (!window.RTCPeerConnection) {
            this.showNotification('WebRTC not supported. Voice mode unavailable.', 'warning');
        }
    }

    initSpeechSynthesis() {
        if (!this.speechSynthesis) {
            console.warn('Speech synthesis not available');
            return;
        }
        
        // Populate voice options
        const populateVoices = () => {
            const voices = this.speechSynthesis.getVoices();
            this.elements.ttsVoice.innerHTML = '<option value="">System Default</option>' +
                voices.map((v, i) => `<option value="${i}">${v.name} (${v.lang})</option>`).join('');
        };
        
        populateVoices();
        if (this.speechSynthesis.onvoiceschanged !== undefined) {
            this.speechSynthesis.onvoiceschanged = populateVoices;
        }
        
        // Restore saved rate
        const savedRate = localStorage.getItem('ttsRate');
        if (savedRate) this.elements.ttsRate.value = savedRate;
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
        this.speechRecognition.lang = this.elements.sttLang.value;
        
        let finalTranscript = '';
        
        this.speechRecognition.onresult = (event) => {
            let interimTranscript = '';
            for (let i = event.resultIndex; i < event.results.length; i++) {
                const transcript = event.results[i][0].transcript;
                if (event.results[i].isFinal) {
                    finalTranscript += transcript;
                } else {
                    interimTranscript += transcript;
                }
            }
            
            if (finalTranscript) {
                this.elements.messageInput.value = finalTranscript;
                if (this.elements.autoSendStt.checked) {
                    this.sendTextMessage();
                }
                finalTranscript = '';
            } else {
                this.elements.messageInput.value = interimTranscript;
            }
        };
        
        this.speechRecognition.onerror = (e) => {
            console.log('STT error:', e.error);
        };
    }

    async connect() {
        this.botId = this.elements.botId.value.trim();
        if (!this.botId) {
            this.showNotification('Please enter a Bot ID', 'error');
            return;
        }
        
        this.discoveryUrl = this.elements.discoveryUrl.value.trim();
        localStorage.setItem('discoveryUrl', this.discoveryUrl);
        
        this.setConnectionStatus('connecting', 'Connecting...');
        this.elements.connectBtn.disabled = true;
        
        try {
            // Look up bot in discovery
            const response = await fetch(`${this.discoveryUrl}/v1/lookup/${this.botId}`);
            if (!response.ok) {
                throw new Error(`Bot not found: ${response.status}`);
            }
            
            const