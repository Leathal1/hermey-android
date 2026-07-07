package sse

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testListener struct {
	mu          sync.Mutex
	tokens      []string
	toolCalls   []string
	toolResults []string
	reasoning   []string
	streamEnd   int32
	errors      []string
	cancelled   int32
}

func (l *testListener) OnToken(text string)     { l.mu.Lock(); l.tokens = append(l.tokens, text); l.mu.Unlock() }
func (l *testListener) OnToolCall(callJSON string) { l.mu.Lock(); l.toolCalls = append(l.toolCalls, callJSON); l.mu.Unlock() }
func (l *testListener) OnToolResult(resultJSON string) { l.mu.Lock(); l.toolResults = append(l.toolResults, resultJSON); l.mu.Unlock() }
func (l *testListener) OnReasoning(text string) { l.mu.Lock(); l.reasoning = append(l.reasoning, text); l.mu.Unlock() }
func (l *testListener) OnStreamEnd()            { atomic.AddInt32(&l.streamEnd, 1) }
func (l *testListener) OnError(msg string)      { l.mu.Lock(); l.errors = append(l.errors, msg); l.mu.Unlock() }
func (l *testListener) OnCancel()                 { atomic.AddInt32(&l.cancelled, 1) }

func TestSSEAllEventTypes(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("no flusher")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, ":heartbeat\n\n")
		fmt.Fprint(w, "event: token\ndata: {\"content\":\"hello\"}\n\n")
		fmt.Fprint(w, "event: reasoning\ndata: {\"content\":\"thinking\"}\n\n")
		fmt.Fprint(w, "event: tool_call\ndata: {\"id\":\"1\",\"name\":\"test\"}\n\n")
		fmt.Fprint(w, "event: tool_result\ndata: {\"id\":\"1\",\"result\":\"ok\"}\n\n")
		fmt.Fprint(w, "event: stream_end\ndata: {}\n\n")
		flusher.Flush()
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	stream := NewStream(ts.URL, "s1", ts.Client())
	l := &testListener{}
	if err := stream.Subscribe(l); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	if got, want := len(l.tokens), 1; got != want {
		t.Fatalf("tokens len = %d, want %d", got, want)
	}
	if l.tokens[0] != "hello" {
		t.Errorf("token = %q, want hello", l.tokens[0])
	}
	if len(l.reasoning) != 1 || l.reasoning[0] != "thinking" {
		t.Errorf("reasoning = %v", l.reasoning)
	}
	if len(l.toolCalls) != 1 {
		t.Errorf("toolCalls len = %d", len(l.toolCalls))
	}
	if len(l.toolResults) != 1 {
		t.Errorf("toolResults len = %d", len(l.toolResults))
	}
	if atomic.LoadInt32(&l.streamEnd) != 1 {
		t.Errorf("streamEnd = %d", l.streamEnd)
	}
}

func TestSSECancelFromClient(t *testing.T) {
	done := make(chan struct{})
	handler := func(w http.ResponseWriter, r *http.Request) {
		defer close(done)
		flusher, _ := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		for i := 0; i < 100; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
			}
			fmt.Fprintf(w, "event: token\ndata: {\"content\":\"%d\"}\n\n", i)
			flusher.Flush()
			time.Sleep(20 * time.Millisecond)
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	stream := NewStream(ts.URL, "s1", ts.Client())
	l := &testListener{}

	go func() {
		time.Sleep(80 * time.Millisecond)
		stream.Cancel()
	}()

	stream.Subscribe(l)
	<-done

	if atomic.LoadInt32(&l.cancelled) != 1 {
		t.Errorf("cancelled = %d", l.cancelled)
	}
}

func TestSSEReconnect(t *testing.T) {
	attempts := 0
	var mu sync.Mutex
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat/stream/status" {
			w.WriteHeader(http.StatusOK)
			return
		}
		mu.Lock()
		attempts++
		cur := attempts
		mu.Unlock()

		flusher, _ := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		if cur == 1 {
			// Force a disconnect after one token.
			fmt.Fprint(w, "event: token\ndata: {\"content\":\"first\"}\n\n")
			flusher.Flush()
			conn, _, _ := w.(http.Hijacker).Hijack()
			if conn != nil {
				conn.Close()
			}
			return
		}
		fmt.Fprint(w, "event: token\ndata: {\"content\":\"second\"}\n\n")
		fmt.Fprint(w, "event: stream_end\ndata: {}\n\n")
		flusher.Flush()
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	stream := NewStream(ts.URL, "s1", ts.Client())
	l := &testListener{}

	// Because our reconnect uses fixed exponential backoff, cap the test time
	// by overriding the delay in this test.  We shorten via a small wrapper
	// not exposed; instead just run with the real one-second delay.
	go func() {
		stream.Subscribe(l)
	}()

	time.Sleep(2500 * time.Millisecond)

	l.mu.Lock()
	tokens := append([]string{}, l.tokens...)
	l.mu.Unlock()

	if len(tokens) < 2 {
		t.Fatalf("expected 2 tokens, got %v", tokens)
	}
	if tokens[0] != "first" {
		t.Errorf("first token = %q", tokens[0])
	}
	if tokens[len(tokens)-1] != "second" {
		t.Errorf("last token = %q", tokens[len(tokens)-1])
	}
}

func TestSSEErrorEvent(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: error\ndata: {\"error\":\"boom\"}\n\n")
		w.(http.Flusher).Flush()
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	stream := NewStream(ts.URL, "s1", ts.Client())
	l := &testListener{}
	stream.Subscribe(l)

	l.mu.Lock()
	errs := append([]string{}, l.errors...)
	l.mu.Unlock()
	if len(errs) != 1 || errs[0] != "boom" {
		t.Errorf("errors = %v", errs)
	}
}
