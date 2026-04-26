package server

import (
	"encoding/json"
	"net/http"
)

// PWA provides Progressive Web App support
type PWA struct {
	manifest *WebAppManifest
}

// WebAppManifest represents a PWA manifest
type WebAppManifest struct {
	Name            string      `json:"name"`
	ShortName       string      `json:"short_name"`
	Description     string      `json:"description"`
	StartURL        string      `json:"start_url"`
	Display         string      `json:"display"`
	BackgroundColor string      `json:"background_color"`
	ThemeColor      string      `json:"theme_color"`
	Icons           []Icon      `json:"icons"`
	Categories      []string    `json:"categories"`
	Screenshots     []Screenshot `json:"screenshots,omitempty"`
}

// Icon represents a PWA icon
type Icon struct {
	Src     string `json:"src"`
	Sizes   string `json:"sizes"`
	Type    string `json:"type"`
	Purpose string `json:"purpose,omitempty"`
}

// Screenshot represents a PWA screenshot
type Screenshot struct {
	Src   string `json:"src"`
	Sizes string `json:"sizes"`
	Type  string `json:"type"`
}

// NewPWA creates a new PWA instance
func NewPWA(siteName, siteDescription string) *PWA {
	return &PWA{
		manifest: &WebAppManifest{
			Name:            siteName,
			ShortName:       "Cassocial",
			Description:     siteDescription,
			StartURL:        "/",
			Display:         "standalone",
			BackgroundColor: "#282a36",
			ThemeColor:      "#bd93f9",
			Icons: []Icon{
				{
					Src:     "/static/icons/icon-72x72.png",
					Sizes:   "72x72",
					Type:    "image/png",
					Purpose: "any",
				},
				{
					Src:     "/static/icons/icon-96x96.png",
					Sizes:   "96x96",
					Type:    "image/png",
					Purpose: "any",
				},
				{
					Src:     "/static/icons/icon-128x128.png",
					Sizes:   "128x128",
					Type:    "image/png",
					Purpose: "any",
				},
				{
					Src:     "/static/icons/icon-144x144.png",
					Sizes:   "144x144",
					Type:    "image/png",
					Purpose: "any",
				},
				{
					Src:     "/static/icons/icon-152x152.png",
					Sizes:   "152x152",
					Type:    "image/png",
					Purpose: "any",
				},
				{
					Src:     "/static/icons/icon-192x192.png",
					Sizes:   "192x192",
					Type:    "image/png",
					Purpose: "any maskable",
				},
				{
					Src:     "/static/icons/icon-384x384.png",
					Sizes:   "384x384",
					Type:    "image/png",
					Purpose: "any",
				},
				{
					Src:     "/static/icons/icon-512x512.png",
					Sizes:   "512x512",
					Type:    "image/png",
					Purpose: "any maskable",
				},
			},
			Categories: []string{"social", "productivity"},
		},
	}
}

// ServeManifest serves the PWA manifest.json
func (p *PWA) ServeManifest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/manifest+json")
	json.NewEncoder(w).Encode(p.manifest)
}

// ServeServiceWorker serves the service worker JavaScript
func (p *PWA) ServeServiceWorker(w http.ResponseWriter, r *http.Request) {
	serviceWorker := `// Service Worker for Cassocial PWA
const CACHE_NAME = 'cassocial-v1';
const urlsToCache = [
  '/',
  '/static/css/main.css',
  '/static/js/main.js',
  '/static/icons/icon-192x192.png',
  '/static/icons/icon-512x512.png'
];

// Install service worker
self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then(cache => cache.addAll(urlsToCache))
  );
});

// Fetch from cache, fallback to network
self.addEventListener('fetch', event => {
  event.respondWith(
    caches.match(event.request)
      .then(response => response || fetch(event.request))
  );
});

// Update service worker
self.addEventListener('activate', event => {
  event.waitUntil(
    caches.keys().then(cacheNames => {
      return Promise.all(
        cacheNames.map(cacheName => {
          if (cacheName !== CACHE_NAME) {
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
});
`

	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Service-Worker-Allowed", "/")
	w.Write([]byte(serviceWorker))
}

// GetInstallPromptHTML returns HTML for PWA install prompt
func (p *PWA) GetInstallPromptHTML() string {
	return `<div id="pwa-install-prompt" style="display:none;" role="dialog" aria-labelledby="install-title">
  <div class="install-prompt-content">
    <h3 id="install-title">Install Cassocial</h3>
    <p>Install Cassocial as an app for quick access and offline support.</p>
    <button id="install-button" aria-label="Install application">Install</button>
    <button id="dismiss-button" aria-label="Dismiss installation prompt">Not now</button>
  </div>
</div>

<script>
let deferredPrompt;

window.addEventListener('beforeinstallprompt', (e) => {
  e.preventDefault();
  deferredPrompt = e;
  document.getElementById('pwa-install-prompt').style.display = 'block';
});

document.getElementById('install-button').addEventListener('click', async () => {
  if (deferredPrompt) {
    deferredPrompt.prompt();
    const { outcome } = await deferredPrompt.userChoice;
    console.log('Install outcome:', outcome);
    deferredPrompt = null;
  }
  document.getElementById('pwa-install-prompt').style.display = 'none';
});

document.getElementById('dismiss-button').addEventListener('click', () => {
  document.getElementById('pwa-install-prompt').style.display = 'none';
});
</script>`
}

// GetOfflinePage returns HTML for offline page
func GetOfflinePage() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Offline - Cassocial</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #282a36;
            color: #f8f8f2;
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            margin: 0;
            padding: 20px;
            text-align: center;
        }
        .container { max-width: 600px; }
        h1 { color: #bd93f9; font-size: 2.5em; margin: 0 0 20px 0; }
        p { font-size: 1.1em; line-height: 1.6; opacity: 0.9; }
        .icon { font-size: 5em; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon" role="img" aria-label="Offline icon">📴</div>
        <h1>You're Offline</h1>
        <p>It looks like you've lost your internet connection. Some features may not be available.</p>
        <p>Your saved profiles and links are still accessible.</p>
        <p style="margin-top: 40px;">
            <button onclick="window.location.reload()"
                    style="padding: 12px 24px; background: #bd93f9; color: #282a36;
                           border: none; border-radius: 8px; font-size: 16px; cursor: pointer;"
                    aria-label="Retry connection">
                Retry
            </button>
        </p>
    </div>
</body>
</html>`
}

// GetPWAMetaTags returns HTML meta tags for PWA
func GetPWAMetaTags(themeColor string) string {
	if themeColor == "" {
		themeColor = "#bd93f9"
	}

	return `<!-- PWA Meta Tags -->
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="apple-mobile-web-app-title" content="Cassocial">
<meta name="theme-color" content="` + themeColor + `">
<link rel="manifest" href="/manifest.json">
<link rel="apple-touch-icon" href="/static/icons/icon-192x192.png">
<link rel="apple-touch-icon" sizes="152x152" href="/static/icons/icon-152x152.png">
<link rel="apple-touch-icon" sizes="180x180" href="/static/icons/icon-180x180.png">
<link rel="apple-touch-icon" sizes="167x167" href="/static/icons/icon-192x192.png">`
}
