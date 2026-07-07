package sse

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// mockListener captures all events delivered by the SSE stream.
type mockListener struct {
	mu       sync.Mutex
	tokens   []string
	toolCalls []string
	reasoning []string
	streamEnd bool
	errors   []string
	cancelled bool
}

func (m *mockListener) OnToken(text string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens = append(m.tokens, text)
}
func (m *mockListener) OnToolCall(callJSON string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.toolCalls = append(m.toolCalls, callJSON)
}
func (m *mockListener) OnReasoning(text string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reasoning = append(m.reasoning, text)
}
func (m *mockListener) OnStreamEnd() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streamEnd = true
}
func (m *mockListener) OnError(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors = append(m.errors, msg)
}
func (m *mockListener) OnCancel() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cancelled = true
}

func newMockServer(t *testing.T, events string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("server does not support flushing")
		}
		fmt.Fprint(w, events)
		flusher.Flush()
	}))
}

func TestSSETokenEvents(t *testing.T) {
	events := "event: token\ndata: {\"type\":\"token\",\"content\":\"hello\"}\n\nevent: token\ndata: {\"type\":\"token\",\"content\":\" world\"}\n\n"
	srv := newMockServer(t, events)
	defer srv.Close()

	stream := NewStream(srv.URL, "test-stream", srv.Client())
	ml := &mockListener{}
	if err := stream.Subscribe(ml); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	ml.mu.Lock()
	defer ml.mu.Unlock()
	if len(ml.tokens) != 2 {
		t.Fatalf("tokens = %d, want 2", len(ml.tokens))
	}
	if ml.tokens[0] != "hello" {
		t.Errorf("tokens[0] = %q, want hello", ml.tokens[0])
	}
	if ml.tokens[1] != " world" {
		t.Errorf("tokens[1] = %q, want ' world'", ml.tokens[1])
	}
	if !ml.streamEnd {
		t.Error("streamEnd = false, want true")
	}
}

func TestSSEToolCallEvent(t *testing.T) {
	events := "event: tool_call\ndata: {\"id\":\"tc1\",\"name\":\"search\"}\n\n"
	srv := newMockServer(t, events)
	defer srv.Close()

	stream := NewStream(srv.URL, "test-stream", srv.Client())
	ml := &mockListener{}
	if err := stream.Subscribe(ml); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	ml.mu.Lock()
	defer ml.mu.Unlock()
	if len(ml.toolCalls) != 1 {
		t.Fatalf("toolCalls = %d, want 1", len(ml.toolCalls))
	}
}

func TestSSEReasoningEvent(t *testing.T) {
	events := "event: reasoning\ndata: {\"type\":\"reasoning\",\"content\":\"thinking...\"}\n\n"
	srv := newMockServer(t, events)
	defer srv.Close()

	stream := NewStream(srv.URL, "test-stream", srv.Client())
	ml := &mockListener{}
	if err := stream.Subscribe(ml); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	ml.mu.Lock()
	defer ml.mu.Unlock()
	if len(ml.reasoning) != 1 {
		t.Fatalf("reasoning = %d, want 1", len(ml.reasoning))
	}
	if ml.reasoning[0] != "thinking..." {
		t.Errorf("reasoning[0] = %q, want 'thinking...'", ml.reasoning[0])
	}
}

func TestSSEHeartbeatIgnored(t *testing.T) {
	events := ": heartbeat comment\nevent: token\ndata: {\"type\":\"token\",\"content\":\"ok\"}\n\n"
	srv := newMockServer(t, events)
	defer srv.Close()

	stream := NewStream(srv.URL, "test-stream", srv.Client())
	ml := &mockListener{}
	if err := stream.Subscribe(ml); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	ml.mu.Lock()
	defer ml.mu.Unlock()
	if len(ml.tokens) != 1 {
		t.Fatalf("tokens = %d, want 1 (heartbeat should be ignored)", len(ml.tokens))
	}
	if ml.tokens[0] != "ok" {
		t.Errorf("tokens[0] = %q, want ok", ml.tokens[0])
	}
}

func TestSSEErrorEvent(t *testing.T) {
	events := "event: error\ndata: {\"type\":\"error\",\"error\":\"something broke\"}\n\n"
	srv := newMockServer(t, events)
	defer srv.Close()

	stream := NewStream(srv.URL, "test-stream", srv.Client())
	ml := &mockListener{}
	if err := stream.Subscribe(ml); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	ml.mu.Lock()
	defer ml.mu.Unlock()
	if len(ml.errors) != 1 {
		t.Fatalf("errors = %d, want 1", len(ml.errors))
	}
	if ml.errors[0] != "something broke" {
		t.Errorf("errors[0] = %q, want 'something broke'", ml.errors[0])
	}
}

func TestSSEStreamEnd(t *testing.T) {
	events := "event: stream_end\ndata: {}\n\n"
	srv := newMockServer(t, events)
	defer srv.Close()

	stream := NewStream(srv.URL, "test-stream", srv.Client())
	ml := &mockListener{}
	if err := stream.Subscribe(ml); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	ml.mu.Lock()
	defer ml.mu.Unlock()
	if !ml.streamEnd {
		t.Error("streamEnd = false, want true")
	}
}

func TestStreamStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	active, err := StreamStatus(srv.URL, "test-stream", srv.Client())
	if err != nil {
		t.Fatalf("stream status: %v", err)
	}
	if !active {
		t.Error("active = false, want true")
	}
}
