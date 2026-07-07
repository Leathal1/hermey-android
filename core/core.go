// Package core is the Go shared library for Hermdroid.
// It is compiled to an Android AAR via gomobile bind and consumed by the
// Kotlin/Compose UI shell.
//
// Architecture:
//   - HermeyClient: gomobile-facing facade exposing login/logout and all 45
//     REST endpoints as typed methods.
//   - auth.Client: cookie jar + login/logout.
//   - client.APIClient: HTTP transport and error handling.
//   - endpoints: typed request/response wrappers for every endpoint.
//   - sse.Stream: Server-Sent Events chat stream with Kotlin callbacks.
//   - cache/cachebridge: offline bbolt cache (Phase 6).
//
// The JNI boundary uses callback interfaces + flat DTOs only.
// gomobile cannot expose goroutines, channels, or generics.
package core

import (
	"github.com/Leathal1/hermey-android/core/auth"
	"github.com/Leathal1/hermey-android/core/client"
	"github.com/Leathal1/hermey-android/core/endpoints"
	"github.com/Leathal1/hermey-android/core/models"
	"github.com/Leathal1/hermey-android/core/sse"
	"github.com/Leathal1/hermey-android/core/stream"
)

// Export types so gomobile generates bindings for the streaming bridge.
// Kotlin implements sse.EventListener and passes it to StreamManager.Start.
var (
	_ sse.EventListener = (*EventListenerProxy)(nil)
	_ *stream.Manager
)

// EventListenerProxy is a concrete implementation of sse.EventListener that
// forwards all events to an embedded listener. It exists purely so gomobile
// emits the EventListener interface in the AAR.
type EventListenerProxy struct {
	Listener sse.EventListener
}

func (p *EventListenerProxy) OnToken(text string) {
	if p.Listener != nil {
		p.Listener.OnToken(text)
	}
}
func (p *EventListenerProxy) OnToolCall(callJSON string) {
	if p.Listener != nil {
		p.Listener.OnToolCall(callJSON)
	}
}
func (p *EventListenerProxy) OnToolResult(resultJSON string) {
	if p.Listener != nil {
		p.Listener.OnToolResult(resultJSON)
	}
}
func (p *EventListenerProxy) OnReasoning(text string) {
	if p.Listener != nil {
		p.Listener.OnReasoning(text)
	}
}
func (p *EventListenerProxy) OnStreamEnd() {
	if p.Listener != nil {
		p.Listener.OnStreamEnd()
	}
}
func (p *EventListenerProxy) OnError(msg string) {
	if p.Listener != nil {
		p.Listener.OnError(msg)
	}
}
func (p *EventListenerProxy) OnCancel() {
	if p.Listener != nil {
		p.Listener.OnCancel()
	}
}


// Version is the Hermdroid core library version.
const Version = "0.1.0"

// HermeyClient is the main Android-facing client for the Hermdroid Go core.
type HermeyClient struct {
	baseURL    string
	authClient *auth.Client
	apiClient  *client.APIClient
}

// NewHermeyClient creates a new HermeyClient connected to the given server URL.
func NewHermeyClient(baseURL string) (*HermeyClient, error) {
	ac, err := auth.NewClient(baseURL)
	if err != nil {
		return nil, err
	}
	api := client.NewAPIClient(ac)
	api.SetBaseURL(baseURL)
	return &HermeyClient{
		baseURL:    baseURL,
		authClient: ac,
		apiClient:  api,
	}, nil
}

// BaseURL returns the configured server URL.
func (c *HermeyClient) BaseURL() string {
	return c.baseURL
}

// Version returns the core library version.
func (c *HermeyClient) Version() string {
	return Version
}

// IsAuthenticated reports whether a session cookie is present.
func (c *HermeyClient) IsAuthenticated() bool {
	return c.authClient.IsAuthenticated()
}

// Login authenticates with the configured server.
func (c *HermeyClient) Login(password string) error {
	return c.authClient.Login(password)
}

