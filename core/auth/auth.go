// Package auth handles login, logout, cookie persistence, and CSRF/Origin rules
// for the hermes-webui server.
//
// The server uses HMAC-signed HTTP-only cookies set via HERMES_WEBUI_PASSWORD.
// CSRF note: the server validates Origin/Referer on POSTs. From a non-browser client,
// omit both headers and the server treats it as a curl-equivalent call.
package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// Client handles authentication with a hermes-webui server.
type Client struct {
	baseURL    string
	httpClient *http.Client
	jar        *cookiejar.Jar
}

// NewClient creates a new auth client for the given server URL.
func NewClient(baseURL string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("auth: create cookie jar: %w", err)
	}
	return &Client{
		baseURL: baseURL,
		jar:     jar,
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
	}, nil
}

// AuthStatus represents the server's authentication status.
type AuthStatus struct {
	AuthEnabled bool   `json:"auth_enabled"`
	LoggedIn    bool   `json:"logged_in,omitempty"`
	Message     string `json:"message,omitempty"`
}

// Status checks whether the server requires authentication.
func (c *Client) Status() (*AuthStatus, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/auth/status")
	if err != nil {
		return nil, fmt.Errorf("auth: status check: %w", err)
	}
	defer resp.Body.Close()

	var status AuthStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("auth: decode status: %w", err)
	}
	return &status, nil
}

// Login authenticates with the server using the given password.
// On success, the session cookie is stored in the client's cookie jar.
func (c *Client) Login(password string) error {
	body := map[string]string{"password": password}
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("auth: marshal login body: %w", err)
	}

	// CSRF: do NOT set Origin or Referer headers.
	// The server treats non-browser (curl-equivalent) calls as safe.
	req, err := http.NewRequest("POST", c.baseURL+"/api/auth/login", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("auth: create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth: login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("auth: login failed (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Verify we got a cookie
	u, _ := url.Parse(c.baseURL)
	cookies := c.jar.Cookies(u)
	if len(cookies) == 0 {
		return fmt.Errorf("auth: login succeeded but no cookie received")
	}

	return nil
}

// Logout clears the session cookie.
func (c *Client) Logout() error {
	req, err := http.NewRequest("POST", c.baseURL+"/api/auth/logout", nil)
	if err != nil {
		return fmt.Errorf("auth: create logout request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth: logout request: %w", err)
	}
	defer resp.Body.Close()

	// Clear all cookies by creating a new jar
	newJar, _ := cookiejar.New(nil)
	c.jar = newJar
	c.httpClient.Jar = newJar

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth: logout failed (HTTP %d)", resp.StatusCode)
	}
	return nil
}

// HTTPClient returns the underlying http.Client with the auth cookie jar.
// This is used by the APIClient to make authenticated requests.
func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

// IsAuthenticated checks if we have a valid session cookie.
func (c *Client) IsAuthenticated() bool {
	u, _ := url.Parse(c.baseURL)
	return len(c.jar.Cookies(u)) > 0
}
