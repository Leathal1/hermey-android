package models

import (
	"time"
)

// AuthStatus is the response from /api/auth/status.
type AuthStatus struct {
	AuthEnabled bool   `json:"auth_enabled"`
	LoggedIn    bool   `json:"logged_in,omitempty"`
	Message     string `json:"message,omitempty"`
}

// UnmarshalJSON decodes AuthStatus leniently.
func (a *AuthStatus) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, a)
}

// Session represents a chat session.
type Session struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
	LastMessageAt time.Time `json:"last_message_at,omitempty"`
	Model         string    `json:"model,omitempty"`
	ModelProvider string    `json:"model_provider,omitempty"`
	Profile       string    `json:"profile,omitempty"`
	Workspace     string    `json:"workspace,omitempty"`
	Pinned        bool      `json:"pinned,omitempty"`
	Archived      bool      `json:"archived,omitempty"`
	ProjectID     string    `json:"project_id,omitempty"`
	InputTokens   int64     `json:"input_tokens,omitempty"`
	OutputTokens  int64     `json:"output_tokens,omitempty"`
	EstimatedCost float64   `json:"estimated_cost,omitempty"`
	MessageCount  int       `json:"message_count,omitempty"`
}

// UnmarshalJSON decodes Session leniently.
func (s *Session) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, s)
}

// ChatMessage is a single message in a session.
type ChatMessage struct {
	ID        string     `json:"id"`
	Role      string     `json:"role"`
	Content   string     `json:"content,omitempty"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// UnmarshalJSON decodes ChatMessage leniently.
func (m *ChatMessage) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, m)
}

// ToolCall represents a tool invocation.
type ToolCall struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Input  string `json:"input,omitempty"`
	Result string `json:"result,omitempty"`
	Status string `json:"status,omitempty"`
}

// UnmarshalJSON decodes ToolCall leniently.
func (t *ToolCall) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, t)
}

// ChatEvent is a decoded SSE event from the chat stream.
type ChatEvent struct {
	Type    string    `json:"type"`
	Content string    `json:"content,omitempty"`
	Call    *ToolCall `json:"call,omitempty"`
	Error   string    `json:"error,omitempty"`
}

// UnmarshalJSON decodes ChatEvent leniently.
func (c *ChatEvent) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, c)
}

// ModelInfo is a configured model.
type ModelInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
}

// UnmarshalJSON decodes ModelInfo leniently.
func (m *ModelInfo) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, m)
}

// ProviderInfo is a configured provider.
type ProviderInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Enabled bool   `json:"enabled,omitempty"`
}

// UnmarshalJSON decodes ProviderInfo leniently.
func (p *ProviderInfo) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, p)
}

// ProfileInfo is an agent profile.
type ProfileInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Active      bool   `json:"active,omitempty"`
}

// UnmarshalJSON decodes ProfileInfo leniently.
func (p *ProfileInfo) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, p)
}

// WorkspaceEntry is a file or directory entry.
type WorkspaceEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path,omitempty"`
	IsDir bool   `json:"is_dir,omitempty"`
	Size  int64  `json:"size,omitempty"`
	Mime  string `json:"mime,omitempty"`
}

// UnmarshalJSON decodes WorkspaceEntry leniently.
func (w *WorkspaceEntry) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, w)
}

// FileContent is a workspace file's content.
type FileContent struct {
	Path     string `json:"path"`
	Content  string `json:"content,omitempty"`
	Mime     string `json:"mime,omitempty"`
	Size     int64  `json:"size,omitempty"`
	IsBinary bool   `json:"is_binary,omitempty"`
}

// UnmarshalJSON decodes FileContent leniently.
func (f *FileContent) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, f)
}

// CronJob is a scheduled task.
type CronJob struct {
	ID         string    `json:"id"`
	Name       string    `json:"name,omitempty"`
	Schedule   string    `json:"schedule,omitempty"`
	Script     string    `json:"script,omitempty"`
	Enabled    bool      `json:"enabled,omitempty"`
	LastRunAt  time.Time `json:"last_run_at,omitempty"`
	LastStatus string    `json:"last_status,omitempty"`
	NextRunAt  time.Time `json:"next_run_at,omitempty"`
}

// UnmarshalJSON decodes CronJob leniently.
func (c *CronJob) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, c)
}

// Skill is an installed agent skill.
type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

// UnmarshalJSON decodes Skill leniently.
func (s *Skill) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, s)
}

// MemoryEntry is an agent memory note.
type MemoryEntry struct {
	Content string `json:"content"`
	Target  string `json:"target,omitempty"`
}

