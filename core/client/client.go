// Package client provides the APIClient for all hermes-webui REST endpoints.
// It wraps the auth.Client's http.Client and adds generic request helpers.
//
// CSRF note: the server validates Origin/Referer on browser POSTs.  This
// client is a non-browser caller, so it deliberately omits both headers.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Leathal1/hermey-android/core/auth"
)

// APIError is an HTTP error returned by the hermes-webui server.
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Path       string `json:"path,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("api error %d on %s: %s", e.StatusCode, e.Path, e.Message)
}

// APIClient is the typed HTTP client for the hermes-webui API.
// It handles cookie-based auth, CSRF rules, and lenient JSON decoding.
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAPIClient creates a new APIClient using the auth client's HTTP client.
func NewAPIClient(authClient *auth.Client) *APIClient {
	return &APIClient{
		baseURL:    "",
		httpClient: authClient.HTTPClient(),
	}
}

// NewAPIClientWithHTTP creates a client from an arbitrary http.Client.
func NewAPIClientWithHTTP(baseURL string, httpClient *http.Client) *APIClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &APIClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// SetBaseURL sets the server base URL.
func (c *APIClient) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// BaseURL returns the configured server base URL.
func (c *APIClient) BaseURL() string {
	return c.baseURL
}

// HTTPClient returns the underlying http.Client.
func (c *APIClient) HTTPClient() *http.Client {
	return c.httpClient
}

func (c *APIClient) makeURL(path string) string {
	return c.baseURL + path
}

// doRequest performs an HTTP request, sets JSON content type when a body is
// present, and decodes a JSON response when result is non-nil.
func (c *APIClient) doRequest(method, path string, body interface{}, result interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("client: marshal body for %s: %w", path, err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.makeURL(path), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("client: create request %s: %w", path, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client: %s %s: %w", method, path, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
			Path:       path,
		}
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("client: decode %s: %w", path, err)
		}
	}
	return resp, nil
}

// DoGET performs a GET request and decodes the JSON response into result.
func (c *APIClient) DoGET(path string, result interface{}) error {
	_, err := c.doRequest("GET", path, nil, result)
	return err
}

// DoPOST performs a POST request with a JSON body and decodes the response.
func (c *APIClient) DoPOST(path string, body, result interface{}) error {
	_, err := c.doRequest("POST", path, body, result)
	return err
}

// DoDELETE performs a DELETE request.
func (c *APIClient) DoDELETE(path string, result interface{}) error {
	_, err := c.doRequest("DELETE", path, nil, result)
	return err
}

// DoPUT performs a PUT request with a JSON body.
func (c *APIClient) DoPUT(path string, body, result interface{}) error {
	_, err := c.doRequest("PUT", path, body, result)
	return err
}

// DoRawGET performs a GET and returns the raw response body.
func (c *APIClient) DoRawGET(path string) ([]byte, error) {
	resp, err := c.doRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
}

// BuildQuery is a helper to URL-encode query parameters.
func BuildQuery(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return values.Encode()
}
