// Package endpoints contains typed request/response structs and functions for
// every hermes-webui REST endpoint.  Functions accept an *client.APIClient and
// gomobile-friendly request structs, returning typed response structs.
//
// The package is organized by the API categories from the Hermdroid PRD:
//   - Auth & health (4)
//   - Sessions (13)
//   - Chat (7)
//   - Workspace (5)
//   - Models / Providers / Profiles / Settings / Tools (10)
//   - Read-only panels (6)
package endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Leathal1/hermey-android/core/client"
	"github.com/Leathal1/hermey-android/core/models"
)

// ============================================================================
// Auth & Health
// ============================================================================

// AuthStatusResponse is the response from /api/auth/status.
type AuthStatusResponse struct {
	AuthEnabled bool   `json:"auth_enabled"`
	LoggedIn    bool   `json:"logged_in,omitempty"`
	Message     string `json:"message,omitempty"`
}

// GetAuthStatus returns whether the server requires authentication.
func GetAuthStatus(c *client.APIClient) (*AuthStatusResponse, error) {
	var r AuthStatusResponse
	if err := c.DoGET("/api/auth/status", &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// LoginRequest wraps a login payload.
type LoginRequest struct {
	Password string `json:"password"`
}

// Login authenticates and stores the session cookie in the client's jar.
func Login(c *client.APIClient, req *LoginRequest) error {
	return c.DoPOST("/api/auth/login", req, nil)
}

// Logout clears the server session.
func Logout(c *client.APIClient) error {
	return c.DoPOST("/api/auth/logout", nil, nil)
}

// Health checks if the server is reachable.
func Health(c *client.APIClient) error {
	resp, err := c.HTTPClient().Get(c.BaseURL() + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: HTTP %d", resp.StatusCode)
	}
	return nil
}

// ============================================================================
// Sessions
// ============================================================================

// ListSessionsResponse is the response from GET /api/sessions.
type ListSessionsResponse struct {
	Sessions []models.Session `json:"sessions"`
}

// ListSessions returns all sessions.
func ListSessions(c *client.APIClient) (*ListSessionsResponse, error) {
	var r ListSessionsResponse
	if err := c.DoGET("/api/sessions", &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// GetSessionRequest is the query for GET /api/session.
type GetSessionRequest struct {
	SessionID string `json:"session_id"`
	Messages  bool   `json:"messages,omitempty"`
	MsgLimit  int    `json:"msg_limit,omitempty"`
}

// GetSessionResponse is the response from GET /api/session.
type GetSessionResponse struct {
	Session  models.Session   `json:"session"`
	Messages []models.ChatMessage `json:"messages,omitempty"`
}

// GetSession loads a session with optional recent messages.
func GetSession(c *client.APIClient, req *GetSessionRequest) (*GetSessionResponse, error) {
	q := url.Values{}
	q.Set("session_id", req.SessionID)
	q.Set("messages", strconv.FormatBool(req.Messages))
	if req.MsgLimit > 0 {
		q.Set("msg_limit", strconv.Itoa(req.MsgLimit))
	}
	var r GetSessionResponse
	if err := c.DoGET("/api/session?"+q.Encode(), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// NewSessionRequest is the payload for POST /api/session/new.
type NewSessionRequest struct {
	Workspace     string `json:"workspace,omitempty"`
	Model         string `json:"model,omitempty"`
	ModelProvider string `json:"model_provider,omitempty"`
	Profile       string `json:"profile,omitempty"`
	ProjectID     string `json:"project_id,omitempty"`
	Title         string `json:"title,omitempty"`
}

// NewSession creates a new session.
func NewSession(c *client.APIClient, req *NewSessionRequest) (*models.Session, error) {
	var r models.Session
	if err := c.DoPOST("/api/session/new", req, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// RenameSessionRequest renames a session.
type RenameSessionRequest struct {
	SessionID string `json:"session_id"`
	Title     string `json:"title"`
}

// RenameSession renames a session.
func RenameSession(c *client.APIClient, req *RenameSessionRequest) error {
	return c.DoPOST("/api/session/rename", req, nil)
}

// DeleteSessionRequest deletes a session.
type DeleteSessionRequest struct {
	SessionID string `json:"session_id"`
}

// DeleteSession deletes a session.
func DeleteSession(c *client.APIClient, req *DeleteSessionRequest) error {
	return c.DoPOST("/api/session/delete", req, nil)
}

// PinSessionRequest pins or unpins a session.
type PinSessionRequest struct {
	SessionID string `json:"session_id"`
	Pinned    bool   `json:"pinned"`
}

// PinSession pins or unpins a session.
func PinSession(c *client.APIClient, req *PinSessionRequest) error {
	return c.DoPOST("/api/session/pin", req, nil)
}

// ArchiveSessionRequest archives or unarchives a session.
type ArchiveSessionRequest struct {
	SessionID string `json:"session_id"`
	Archived  bool   `json:"archived"`
}

// ArchiveSession archives or unarchives a session.
func ArchiveSession(c *client.APIClient, req *ArchiveSessionRequest) error {
	return c.DoPOST("/api/session/archive", req, nil)
}

// ForkSessionRequest forks a session.
type ForkSessionRequest struct {
	SessionID string `json:"session_id"`
	Title     string `json:"title,omitempty"`
}

// ForkSessionResponse is the response from forking a session.
type ForkSessionResponse struct {
	NewSessionID string `json:"new_session_id"`
}

// ForkSession forks a session at its current state.
func ForkSession(c *client.APIClient, req *ForkSessionRequest) (*ForkSessionResponse, error) {
	var r ForkSessionResponse
	if err := c.DoPOST("/api/session/fork", req, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// MergeSessionRequest merges messages from one session into another.
type MergeSessionRequest struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
}

// MergeSession merges two sessions.
func MergeSession(c *client.APIClient, req *MergeSessionRequest) error {
	return c.DoPOST("/api/session/merge", req, nil)
}

// ExportSessionRequest exports a session to a portable format.
type ExportSessionRequest struct {
	SessionID string `json:"session_id"`
	Format    string `json:"format,omitempty"` // json, markdown, html
}

// ExportSession exports a session as raw JSON bytes.
func ExportSession(c *client.APIClient, req *ExportSessionRequest) ([]byte, error) {
	q := url.Values{}
	q.Set("session_id", req.SessionID)
	if req.Format != "" {
		q.Set("format", req.Format)
	}
	return c.DoRawGET("/api/session/export?" + q.Encode())
}

// ImportSessionRequest imports a previously exported session.
type ImportSessionRequest struct {
	Data      []byte `json:"data"`
	Title     string `json:"title,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
}

// ImportSessionResponse is the response from importing a session.
type ImportSessionResponse struct {
	SessionID string `json:"session_id"`
}

// ImportSession imports a session from exported bytes.
func ImportSession(c *client.APIClient, req *ImportSessionRequest) (*ImportSessionResponse, error) {
	var r ImportSessionResponse
	body := map[string]string{
		"data":       string(req.Data),
		"title":      req.Title,
		"project_id": req.ProjectID,
	}
	if err := c.DoPOST("/api/session/import", body, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// MoveSessionRequest moves a session to a project.
type MoveSessionRequest struct {
	SessionID string `json:"session_id"`
	ProjectID string `json:"project_id"`
}

// MoveSession moves a session to a different project.
func MoveSession(c *client.APIClient, req *MoveSessionRequest) error {
	return c.DoPOST("/api/session/move", req, nil)
}

// TruncateSessionRequest truncates a session at a message.
type TruncateSessionRequest struct {
	SessionID string `json:"session_id"`
	MessageID string `json:"message_id,omitempty"`
	KeepCount int    `json:"keep_count,omitempty"`
}

// TruncateSession truncates a session.
func TruncateSession(c *client.APIClient, req *TruncateSessionRequest) error {
	return c.DoPOST("/api/session/truncate", req, nil)
}

// CompactSessionRequest compacts a session's message history.
type CompactSessionRequest struct {
	SessionID string `json:"session_id"`
}

// CompactSession compacts session history.
func CompactSession(c *client.APIClient, req *CompactSessionRequest) error {
	return c.DoPOST("/api/session/compact", req, nil)
}

// SearchSessionsRequest searches sessions by title or content.
type SearchSessionsRequest struct {
	Query string `json:"q"`
	Limit int    `json:"limit,omitempty"`
}

// SearchSessionsResponse is the response from searching sessions.
type SearchSessionsResponse struct {
	Sessions []models.Session `json:"sessions"`
}

// SearchSessions searches session titles and messages.
func SearchSessions(c *client.APIClient, req *SearchSessionsRequest) (*SearchSessionsResponse, error) {
	q := url.Values{}
	q.Set("q", req.Query)
	if req.Limit > 0 {
		q.Set("limit", strconv.Itoa(req.Limit))
	}
	var r SearchSessionsResponse
	if err := c.DoGET("/api/session/search?"+q.Encode(), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// ============================================================================
// Chat
// ============================================================================

// StartChatRequest starts a new assistant turn.
type StartChatRequest struct {
	SessionID     string `json:"session_id"`
	Message       string `json:"message"`
	Workspace     string `json:"workspace,omitempty"`
	Model         string `json:"model,omitempty"`
	ModelProvider string `json:"model_provider,omitempty"`
	Profile       string `json:"profile,omitempty"`
}

// StartChat starts a chat and returns the stream id.
func StartChat(c *client.APIClient, req *StartChatRequest) (*models.StreamStartResponse, error) {
	var r models.StreamStartResponse
	if err := c.DoPOST("/api/chat/start", req, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// CancelChatRequest cancels an active chat stream.
type CancelChatRequest struct {
	StreamID string `json:"stream_id"`
}

// CancelChat cancels a chat stream.
func CancelChat(c *client.APIClient, req *CancelChatRequest) error {
	q := url.Values{}
	q.Set("stream_id", req.StreamID)
	resp, err := c.HTTPClient().Get(c.BaseURL() + "/api/chat/cancel?" + q.Encode())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cancel chat failed: HTTP %d", resp.StatusCode)
	}
	return nil
}

// SteerChatRequest sends a steering message to an active stream.
type SteerChatRequest struct {
	SessionID string `json:"session_id"`
	Text      string `json:"text"`
}

// SteerChat steers an active chat stream.
func SteerChat(c *client.APIClient, req *SteerChatRequest) (*models.SteerResponse, error) {
	var r models.SteerResponse
	if err := c.DoPOST("/api/chat/steer", req, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// GetChatHistoryRequest loads chat history for a session.
type GetChatHistoryRequest struct {
	SessionID string `json:"session_id"`
	Limit     int    `json:"limit,omitempty"`
}

// GetChatHistoryResponse is the response from /api/chat/history.
type GetChatHistoryResponse struct {
	Messages []models.ChatMessage `json:"messages"`
}

// GetChatHistory returns chat messages for a session.
func GetChatHistory(c *client.APIClient, req *GetChatHistoryRequest) (*GetChatHistoryResponse, error) {
	q := url.Values{}
	q.Set("session_id", req.SessionID)
	if req.Limit > 0 {
		q.Set("limit", strconv.Itoa(req.Limit))
	}
	var r GetChatHistoryResponse
	if err := c.DoGET("/api/chat/history?"+q.Encode(), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// SendFeedbackRequest submits message feedback.
type SendFeedbackRequest struct {
	MessageID string `json:"message_id"`
	Rating    string `json:"rating"` // up, down
	Comment   string `json:"comment,omitempty"`
}

// SendFeedback submits feedback on a message.
func SendFeedback(c *client.APIClient, req *SendFeedbackRequest) error {
	return c.DoPOST("/api/chat/feedback", req, nil)
}

// StreamStatusRequest checks whether a stream is still active.
type StreamStatusRequest struct {
	StreamID string `json:"stream_id"`
}

// StreamStatus checks if a stream is still active.
func StreamStatus(c *client.APIClient, req *StreamStatusRequest) (bool, error) {
	q := url.Values{}
	q.Set("stream_id", req.StreamID)
	resp, err := c.HTTPClient().Get(c.BaseURL() + "/api/chat/stream/status?" + q.Encode())
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

// ============================================================================
// Workspace
// ============================================================================

// ListWorkspaceRequest lists files under a workspace path.
type ListWorkspaceRequest struct {
	SessionID string `json:"session_id"`
	Path      string `json:"path,omitempty"`
}

// ListWorkspace lists files in a session workspace.
func ListWorkspace(c *client.APIClient, req *ListWorkspaceRequest) ([]models.WorkspaceEntry, error) {
	q := url.Values{}
	q.Set("session_id", req.SessionID)
	if req.Path != "" {
		q.Set("path", req.Path)
	}
	var r []models.WorkspaceEntry
	if err := c.DoGET("/api/list?"+q.Encode(), &r); err != nil {
		return nil, err
	}
	return r, nil
}

// GetFileRequest reads a workspace file.
type GetFileRequest struct {
	SessionID string `json:"session_id"`
	Path      string `json:"path"`
}

// GetFile reads a workspace file.
func GetFile(c *client.APIClient, req *GetFileRequest) (*models.FileContent, error) {
	q := url.Values{}
	q.Set("session_id", req.SessionID)
	q.Set("path", req.Path)
	var r models.FileContent
	if err := c.DoGET("/api/file?"+q.Encode(), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// UpdateFileRequest updates a workspace file.
type UpdateFileRequest struct {
	SessionID string `json:"session_id"`
	Path      string `json:"path"`
	Content   string `json:"content"`
}

// UpdateFile writes content to a workspace file.
func UpdateFile(c *client.APIClient, req *UpdateFileRequest) error {
	return c.DoPOST("/api/file/update", req, nil)
}

// DeleteFileRequest deletes a workspace file.
type DeleteFileRequest struct {
	SessionID string `json:"session_id"`
	Path      string `json:"path"`
}

// DeleteFile deletes a workspace file.
func DeleteFile(c *client.APIClient, req *DeleteFileRequest) error {
	return c.DoPOST("/api/file/delete", req, nil)
}

// UploadFileRequest uploads a file to a session workspace.
type UploadFileRequest struct {
	SessionID string `json:"session_id"`
	Filename  string `json:"filename"`
	Content   []byte `json:"content"`
	Path      string `json:"path,omitempty"`
}

// UploadFile uploads a file to the workspace.
func UploadFile(c *client.APIClient, req *UploadFileRequest) (*models.UploadResult, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("session_id", req.SessionID)
	if req.Path != "" {
		_ = writer.WriteField("path", req.Path)
	}
	part, err := writer.CreateFormFile("file", path.Base(req.Filename))
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(req.Content); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL()+"/api/upload", &buf)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.HTTPClient().Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("upload failed: HTTP %d: %s", resp.StatusCode, string(body))
	}
	var r models.UploadResult
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

// ============================================================================
// Models / Providers / Profiles / Settings / Reasoning / Tools
// ============================================================================

// ListModels returns available models.
func ListModels(c *client.APIClient) ([]models.ModelInfo, error) {
	var r []models.ModelInfo
	if err := c.DoGET("/api/models", &r); err != nil {
		return nil, err
	}
	return r, nil
}

// ListProviders returns available model providers.
func ListProviders(c *client.APIClient) ([]models.ProviderInfo, error) {
	var r []models.ProviderInfo
	if err := c.DoGET("/api/providers", &r); err != nil {
		return nil, err
	}
	return r, nil
}

// ListProfiles returns available agent profiles.
func ListProfiles(c *client.APIClient) ([]models.ProfileInfo, error) {
	var r []models.ProfileInfo
	if err := c.DoGET("/api/profiles", &r); err != nil {
		return nil, err
	}
	return r, nil
}

// GetSettings returns server settings.
func GetSettings(c *client.APIClient) (*models.ServerSettings, error) {
	var r models.ServerSettings
	if err := c.DoGET("/api/settings", &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// UpdateSettingsRequest updates server settings.
type UpdateSettingsRequest struct {
	BotName     string   `json:"bot_name,omitempty"`
	Theme       string   `json:"theme,omitempty"`
	DefaultModel string  `json:"default_model,omitempty"`
}

// UpdateSettings updates server settings.
func UpdateSettings(c *client.APIClient, req *UpdateSettingsRequest) (*models.ServerSettings, error) {
	var r models.ServerSettings
	if err := c.DoPOST("/api/settings/update", req, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// GetReasoning returns reasoning settings.
func GetReasoning(c *client.APIClient) (*models.ReasoningSettings, error) {
	var r models.ReasoningSettings
	if err := c.DoGET("/api/reasoning", &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// UpdateReasoningRequest updates reasoning settings.
type UpdateReasoningRequest struct {
	Display string `json:"display,omitempty"`
	Effort  string `json:"effort,omitempty"`
}

// UpdateReasoning updates reasoning settings.
func UpdateReasoning(c *client.APIClient, req *UpdateReasoningRequest) (*models.ReasoningSettings, error) {
	var r models.ReasoningSettings
	if err := c.DoPOST("/api/reasoning/update", req, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// ListMCPConfigs returns configured MCP servers.
func ListMCPConfigs(c *client.APIClient) ([]models.MCPConfig, error) {
	var r []models.MCPConfig
	if err := c.DoGET("/api/mcp", &r); err != nil {
		return nil, err
	}
	return r, nil
}

// ListMCPTools returns tools exposed by MCP servers.
func ListMCPTools(c *client.APIClient) ([]models.MCPTool, error) {
	var r []models.MCPTool
	if err := c.DoGET("/api/mcp/tools", &r); err != nil {
		return nil, err
	}
	return r, nil
}

// ListTools returns available tool configurations.
func ListTools(c *client.APIClient) ([]models.ToolConfig, error) {
	var r []models.ToolConfig
	if err := c.DoGET("/api/tools", &r); err != nil {
		return nil, err
	}
	return r, nil
}

// ============================================================================
// Read-Only Panels
// ============================================================================

// ListCronJobs returns scheduled tasks.
func ListCronJobs(c *client.APIClient) ([]models.CronJob, error) {
	var r []models.CronJob
	if err := c.DoGET("/api/crons", &r); err != nil {
		return nil, err
	}
	return r, nil
}

// ListSkills returns installed agent skills.
func ListSkills(c *client.APIClient) ([]models.Skill, error) {
	var r []models.Skill
	if err := c.DoGET("/api/skills", &r); err != nil {
		return nil, err
	}
	return r, nil
}

// GetMemory returns agent memory notes.
func GetMemory(c *client.APIClient) ([]models.MemoryEntry, error) {
	var r []models.MemoryEntry
	if err := c.DoGET("/api/memory", &r); err != nil {
		return nil, err
	}
	return r, nil
}

// ListProjects returns all projects.
func ListProjects(c *client.APIClient) ([]models.Project, error) {
	var r []models.Project
	if err := c.DoGET("/api/projects", &r); err != nil {
		return nil, err
	}
	return r, nil
}

// ListKanbanTasks returns kanban tasks.
func ListKanbanTasks(c *client.APIClient) ([]models.KanbanTask, error) {
	var r []models.KanbanTask
	if err := c.DoGET("/api/kanban", &r); err != nil {
		return nil, err
	}
	return r, nil
}

// ListInboxItems returns agent inbox items.
func ListInboxItems(c *client.APIClient) ([]models.InboxItem, error) {
	var r []models.InboxItem
	if err := c.DoGET("/api/inbox", &r); err != nil {
		return nil, err
	}
	return r, nil
}

// ============================================================================
// Utility helpers
// ============================================================================

// ParseSSEData parses a Server-Sent Event data line into a ChatEvent.
func ParseSSEData(eventType, data string) *models.ChatEvent {
	if !strings.HasPrefix(eventType, "data:") {
		return nil
	}
	payload := strings.TrimPrefix(eventType, "data:")
	payload = strings.TrimSpace(payload)
	var ev models.ChatEvent
	_ = json.Unmarshal([]byte(payload), &ev)
	return &ev
}

// unused import guard
var _ = time.Now
