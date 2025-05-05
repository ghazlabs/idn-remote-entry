package crawler

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const (
	navigateTimeout   = 10 * time.Second
	navigationTimeout = 10 * time.Second
	loadTimeout       = 10 * time.Second
	idleTimeout       = 10 * time.Second
	stableTimeout     = 10 * time.Second
	htmlTimeout       = 5 * time.Second
)

// CloudflareBypassService handles Cloudflare protection bypassing
type CloudflareBypassService struct {
	browser      *rod.Browser
	cookiesCache map[string][]*proto.NetworkCookie
	mutex        sync.RWMutex
	userAgent    string
}

// NewCloudflareBypassService creates a new service for bypassing Cloudflare
// Set to true for headless mode, false to see the browser visually
// Using false (non-headless) is recommended for debugging Cloudflare challenges
func NewCloudflareBypassService(headless bool) (*CloudflareBypassService, error) {
	// Custom user agent that's less likely to be flagged
	userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"

	// Launch browser with stealth settings
	path, _ := launcher.LookPath()
	launchOpts := launcher.New().
		Bin(path).
		Headless(headless).
		Set("disable-default-apps").
		Set("disable-blink-features", "AutomationControlled").
		Set("disable-web-security"). // Helps with some protections
		Set("window-size", "1920,1080").
		Set("user-agent", userAgent)

	// For deployments in containers, you may need these
	if !headless {
		launchOpts.Set("no-sandbox").
			Set("disable-gpu").
			Set("disable-dev-shm-usage")
	}

	controlURL := launchOpts.MustLaunch()

	// Create browser with stealth settings
	browser := rod.New().
		ControlURL(controlURL).
		MustConnect()

	return &CloudflareBypassService{
		browser:      browser,
		cookiesCache: make(map[string][]*proto.NetworkCookie),
		userAgent:    userAgent,
	}, nil
}