// UnmarshalJSON decodes MemoryEntry leniently.
func (m *MemoryEntry) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, m)
}

// ServerSettings is server configuration.
type ServerSettings struct {
	BotName     string   `json:"bot_name,omitempty"`
	Version     string   `json:"version,omitempty"`
	Theme       string   `json:"theme,omitempty"`
	AuthEnabled bool     `json:"auth_enabled,omitempty"`
	Models      []string `json:"models,omitempty"`
	Profiles    []string `json:"profiles,omitempty"`
}

// UnmarshalJSON decodes ServerSettings leniently.
func (s *ServerSettings) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, s)
}

// Project is a project.
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// UnmarshalJSON decodes Project leniently.
func (p *Project) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, p)
}

// KanbanTask is a task on the kanban board.
type KanbanTask struct {
	ID       string `json:"id"`
	Title    string `json:"title,omitempty"`
	Status   string `json:"status,omitempty"`
	Assignee string `json:"assignee,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

// UnmarshalJSON decodes KanbanTask leniently.
func (k *KanbanTask) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, k)
}

// InboxItem is an item in the agent inbox.
type InboxItem struct {
	ID        string    `json:"id"`
	Title     string    `json:"title,omitempty"`
	Body      string    `json:"body,omitempty"`
	Source    string    `json:"source,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	Read      bool      `json:"read,omitempty"`
}

// UnmarshalJSON decodes InboxItem leniently.
func (i *InboxItem) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, i)
}

// ReasoningSettings is reasoning configuration.
type ReasoningSettings struct {
	Display string `json:"display,omitempty"`
	Effort  string `json:"effort,omitempty"`
}

// UnmarshalJSON decodes ReasoningSettings leniently.
func (r *ReasoningSettings) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, r)
}

// MCPConfig is an MCP server configuration.
type MCPConfig struct {
	Name    string `json:"name"`
	Command string `json:"command,omitempty"`
	Args    string `json:"args,omitempty"`
	Env     string `json:"env,omitempty"`
	Enabled bool   `json:"enabled,omitempty"`
}

// UnmarshalJSON decodes MCPConfig leniently.
func (m *MCPConfig) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, m)
}

// MCPTool is a tool exposed by an MCP server.
type MCPTool struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Server      string `json:"server,omitempty"`
}

// UnmarshalJSON decodes MCPTool leniently.
func (m *MCPTool) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, m)
}

// ToolConfig is an available tool configuration.
type ToolConfig struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
}

// UnmarshalJSON decodes ToolConfig leniently.
func (t *ToolConfig) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, t)
}

// StreamStartResponse is returned by POST /api/chat/start.
type StreamStartResponse struct {
	StreamID  string `json:"stream_id"`
	SessionID string `json:"session_id,omitempty"`
}

// UnmarshalJSON decodes StreamStartResponse leniently.
func (s *StreamStartResponse) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, s)
}

// SteerResponse is returned by POST /api/chat/steer.
type SteerResponse struct {
	Accepted bool   `json:"accepted,omitempty"`
	Fallback string `json:"fallback,omitempty"`
	StreamID string `json:"stream_id,omitempty"`
}

// UnmarshalJSON decodes SteerResponse leniently.
func (s *SteerResponse) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, s)
}

// UploadResult is returned by file uploads.
type UploadResult struct {
	Filename string `json:"filename"`
	Path     string `json:"path,omitempty"`
	Mime     string `json:"mime,omitempty"`
	Size     int64  `json:"size,omitempty"`
	IsImage  bool   `json:"is_image,omitempty"`
}

// UnmarshalJSON decodes UploadResult leniently.
func (u *UploadResult) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, u)
}

// HealthStatus is returned by /health.
type HealthStatus struct {
	Status    string `json:"status"`
	Version   string `json:"version,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// UnmarshalJSON decodes HealthStatus leniently.
func (h *HealthStatus) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, h)
}

// Tag is a session tag.
type Tag struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// UnmarshalJSON decodes Tag leniently.
func (t *Tag) UnmarshalJSON(data []byte) error {
	return LenientUnmarshal(data, t)
}

// RawJSON is a wrapper around json.RawMessage for gomobile-friendly
// transport of arbitrary payloads.
type RawJSON struct {
	Bytes []byte `json:"bytes"`
}

// String returns the raw JSON as a string.
func (r *RawJSON) String() string {
	if r == nil {
		return ""
	}
	return string(r.Bytes)
}
