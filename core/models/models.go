// Package models defines the data types for the hermes-webui API.
// All types use lenient JSON decoding with json.RawMessage fallbacks
// to handle upstream JSON drift without breaking the client.
package models

import "time"

// Session represents a chat session from /api/sessions.
type Session struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastMessageAt  time.Time `json:"last_message_at,omitempty"`
	Model          string    `json:"model,omitempty"`
	ModelProvider  string    `json:"model_provider,omitempty"`
	Profile        string    `json:"profile,omitempty"`
	Workspace      string    `json:"workspace,omitempty"`
	Pinned         bool      `json:"pinned"`
	Archived       bool      `json:"archived"`
	ProjectID      string    `json:"project_id,omitempty"`
	InputTokens    int64     `json:"input_tokens,omitempty"`
	OutputTokens   int64     `json:"output_tokens,omitempty"`
	EstimatedCost  float64   `json:"estimated_cost,omitempty"`
	MessageCount   int       `json:"message_count,omitempty"`
}

// ChatMessage represents a single message in a session.
type ChatMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user", "assistant", "system", "tool"
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool invocation by the agent.
type ToolCall struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Input    string `json:"input,omitempty"`
	Result   string `json:"result,omitempty"`
	Status   string `json:"status,omitempty"` // "pending", "running", "done", "error"
}

// ChatEvent represents a decoded SSE event from the chat stream.
type ChatEvent struct {
	Type    string    `json:"type"` // token, tool_call, tool_result, reasoning, stream_end, error, cancel
	Content string    `json:"content,omitempty"`
	Call    *ToolCall `json:"call,omitempty"`
	Error   string    `json:"error,omitempty"`
}

// ModelInfo represents a configured model from /api/models.
type ModelInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

// ProfileInfo represents an agent profile from /api/profiles.
type ProfileInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Active      bool   `json:"active,omitempty"`
}

// WorkspaceEntry represents a file or directory from /api/list.
type WorkspaceEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size,omitempty"`
	Mime  string `json:"mime,omitempty"`
}

// FileContent represents a file's content from /api/file.
type FileContent struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Mime     string `json:"mime,omitempty"`
	Size     int64  `json:"size,omitempty"`
	IsBinary bool   `json:"is_binary,omitempty"`
}

// CronJob represents a scheduled task from /api/crons.
type CronJob struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Schedule    string    `json:"schedule"`
	Script      string    `json:"script,omitempty"`
	Enabled     bool      `json:"enabled"`
	LastRunAt   time.Time `json:"last_run_at,omitempty"`
	LastStatus  string    `json:"last_status,omitempty"`
	NextRunAt   time.Time `json:"next_run_at,omitempty"`
}

// Skill represents an agent skill from /api/skills.
type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category,omitempty"`
}

// MemoryEntry represents a memory note from /api/memory.
type MemoryEntry struct {
	Content string `json:"content"`
	Target  string `json:"target"` // "memory" or "user"
}

// ServerSettings represents server configuration from /api/settings.
type ServerSettings struct {
	BotName string `json:"bot_name,omitempty"`
	Version string `json:"version,omitempty"`
	Theme   string `json:"theme,omitempty"`
}

// Project represents a project from /api/projects.
type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UploadResult represents the response from /api/upload.
type UploadResult struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Mime     string `json:"mime"`
	Size     int64  `json:"size"`
	IsImage  bool   `json:"is_image"`
}

// ReasoningSettings represents reasoning configuration from /api/reasoning.
type ReasoningSettings struct {
	Display string `json:"display,omitempty"`
	Effort  string `json:"effort,omitempty"`
}

// StreamStartResponse is the response from POST /api/chat/start.
type StreamStartResponse struct {
	StreamID  string `json:"stream_id"`
	SessionID string `json:"session_id"`
}

// SteerResponse is the response from POST /api/chat/steer.
type SteerResponse struct {
	Accepted bool   `json:"accepted"`
	Fallback string `json:"fallback,omitempty"`
	StreamID string `json:"stream_id,omitempty"`
}
