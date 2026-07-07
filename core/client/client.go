// Package client provides the APIClient for all hermes-webui REST endpoints.
// It wraps the auth.Client's http.Client and adds typed methods for every endpoint.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Leathal1/hermey-android/core/auth"
	"github.com/Leathal1/hermey-android/core/models"
)

// APIClient is the typed HTTP client for the hermes-webui API.
// It handles cookie-based auth, CSRF rules, and lenient JSON decoding.
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAPIClient creates a new APIClient using the auth client's HTTP client.
func NewAPIClient(authClient *auth.Client) *APIClient {
	return &APIClient{
		baseURL:    "", // derived from auth client
		httpClient: authClient.HTTPClient(),
	}
}

// SetBaseURL sets the server base URL.
func (c *APIClient) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// doGet performs a GET request and decodes the JSON response.
func (c *APIClient) doGet(path string, result interface{}) error {
	resp, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("client: GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("client: GET %s: HTTP %d: %s", path, resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("client: decode %s: %w", path, err)
	}
	return nil
}

// doPost performs a POST request with a JSON body and decodes the response.
func (c *APIClient) doPost(path string, body, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("client: marshal body for %s: %w", path, err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest("POST", c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("client: create request %s: %w", path, err)
	}
	req.Header.Set("Content-Type", "application/json")
	// CSRF: do NOT set Origin or Referer

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("client: POST %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("client: POST %s: HTTP %d: %s", path, resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("client: decode %s: %w", path, err)
		}
	}
	return nil
}

// ── Health ──

// Health checks if the server is reachable.
func (c *APIClient) Health() error {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("client: health check: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("client: health check: HTTP %d", resp.StatusCode)
	}
	return nil
}

// ── Sessions ──

// SessionsListResponse is the response from GET /api/sessions.
type SessionsListResponse struct {
	Sessions []models.Session `json:"sessions"`
}