// Logout clears the server session and cookie jar.
func (c *HermeyClient) Logout() error {
	return c.authClient.Logout()
}

// AuthStatus checks whether the server requires authentication.
func (c *HermeyClient) AuthStatus() (*endpoints.AuthStatusResponse, error) {
	return endpoints.GetAuthStatus(c.apiClient)
}

// Health checks if the server is reachable.
func (c *HermeyClient) Health() error {
	return endpoints.Health(c.apiClient)
}

// ============================================================================
// Sessions
// ============================================================================

// ListSessions returns all sessions.
func (c *HermeyClient) ListSessions() (*endpoints.ListSessionsResponse, error) {
	return endpoints.ListSessions(c.apiClient)
}

// GetSession loads a session with optional messages.
func (c *HermeyClient) GetSession(sessionID string, messages bool, msgLimit int) (*endpoints.GetSessionResponse, error) {
	return endpoints.GetSession(c.apiClient, &endpoints.GetSessionRequest{
		SessionID: sessionID,
		Messages:  messages,
		MsgLimit:  msgLimit,
	})
}

// NewSession creates a new session.
func (c *HermeyClient) NewSession(workspace, model, modelProvider, profile string) (*models.Session, error) {
	return endpoints.NewSession(c.apiClient, &endpoints.NewSessionRequest{
		Workspace:     workspace,
		Model:         model,
		ModelProvider: modelProvider,
		Profile:       profile,
	})
}

// RenameSession renames a session.
func (c *HermeyClient) RenameSession(sessionID, title string) error {
	return endpoints.RenameSession(c.apiClient, &endpoints.RenameSessionRequest{
		SessionID: sessionID,
		Title:     title,
	})
}

// DeleteSession deletes a session.
func (c *HermeyClient) DeleteSession(sessionID string) error {
	return endpoints.DeleteSession(c.apiClient, &endpoints.DeleteSessionRequest{SessionID: sessionID})
}

// PinSession pins or unpins a session.
func (c *HermeyClient) PinSession(sessionID string, pinned bool) error {
	return endpoints.PinSession(c.apiClient, &endpoints.PinSessionRequest{SessionID: sessionID, Pinned: pinned})
}

// ArchiveSession archives or unarchives a session.
func (c *HermeyClient) ArchiveSession(sessionID string, archived bool) error {
	return endpoints.ArchiveSession(c.apiClient, &endpoints.ArchiveSessionRequest{SessionID: sessionID, Archived: archived})
}

// ForkSession forks a session.
func (c *HermeyClient) ForkSession(sessionID, title string) (*endpoints.ForkSessionResponse, error) {
	return endpoints.ForkSession(c.apiClient, &endpoints.ForkSessionRequest{SessionID: sessionID, Title: title})
}

// MergeSession merges two sessions.
func (c *HermeyClient) MergeSession(sourceID, targetID string) error {
	return endpoints.MergeSession(c.apiClient, &endpoints.MergeSessionRequest{SourceID: sourceID, TargetID: targetID})
}

// ExportSession exports a session as raw JSON bytes.
func (c *HermeyClient) ExportSession(sessionID, format string) ([]byte, error) {
	return endpoints.ExportSession(c.apiClient, &endpoints.ExportSessionRequest{SessionID: sessionID, Format: format})
}

// ImportSession imports a previously exported session.
func (c *HermeyClient) ImportSession(data []byte, title, projectID string) (*endpoints.ImportSessionResponse, error) {
	return endpoints.ImportSession(c.apiClient, &endpoints.ImportSessionRequest{
		Data:      data,
		Title:     title,
		ProjectID: projectID,
	})
}

// CompactSession compacts a session.
func (c *HermeyClient) CompactSession(sessionID string) error {
	return endpoints.CompactSession(c.apiClient, &endpoints.CompactSessionRequest{SessionID: sessionID})
}

