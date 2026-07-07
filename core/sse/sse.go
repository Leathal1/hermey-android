// Package sse provides a Server-Sent Events client for the hermes-webui chat stream.
//
// SSE event types handled:
//   - token: append text to current assistant message
//   - tool_call: render collapsible tool-call card
//   - tool_result: attach to last tool call
//   - reasoning: collapsible "thinking" block
//   - stream_end: finalize message, close connection
//   - error: show inline error, close connection
//   - cancel: user cancelled, close connection
//
// Heartbeat lines (starting with ':') are ignored.
// Cloudflare Tunnel idle timeout ~100s; heartbeats every 30s keep it alive.
package sse

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Leathal1/hermey-android/core/models"
)

// EventListener receives SSE events from the chat stream.
// Implemented by the Kotlin side via gomobile callback interface.
type EventListener interface {
	OnToken(text string)
	OnToolCall(callJSON string)
	OnToolResult(resultJSON string)
	OnReasoning(text string)
	OnStreamEnd()
	OnError(msg string)
	OnCancel()
}

// Stream manages an active SSE connection to a chat stream.
type Stream struct {
	streamID  string
	baseURL   string
	client    *http.Client
	listener  EventListener
	cancelFn  context.CancelFunc
	mu        sync.Mutex
	running   bool
	cancelled int32
}

// NewStream creates a new Stream for the given stream ID.
func NewStream(baseURL, streamID string, httpClient *http.Client) *Stream {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 0}
	}
	return &Stream{
		baseURL:  baseURL,
		streamID: streamID,
		client:   httpClient,
	}
}

// Subscribe connects to the SSE stream and begins delivering events to the listener.
// This method blocks until the stream ends or is cancelled.
// Call from a goroutine in the gomobile bridge.
func (s *Stream) Subscribe(listener EventListener) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("sse: stream %s already subscribed", s.streamID)
	}
	s.listener = listener
	s.running = true
	s.cancelled = 0
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel
	defer cancel()

	for attempt := 0; ; attempt++ {
		select {
		case <-ctx.Done():
			s.emitCancel()
			return nil
		default:
		}

		if attempt > 0 {
			wait := reconnectDelay(attempt)
			time.Sleep(wait)

			active, err := s.statusCheck()
			if err != nil {
				listener.OnError(fmt.Sprintf("reconnect status check failed: %v", err))
				return err
			}
			if !active {
				listener.OnError("stream no longer active on server")
				return fmt.Errorf("sse: stream %s is no longer active", s.streamID)
			}
		}

		terminal, err := s.connect(ctx)
		if err == nil {
			if terminal {
				return nil
			}
			continue
		}

		if ctx.Err() != nil {
			s.emitCancel()
			return nil
		}

		if !isRetryable(err) {
			return err
		}

		s.listener.OnError(fmt.Sprintf("stream disconnected (attempt %d): %v", attempt+1, err))
	}
}

func (s *Stream) emitCancel() {
	if atomic.CompareAndSwapInt32(&s.cancelled, 0, 1) {
		s.listener.OnCancel()
	}
}

// connect establishes one SSE connection and consumes it until completion.
// The returned bool is true when a terminal event was dispatched.
func (s *Stream) connect(ctx context.Context) (bool, error) {
	url := fmt.Sprintf("%s/api/chat/stream?stream_id=%s", s.baseURL, s.streamID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.listener.OnError(fmt.Sprintf("create request: %v", err))
		return false, err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := s.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			s.emitCancel()
			return false, nil
		}
		return false, fmt.Errorf("%w: connect: %v", errRetryable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.listener.OnError(fmt.Sprintf("HTTP %d", resp.StatusCode))
		return false, fmt.Errorf("sse: HTTP %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var eventType, data string

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			s.emitCancel()
			return false, nil
		default:
		}

		line := scanner.Text()

		if strings.HasPrefix(line, ":") {
			continue
		}

		if line == "" {
			if data != "" {
				term, err := s.dispatch(eventType, data)
				if err != nil {
					return false, err
				}
				if term {
					return true, nil
				}
			}
			eventType = ""
			data = ""
			continue
		}

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(line[6:])
		} else if strings.HasPrefix(line, "data:") {
			data = strings.TrimSpace(line[5:])
		}
	}

	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			s.emitCancel()
			return false, nil
		}
		return false, fmt.Errorf("%w: scanner: %v", errRetryable, err)
	}

	s.listener.OnStreamEnd()
	return true, nil
}

var errRetryable = fmt.Errorf("retryable stream error")

func isRetryable(err error) bool {
	return err != nil && strings.Contains(err.Error(), errRetryable.Error())
}

func reconnectDelay(attempt int) time.Duration {
	d := time.Duration(1<<uint(attempt-1)) * time.Second
	if d > 30*time.Second {
		d = 30 * time.Second
	}
	return d
}

func (s *Stream) statusCheck() (bool, error) {
	return StreamStatus(s.baseURL, s.streamID, s.client)
}

// Cancel stops the active stream.
func (s *Stream) Cancel() {
	s.mu.Lock()
	fn := s.cancelFn
	s.mu.Unlock()
	if fn != nil {
		fn()
	}
}

// dispatch routes an SSE event to the appropriate listener method.
// Returns true for terminal events (stream_end, error, cancel).
func (s *Stream) dispatch(eventType, data string) (bool, error) {
	switch eventType {
	case "token":
		var event models.ChatEvent
		if err := json.Unmarshal([]byte(data), &event); err == nil {
			s.listener.OnToken(event.Content)
		}
	case "tool_call":
		s.listener.OnToolCall(data)
	case "tool_result":
		s.listener.OnToolResult(data)
	case "reasoning":
		var event models.ChatEvent
		if err := json.Unmarshal([]byte(data), &event); err == nil {
			s.listener.OnReasoning(event.Content)
		}
	case "stream_end":
		s.listener.OnStreamEnd()
		return true, nil
	case "error":
		var event models.ChatEvent
		if err := json.Unmarshal([]byte(data), &event); err == nil {
			s.listener.OnError(event.Error)
		} else {
			s.listener.OnError(data)
		}
		return true, nil
	case "cancel":
		s.emitCancel()
		return true, nil
	default:
	}
	return false, nil
}

// StreamStatus checks if a stream is still active (for reconnect).
func StreamStatus(baseURL, streamID string, httpClient *http.Client) (bool, error) {
	url := fmt.Sprintf("%s/api/chat/stream/status?stream_id=%s", baseURL, streamID)
	resp, err := httpClient.Get(url)
	if err != nil {
		return false, fmt.Errorf("sse: status check: %w", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}
