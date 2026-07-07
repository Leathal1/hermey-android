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
	"time"

	"github.com/Leathal1/hermey-android/core/models"
)

// EventListener receives SSE events from the chat stream.
// Implemented by the Kotlin side via gomobile callback interface.
type EventListener interface {
	OnToken(text string)
	OnToolCall(callJSON string)
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
}

// NewStream creates a new Stream for the given stream ID.
func NewStream(baseURL, streamID string, httpClient *http.Client) *Stream {
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
	s.listener = listener

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel
	defer cancel()

	url := fmt.Sprintf("%s/api/chat/stream?stream_id=%s", s.baseURL, s.streamID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		listener.OnError(fmt.Sprintf("create request: %v", err))
		return err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := s.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			listener.OnCancel()
			return nil
		}
		listener.OnError(fmt.Sprintf("connect: %v", err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		listener.OnError(fmt.Sprintf("HTTP %d", resp.StatusCode))
		return fmt.Errorf("sse: HTTP %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 1MB max line

	var eventType, data string

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			listener.OnCancel()
			return nil
		default:
		}

		line := scanner.Text()

		// Heartbeat comment
		if strings.HasPrefix(line, ":") {
			continue
		}

		// Empty line = dispatch event
		if line == "" {
			if data != "" {
				s.dispatch(eventType, data)
			}
			eventType = ""
			data = ""
			continue
		}

		// Parse field:value
		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(line[6:])
		} else if strings.HasPrefix(line, "data:") {
			data = strings.TrimSpace(line[5:])
		}
	}

	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			listener.OnCancel()
			return nil
		}
		listener.OnError(fmt.Sprintf("scanner: %v", err))
		return err
	}

	// Stream ended normally
	listener.OnStreamEnd()
	return nil
}

// Cancel stops the active stream.
func (s *Stream) Cancel() {
	if s.cancelFn != nil {
		s.cancelFn()
	}
}

// dispatch routes an SSE event to the appropriate listener method.
func (s *Stream) dispatch(eventType, data string) {
	switch eventType {
	case "token":
		var event models.ChatEvent
		if err := json.Unmarshal([]byte(data), &event); err == nil {
			s.listener.OnToken(event.Content)
		}
	case "tool_call":
		s.listener.OnToolCall(data)
	case "tool_result":
		s.listener.OnToolCall(data)
	case "reasoning":
		var event models.ChatEvent
		if err := json.Unmarshal([]byte(data), &event); err == nil {
			s.listener.OnReasoning(event.Content)
		}
	case "stream_end":
		s.listener.OnStreamEnd()
	case "error":
		var event models.ChatEvent
		if err := json.Unmarshal([]byte(data), &event); err == nil {
			s.listener.OnError(event.Error)
		} else {
			s.listener.OnError(data)
		}
	case "cancel":
		s.listener.OnCancel()
	default:
		// Unknown event type — ignore gracefully
	}
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

// unused import guard
var _ = time.Now