// SearchSessions searches sessions by query.
func (c *HermeyClient) SearchSessions(query string, limit int) (*endpoints.SearchSessionsResponse, error) {
	return endpoints.SearchSessions(c.apiClient, &endpoints.SearchSessionsRequest{Query: query, Limit: limit})
}

// ============================================================================
// Chat
// ============================================================================

// StartChat starts a new assistant turn and returns the stream id.
func (c *HermeyClient) StartChat(sessionID, message, workspace, model string) (*models.StreamStartResponse, error) {
	return endpoints.StartChat(c.apiClient, &endpoints.StartChatRequest{
		SessionID: sessionID,
		Message:   message,
		Workspace: workspace,
		Model:     model,
	})
}

// CancelChat cancels an active chat stream.
func (c *HermeyClient) CancelChat(streamID string) error {
	return endpoints.CancelChat(c.apiClient, &endpoints.CancelChatRequest{StreamID: streamID})
}

// SteerChat sends a steering message to an active stream.
func (c *HermeyClient) SteerChat(sessionID, text string) (*models.SteerResponse, error) {
	return endpoints.SteerChat(c.apiClient, &endpoints.SteerChatRequest{SessionID: sessionID, Text: text})
}

// GetChatHistory returns chat messages for a session.
func (c *HermeyClient) GetChatHistory(sessionID string, limit int) (*endpoints.GetChatHistoryResponse, error) {
	return endpoints.GetChatHistory(c.apiClient, &endpoints.GetChatHistoryRequest{SessionID: sessionID, Limit: limit})
}

// SendFeedback submits feedback on a message.
func (c *HermeyClient) SendFeedback(messageID, rating, comment string) error {
	return endpoints.SendFeedback(c.apiClient, &endpoints.SendFeedbackRequest{
		MessageID: messageID,
		Rating:    rating,
		Comment:   comment,
	})
}

// StreamStatus checks if a stream is still active.
func (c *HermeyClient) StreamStatus(streamID string) (bool, error) {
	return endpoints.StreamStatus(c.apiClient, &endpoints.StreamStatusRequest{StreamID: streamID})
}

// SubscribeStream connects to the SSE chat stream and delivers events to the listener.
// This method blocks; call it from a background thread on Android.
func (c *HermeyClient) SubscribeStream(streamID string, listener sse.EventListener) error {
	stream := sse.NewStream(c.baseURL, streamID, c.apiClient.HTTPClient())
	return stream.Subscribe(listener)
}

// ============================================================================
// Workspace
// ============================================================================

// ListWorkspace lists files in a session workspace.
func (c *HermeyClient) ListWorkspace(sessionID, path string) ([]models.WorkspaceEntry, error) {
	return endpoints.ListWorkspace(c.apiClient, &endpoints.ListWorkspaceRequest{SessionID: sessionID, Path: path})
}

// GetFile reads a workspace file.
func (c *HermeyClient) GetFile(sessionID, filePath string) (*models.FileContent, error) {
	return endpoints.GetFile(c.apiClient, &endpoints.GetFileRequest{SessionID: sessionID, Path: filePath})
}

// UpdateFile writes content to a workspace file.
func (c *HermeyClient) UpdateFile(sessionID, filePath, content string) error {
	return endpoints.UpdateFile(c.apiClient, &endpoints.UpdateFileRequest{
		SessionID: sessionID,
		Path:      filePath,
		Content:   content,
	})
}

// DeleteFile deletes a workspace file.
func (c *HermeyClient) DeleteFile(sessionID, filePath string) error {
	return endpoints.DeleteFile(c.apiClient, &endpoints.DeleteFileRequest{SessionID: sessionID, Path: filePath})
}

// UploadFile uploads a file to the workspace.
func (c *HermeyClient) UploadFile(sessionID, filename string, content []byte, path string) (*models.UploadResult, error) {
	return endpoints.UploadFile(c.apiClient, &endpoints.UploadFileRequest{
		SessionID: sessionID,
		Filename:  filename,
		Content:   content,
		Path:      path,
	})
}

