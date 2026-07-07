package models

import (
	"encoding/json"
	"testing"
)

func TestSessionLenientDecoding(t *testing.T) {
	// Extra unknown field "future_field" should not break decoding.
	raw := `{
		"id": "s1",
		"title": "Test Session",
		"created_at": "2026-07-07T12:00:00Z",
		"updated_at": "2026-07-07T12:01:00Z",
		"pinned": true,
		"archived": false,
		"future_field": "ignored",
		"message_count": 42
	}`
	var s Session
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if s.ID != "s1" {
		t.Errorf("ID = %q, want s1", s.ID)
	}
	if s.Title != "Test Session" {
		t.Errorf("Title = %q, want Test Session", s.Title)
	}
	if !s.Pinned {
		t.Error("Pinned = false, want true")
	}
	if s.MessageCount != 42 {
		t.Errorf("MessageCount = %d, want 42", s.MessageCount)
	}
	if s.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestSessionMissingFields(t *testing.T) {
	// Minimal JSON — optional fields should default cleanly.
	raw := `{"id": "s2", "title": "Min"}`
	var s Session
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if s.ID != "s2" {
		t.Errorf("ID = %q, want s2", s.ID)
	}
	if s.Model != "" {
		t.Errorf("Model = %q, want empty", s.Model)
	}
	if s.MessageCount != 0 {
		t.Errorf("MessageCount = %d, want 0", s.MessageCount)
	}
}

func TestChatMessageWithToolCalls(t *testing.T) {
	raw := `{
		"id": "m1",
		"role": "assistant",
		"content": "I called a tool",
		"tool_calls": [
			{"id": "tc1", "name": "search", "status": "done"}
		]
	}`
	var msg ChatMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(msg.ToolCalls) != 1 {
		t.Fatalf("ToolCalls = %d, want 1", len(msg.ToolCalls))
	}
	if msg.ToolCalls[0].Name != "search" {
		t.Errorf("ToolCalls[0].Name = %q, want search", msg.ToolCalls[0].Name)
	}
}

func TestChatEventExtraFields(t *testing.T) {
	// Extra unknown fields should be ignored (lenient in the "extra field" sense).
	raw := `{"type": "token", "content": "hello", "future_field": "ignored"}`
	var event ChatEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if event.Type != "token" {
		t.Errorf("Type = %q, want token", event.Type)
	}
	if event.Content != "hello" {
		t.Errorf("Content = %q, want hello", event.Content)
	}
}

func TestChatEventMissingFields(t *testing.T) {
	// Missing optional fields should default cleanly.
	raw := `{"type": "stream_end"}`
	var event ChatEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if event.Type != "stream_end" {
		t.Errorf("Type = %q, want stream_end", event.Type)
	}
	if event.Content != "" {
		t.Errorf("Content = %q, want empty", event.Content)
	}
	if event.Call != nil {
		t.Errorf("Call = %v, want nil", event.Call)
	}
}

func TestModelInfoDecoding(t *testing.T) {
	raw := `{"id": "gpt-4", "name": "GPT-4", "provider": "openai"}`
	var m ModelInfo
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m.Provider != "openai" {
		t.Errorf("Provider = %q, want openai", m.Provider)
	}
}

func TestStreamStartResponse(t *testing.T) {
	raw := `{"stream_id": "st1", "session_id": "s1"}`
	var r StreamStartResponse
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if r.StreamID != "st1" || r.SessionID != "s1" {
		t.Errorf("got stream=%q session=%q, want st1/s1", r.StreamID, r.SessionID)
	}
}
