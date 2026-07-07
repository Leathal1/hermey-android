package stream

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Leathal1/hermey-android/core/auth"
	"github.com/Leathal1/hermey-android/core/client"
	"github.com/Leathal1/hermey-android/core/endpoints"
	"github.com/Leathal1/hermey-android/core/sse"
)

type testListener struct {
	mu        sync.Mutex
	tokens    []string
	streamEnd int32
	cancelled int32
	errors    []string
}

func (l *testListener) OnToken(text string)      { l.mu.Lock(); l.tokens = append(l.tokens, text); l.mu.Unlock() }
func (l *testListener) OnToolCall(callJSON string) {}
func (l *testListener) OnToolResult(resultJSON string) {}
func (l *testListener) OnReasoning(text string)   {}
func (l *testListener) OnStreamEnd()               { atomic.AddInt32(&l.streamEnd, 1) }
func (l *testListener) OnError(msg string)         { l.mu.Lock(); l.errors = append(l.errors, msg); l.mu.Unlock() }
func (l *testListener) OnCancel()                  { atomic.AddInt32(&l.cancelled, 1) }

func newTestManager(ts *httptest.Server) *Manager {
	ac, _ := auth.NewClient(ts.URL)
	api := client.NewAPIClient(ac)
	api.SetBaseURL(ts.URL)
	return NewManager(api)
}

func TestStreamManagerStartAndEnd(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/chat/start":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"stream_id":"str-1","session_id":"ses-1"}`)
		case "/api/chat/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: token\ndata: {\"content\":\"hi\"}\n\n")
			fmt.Fprint(w, "event: stream_end\ndata: {}\n\n")
			w.(http.Flusher).Flush()
		default:
			http.NotFound(w, r)
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	m := newTestManager(ts)
	l := &testListener{}

	resp, err := m.Start(&endpoints.StartChatRequest{SessionID: "ses-1", Message: "hello"}, l)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if resp.StreamID != "str-1" {
		t.Errorf("stream id = %q", resp.StreamID)
	}

	if !m.Wait(2 * time.Second) {
		t.Fatal("stream did not finish")
	}

	l.mu.Lock()
	tokens := append([]string{}, l.tokens...)
	l.mu.Unlock()

	if len(tokens) != 1 || tokens[0] != "hi" {
		t.Errorf("tokens = %v", tokens)
	}
	if atomic.LoadInt32(&l.streamEnd) != 1 {
		t.Errorf("streamEnd = %d", l.streamEnd)
	}
	if m.State() != StateDone {
		t.Errorf("state = %s", m.State())
	}
}

func TestStreamManagerCancel(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/chat/start":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"stream_id":"str-2"}`)
		case "/api/chat/cancel":
			w.WriteHeader(http.StatusOK)
		case "/api/chat/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			for {
				select {
				case <-r.Context().Done():
					return
				default:
				}
				fmt.Fprint(w, "event: token\ndata: {\"content\":\"x\"}\n\n")
				w.(http.Flusher).Flush()
				time.Sleep(50 * time.Millisecond)
			}
		default:
			http.NotFound(w, r)
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	m := newTestManager(ts)
	l := &testListener{}

	if _, err := m.Start(&endpoints.StartChatRequest{SessionID: "ses-2", Message: "hello"}, l); err != nil {
		t.Fatalf("Start: %v", err)
	}

	time.Sleep(80 * time.Millisecond)
	if err := m.Cancel(); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	if !m.Wait(2 * time.Second) {
		t.Fatal("stream did not finish")
	}

	if m.State() != StateCancelled {
		t.Errorf("state = %s", m.State())
	}
}

func TestStreamManagerSteer(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/chat/start":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"stream_id":"str-3"}`)
		case "/api/chat/steer":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"accepted":true}`)
		case "/api/chat/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			for {
				select {
				case <-r.Context().Done():
					return
				default:
				}
				fmt.Fprint(w, "event: token\ndata: {\"content\":\"x\"}\n\n")
				w.(http.Flusher).Flush()
				time.Sleep(50 * time.Millisecond)
			}
		default:
			http.NotFound(w, r)
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	m := newTestManager(ts)
	l := &testListener{}

	if _, err := m.Start(&endpoints.StartChatRequest{SessionID: "ses-3", Message: "hello"}, l); err != nil {
		t.Fatalf("Start: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	resp, err := m.Steer("faster")
	if err != nil {
		t.Fatalf("Steer: %v", err)
	}
	if !resp.Accepted {
		t.Errorf("steer not accepted")
	}

	_ = m.Cancel()
	m.Wait(2 * time.Second)
}

func TestStreamManagerDoubleStart(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/chat/start":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"stream_id":"str-4"}`)
		case "/api/chat/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: stream_end\ndata: {}\n\n")
			w.(http.Flusher).Flush()
		default:
			http.NotFound(w, r)
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	m := newTestManager(ts)
	l := &testListener{}

	if _, err := m.Start(&endpoints.StartChatRequest{SessionID: "ses-4", Message: "hello"}, l); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if _, err := m.Start(&endpoints.StartChatRequest{SessionID: "ses-4", Message: "again"}, l); err == nil {
		t.Fatal("expected error on double start")
	}
}

// Compile-time check that Manager satisfies the expected surface.
var _ sse.EventListener = (&testListener{})