// GetHTML fetches the HTML of a page while bypassing Cloudflare protection
func (s *CloudflareBypassService) GetHTML(url string) (string, error) {
	log.Println("Navigating to URL:", url)
	domain := extractDomain(url)

	// Check if we have existing cookies for this domain
	s.mutex.RLock()
	cookies, hasCookies := s.cookiesCache[domain]
	s.mutex.RUnlock()

	// Create a new page
	page := s.browser.MustPage()
	defer page.MustClose()

	// Block loading certain non-essential resources, but be careful not to block
	// resources that might be needed for Cloudflare challenges
	// https://go-rod.github.io/#/network/README?id=blocking-certain-resources-from-loading
	router := page.HijackRequests()

	// Only block these specific resources
	router.MustAdd("*", func(ctx *rod.Hijack) {
		reqType := ctx.Request.Type()
		reqURL := ctx.Request.URL().String()

		// Block images, ads, trackers, and analytics
		if reqType == proto.NetworkResourceTypeImage ||
			reqType == proto.NetworkResourceTypeWebSocket || // we don't need websockets to fetch html
			reqType == proto.NetworkResourceTypeStylesheet ||
			reqType == proto.NetworkResourceTypeMedia ||
			reqType == proto.NetworkResourceTypeFont {

			// Don't block any Cloudflare-related resources
			if strings.Contains(reqURL, "cloudflare") ||
				strings.Contains(reqURL, "cf-") ||
				strings.Contains(reqURL, "turnstile") {
				ctx.ContinueRequest(&proto.FetchContinueRequest{})
				return
			}

			// Block the resource
			ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
			return
		}

		// Let all other requests through, especially scripts and stylesheets needed for CF
		ctx.ContinueRequest(&proto.FetchContinueRequest{})
	})

	go router.Run()
	defer router.Stop()

	// Apply user agent and viewport
	page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: s.userAgent,
	})
	page.MustSetViewport(1920, 1080, 1.0, false)

	// Set real browser headers
	page.MustSetExtraHeaders(
		"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
		"accept-language", "en-US,en;q=0.9",
		"cache-control", "no-cache",
		"sec-ch-ua", `"Chromium";v="122", "Not(A:Brand";v="24", "Google Chrome";v="122"`,
		"sec-ch-ua-mobile", "?0",
		"sec-ch-ua-platform", "\"macOS\"",
		"sec-fetch-dest", "document",
		"sec-fetch-mode", "navigate",
		"sec-fetch-site", "none",
		"sec-fetch-user", "?1",
		"upgrade-insecure-requests", "1",
		"pragma", "no-cache",
	)

	// Apply stealth scripts to avoid detection
	// These scripts help mask the fact that this is an automated browser
	page.MustEvalOnNewDocument(`
		// Override the 'webdriver' property
		Object.defineProperty(navigator, 'webdriver', { 
			get: () => false 
		});
		
		// Add language and platform details
		Object.defineProperty(navigator, 'languages', { 
			get: () => ['en-US', 'en', 'es'] 
		});
		
		// Add chrome object
		window.chrome = {
			app: {
				isInstalled: true,
			},
			webstore: {
				onInstallStageChanged: {},
				onDownloadProgress: {},
			},
			runtime: {
				PlatformOs: {
					MAC: 'mac',
					WIN: 'win',
					ANDROID: 'android',
					CROS: 'cros',
					LINUX: 'linux',
					OPENBSD: 'openbsd',
				},
				PlatformArch: {
					ARM: 'arm',
					X86_32: 'x86-32',
					X86_64: 'x86-64',
				},
				PlatformNaclArch: {
					ARM: 'arm',
					X86_32: 'x86-32',
					X86_64: 'x86-64',
				},
				RequestUpdateCheckStatus: {
					THROTTLED: 'throttled',
					NO_UPDATE: 'no_update',
					UPDATE_AVAILABLE: 'update_available',
				},
				OnInstalledReason: {
					INSTALL: 'install',
					UPDATE: 'update',
					CHROME_UPDATE: 'chrome_update',
					SHARED_MODULE_UPDATE: 'shared_module_update',
				},
				OnRestartRequiredReason: {
					APP_UPDATE: 'app_update',
					OS_UPDATE: 'os_update',
					PERIODIC: 'periodic',
				},
			},
		};
	`)

	// If we have cookies for this domain, apply them
	if hasCookies {
		s.browser.MustSetCookies(cookies...)
	}

	// Set timeout for navigation to prevent hanging
	page.Timeout(navigateTimeout)

	// Navigate to URL
	err := page.Navigate(url)
	if err != nil {
		return "", fmt.Errorf("navigation failed: %w", err)
	}

	// Wait for page to load with reasonable timeout
	waitLoad := page.Timeout(loadTimeout).MustWaitNavigation()
	waitLoad()

	// Handle Cloudflare challenge if detected
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		html, err := page.HTML()
		if err != nil {
			log.Printf("Error getting HTML, retry %d: %v", i, err)
			time.Sleep(1 * time.Second)
			continue // Ignore error for now, we will retry
		}

		// Check for Cloudflare challenge indicators
		if strings.Contains(html, "cf-browser-verification") ||
			strings.Contains(html, "cf-challenge") ||
			strings.Contains(html, "Just a moment") ||
			strings.Contains(html, "challenge-running") ||
			strings.Contains(html, "cf-turnstile-response") {

			log.Printf("Detected Cloudflare challenge, attempt %d/%d\n", i+1, maxRetries)

			// Wait for the challenge to be solved naturally or try to solve it
			page.Timeout(stableTimeout).MustWaitStable()

			// Try clicking on the checkbox or other elements if they exist
			_, _ = page.Eval(`
				try {
					// If there's a checkbox to click
					let checkbox = document.querySelector('.cf-checkbox') || document.querySelector('#cf-please-wait');
					if (checkbox) checkbox.click();
					
					// Wait and check if the challenge is solved
					setTimeout(() => {
						let overlay = document.querySelector('#cf-spinner-please-wait');
						if (overlay) overlay.remove();
					}, 1000);
				} catch(e) { console.log('Auto-solve attempt failed:', e); }
			`)

			// Give some time for the challenge to process
			time.Sleep(3 * time.Second)

			// Wait for any potential redirects after challenge completion
			waitComplete := page.Timeout(navigationTimeout).MustWaitNavigation()
			waitComplete()
		} else {
			// No challenge detected or it's been solved
			break
		}
	}

	// Wait for any dynamic content to load with timeouts to prevent hanging
	waitIdle := page.Timeout(idleTimeout).MustWaitRequestIdle()
	waitIdle()

	// Store cookies for future use
	newCookies, err := page.Cookies(nil)
	if err == nil && len(newCookies) > 0 {
		s.mutex.Lock()
		s.cookiesCache[domain] = newCookies
		s.mutex.Unlock()
	}

	// Get final HTML with timeout
	return page.Timeout(htmlTimeout).HTML()
}

// Close cleans up resources
func (s *CloudflareBypassService) Close() {
	s.browser.MustClose()
}

// Helper to extract domain from URL
func extractDomain(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "www.")

	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}

	return url
}