// ListSessions returns all sessions.
func (c *APIClient) ListSessions() (*SessionsListResponse, error) {
	var result SessionsListResponse
	if err := c.doGet("/api/sessions", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SessionDetailResponse is the response from GET /api/session.
type SessionDetailResponse struct {
	Session  models.Session    `json:"session"`
	Messages []models.ChatMessage `json:"messages"`
}

// GetSession loads a session with the last N messages.
func (c *APIClient) GetSession(sessionID string, msgLimit int) (*SessionDetailResponse, error) {
	path := fmt.Sprintf("/api/session?session_id=%s&messages=1&msg_limit=%d",
		url.QueryEscape(sessionID), msgLimit)
	var result SessionDetailResponse
	if err := c.doGet(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// NewSession creates a new session.
func (c *APIClient) NewSession(workspace, model, modelProvider, profile string) (*models.Session, error) {
	body := map[string]string{}
	if workspace != "" { body["workspace"] = workspace }
	if model != "" { body["model"] = model }
	if modelProvider != "" { body["model_provider"] = modelProvider }
	if profile != "" { body["profile"] = profile }

	var result models.Session
	if err := c.doPost("/api/session/new", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RenameSession renames a session.
func (c *APIClient) RenameSession(sessionID, title string) error {
	return c.doPost("/api/session/rename", map[string]string{
		"session_id": sessionID,
		"title":      title,
	}, nil)
}

// DeleteSession deletes a session.
func (c *APIClient) DeleteSession(sessionID string) error {
	return c.doPost("/api/session/delete", map[string]string{
		"session_id": sessionID,
	}, nil)
}

// PinSession pins or unpins a session.
func (c *APIClient) PinSession(sessionID string, pinned bool) error {
	return c.doPost("/api/session/pin", map[string]interface{}{
		"session_id": sessionID,
		"pinned":     pinned,
	}, nil)
}

// ArchiveSession archives or unarchives a session.
func (c *APIClient) ArchiveSession(sessionID string, archived bool) error {
	return c.doPost("/api/session/archive", map[string]interface{}{
		"session_id": sessionID,
		"archived":   archived,
	}, nil)
}

// ── Chat ──

// StartChat starts a new chat message in a session.
func (c *APIClient) StartChat(sessionID, message, workspace, model string) (*models.StreamStartResponse, error) {
	body := map[string]string{
		"session_id": sessionID,
		"message":    message,
	}
	if workspace != "" { body["workspace"] = workspace }
	if model != "" { body["model"] = model }

	var result models.StreamStartResponse
	if err := c.doPost("/api/chat/start", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CancelChat cancels an active chat stream.
func (c *APIClient) CancelChat(streamID string) error {
	path := fmt.Sprintf("/api/chat/cancel?stream_id=%s", url.QueryEscape(streamID))
	resp, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("client: cancel chat: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

// SteerChat sends a /steer message to an active stream.
func (c *APIClient) SteerChat(sessionID, text string) (*models.SteerResponse, error) {
	var result models.SteerResponse
	if err := c.doPost("/api/chat/steer", map[string]string{
		"session_id": sessionID,
		"text":       text,
	}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Models / Providers / Profiles ──

// ListModels returns available models.
func (c *APIClient) ListModels() ([]models.ModelInfo, error) {
	var result []models.ModelInfo
	if err := c.doGet("/api/models", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListProfiles returns available agent profiles.
func (c *APIClient) ListProfiles() ([]models.ProfileInfo, error) {
	var result []models.ProfileInfo
	if err := c.doGet("/api/profiles", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetSettings returns server settings.
func (c *APIClient) GetSettings() (*models.ServerSettings, error) {
	var result models.ServerSettings
	if err := c.doGet("/api/settings", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Workspace ──

// ListWorkspace returns directory listing.
func (c *APIClient) ListWorkspace(sessionID, path string) ([]models.WorkspaceEntry, error) {
	reqPath := fmt.Sprintf("/api/list?session_id=%s&path=%s",
		url.QueryEscape(sessionID), url.QueryEscape(path))
	var result []models.WorkspaceEntry
	if err := c.doGet(reqPath, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetFile reads a text file from the workspace.
func (c *APIClient) GetFile(sessionID, path string) (*models.FileContent, error) {
	reqPath := fmt.Sprintf("/api/file?session_id=%s&path=%s",
		url.QueryEscape(sessionID), url.QueryEscape(path))
	var result models.FileContent
	if err := c.doGet(reqPath, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Read-Only Panels ──

// ListCronJobs returns scheduled tasks.
func (c *APIClient) ListCronJobs() ([]models.CronJob, error) {
	var result []models.CronJob
	if err := c.doGet("/api/crons", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListSkills returns installed agent skills.
func (c *APIClient) ListSkills() ([]models.Skill, error) {
	var result []models.Skill
	if err := c.doGet("/api/skills", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetMemory returns agent memory notes.
func (c *APIClient) GetMemory() ([]models.MemoryEntry, error) {
	var result []models.MemoryEntry
	if err := c.doGet("/api/memory", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ── Projects ──

// ListProjects returns all projects.
func (c *APIClient) ListProjects() ([]models.Project, error) {
	var result []models.Project
	if err := c.doGet("/api/projects", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ── Reasoning ──

// GetReasoning returns reasoning settings.
func (c *APIClient) GetReasoning() (*models.ReasoningSettings, error) {
	var result models.ReasoningSettings
	if err := c.doGet("/api/reasoning", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Upload ──

// UploadFile uploads a file to a session.
// filePath is the local path, filename is the remote name.
func (c *APIClient) UploadFile(sessionID, filePath, filename string) (*models.UploadResult, error) {
	// Multipart upload — simplified for gomobile (file content passed as []byte)
	// Full implementation in upload.go
	return nil, fmt.Errorf("client: UploadFile not yet implemented for gomobile")
}

// unused imports guard
var _ = strconv.Itoa
