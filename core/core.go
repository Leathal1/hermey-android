// Package core is the Go shared library for Hermdroid.
// It is compiled to an Android AAR via gomobile bind and consumed by the Kotlin/Compose UI shell.
//
// Architecture:
//   - APIClient: HTTP client with cookie jar, CSRF/Origin rules, all 45 REST endpoints
//   - SSEClient: Server-Sent Events parser (7 event types, heartbeat, reconnect)
//   - StreamStateMachine: start, cancel, steer, branch, truncate
//   - Cache: bbolt-backed offline read-only cache
//   - Models: lenient JSON decoding with fallback struct tags
//
// The JNI boundary uses callback interfaces + flat DTOs only.
// gomobile cannot expose goroutines, channels, or generics.
package core

// Version is the Hermdroid core library version.
const Version = "0.1.0"

// HermeyClient is the main entry point for the Go core library.
// It wraps APIClient, SSEClient, StreamStateMachine, and Cache.
// Exposed to Kotlin via gomobile bind.
type HermeyClient struct {
	baseURL string
}

// NewHermeyClient creates a new HermeyClient connected to the given hermes-webui server URL.
func NewHermeyClient(baseURL string) *HermeyClient {
	return &HermeyClient{baseURL: baseURL}
}

// BaseURL returns the configured server URL.
func (c *HermeyClient) BaseURL() string {
	return c.baseURL
}

// Version returns the core library version.
func (c *HermeyClient) Version() string {
	return Version
}