// ============================================================================
// Models / Providers / Profiles / Settings / Reasoning / Tools
// ============================================================================

// ListModels returns available models.
func (c *HermeyClient) ListModels() ([]models.ModelInfo, error) {
	return endpoints.ListModels(c.apiClient)
}

// ListProviders returns available model providers.
func (c *HermeyClient) ListProviders() ([]models.ProviderInfo, error) {
	return endpoints.ListProviders(c.apiClient)
}

// ListProfiles returns available agent profiles.
func (c *HermeyClient) ListProfiles() ([]models.ProfileInfo, error) {
	return endpoints.ListProfiles(c.apiClient)
}

// GetSettings returns server settings.
func (c *HermeyClient) GetSettings() (*models.ServerSettings, error) {
	return endpoints.GetSettings(c.apiClient)
}

// UpdateSettings updates server settings.
func (c *HermeyClient) UpdateSettings(botName, theme, defaultModel string) (*models.ServerSettings, error) {
	return endpoints.UpdateSettings(c.apiClient, &endpoints.UpdateSettingsRequest{
		BotName:      botName,
		Theme:        theme,
		DefaultModel: defaultModel,
	})
}

// GetReasoning returns reasoning settings.
func (c *HermeyClient) GetReasoning() (*models.ReasoningSettings, error) {
	return endpoints.GetReasoning(c.apiClient)
}

// UpdateReasoning updates reasoning settings.
func (c *HermeyClient) UpdateReasoning(display, effort string) (*models.ReasoningSettings, error) {
	return endpoints.UpdateReasoning(c.apiClient, &endpoints.UpdateReasoningRequest{Display: display, Effort: effort})
}

// ListMCPConfigs returns configured MCP servers.
func (c *HermeyClient) ListMCPConfigs() ([]models.MCPConfig, error) {
	return endpoints.ListMCPConfigs(c.apiClient)
}

// ListMCPTools returns tools exposed by MCP servers.
func (c *HermeyClient) ListMCPTools() ([]models.MCPTool, error) {
	return endpoints.ListMCPTools(c.apiClient)
}

// ListTools returns available tool configurations.
func (c *HermeyClient) ListTools() ([]models.ToolConfig, error) {
	return endpoints.ListTools(c.apiClient)
}

// ============================================================================
// Streaming state machine
// ============================================================================

// NewStreamManager returns a stream state machine bound to this client.
func (c *HermeyClient) NewStreamManager() *stream.Manager {
	return stream.NewManager(c.apiClient)
}

// StreamManager is an alias for stream.Manager for gomobile naming clarity.
type StreamManager = stream.Manager

// ============================================================================
// Read-Only Panels
// ============================================================================

// ListCronJobs returns scheduled tasks.
func (c *HermeyClient) ListCronJobs() ([]models.CronJob, error) {
	return endpoints.ListCronJobs(c.apiClient)
}

// ListSkills returns installed agent skills.
func (c *HermeyClient) ListSkills() ([]models.Skill, error) {
	return endpoints.ListSkills(c.apiClient)
}

// GetMemory returns agent memory notes.
func (c *HermeyClient) GetMemory() ([]models.MemoryEntry, error) {
	return endpoints.GetMemory(c.apiClient)
}

// ListProjects returns all projects.
func (c *HermeyClient) ListProjects() ([]models.Project, error) {
	return endpoints.ListProjects(c.apiClient)
}

// ListKanbanTasks returns kanban tasks.
func (c *HermeyClient) ListKanbanTasks() ([]models.KanbanTask, error) {
	return endpoints.ListKanbanTasks(c.apiClient)
}

// ListInboxItems returns agent inbox items.
func (c *HermeyClient) ListInboxItems() ([]models.InboxItem, error) {
	return endpoints.ListInboxItems(c.apiClient)
}

