// Service Worker for BotCall PWA
const VERSION = 'v1.0.0';
const CACHE_NAME = `botcall-${VERSION}`;

const urlsToCache = [
    '/',
    '/index.html',
    '/app.js',
    '/manifest.json'
];

// Install - cache assets
self.addEventListener('install', (event) => {
    event.waitUntil(
        caches.open(CACHE_NAME)
            .then((cache) => {
                console.log('Cache opened');
                return cache.addAll(urlsToCache);
            })
            .then(() => self.skipWaiting())
    );
});

// Activate - clean old caches
self.addEventListener('activate', (event) => {
    event.waitUntil(
        caches.keys().then((cacheNames) => {
            return Promise.all(
                cacheNames
                    .filter(name => name !== CACHE_NAME)
                    .map(name => caches.delete(name))
            );
        })
        .then(() => self.clients.claim())
    );
});

// Fetch - serve from cache or network
self.addEventListener('fetch', (event) => {
    // Skip WebSocket requests
    if (event.request.url.includes('ws://') || event.request.url.includes('wss://')) {
        return;
    }

    event.respondWith(
        caches.match(event.request)
            .then((response) => {
                // Return cached version or fetch from network
                if (response) {
                    return response;
                }
                return fetch(event.request);
            }
        )
    );
});
