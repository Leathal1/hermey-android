// Package stream exposes the stream control state machine and the gomobile
// bridge entry point for Hermdroid chat streaming.
//
// The state machine owns stream lifecycle: start -> streaming -> (cancel |
// steer | branch | truncate) -> terminal.  All outbound events are delivered
// one at a time through the SSE EventListener callback so the Kotlin side stays
// single-threaded at the boundary.
package stream

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Leathal1/hermey-android/core/client"
	"github.com/Leathal1/hermey-android/core/endpoints"
	"github.com/Leathal1/hermey-android/core/models"
	"github.com/Leathal1/hermey-android/core/sse"
)

// State is the lifecycle state of a chat stream.
type State string

const (
	StateIdle           State = "idle"
	StateStarting       State = "starting"
	StateStreaming      State = "streaming"
	StateReconnecting   State = "reconnecting"
	StateCancelled      State = "cancelled"
	StateError          State = "error"
	StateDone           State = "done"
)

// Manager is the gomobile-facing stream state machine.
// It handles start/cancel/steer/branch/truncate and drives the SSE loop.
type Manager struct {
	api       *client.APIClient
	baseURL   string
	http      *http.Client
	mu        sync.Mutex
	state     State
	streamID  string
	sessionID string
	stream    *sse.Stream
	listener  sse.EventListener
}

// NewManager creates a stream state machine bound to an API client.
func NewManager(api *client.APIClient) *Manager {
	return &Manager{
		api:     api,
		baseURL: api.BaseURL(),
		http:    api.HTTPClient(),
		state:   StateIdle,
	}
}

// State returns the current stream state.
func (m *Manager) State() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

func (m *Manager) setState(s State) {
	m.mu.Lock()
	m.state = s
	m.mu.Unlock()
}

// transition moves to a terminal state only if the current state permits it.
func (m *Manager) transition(s State) {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch s {
	case StateDone:
		if m.state == StateStreaming || m.state == StateReconnecting || m.state == StateStarting {
			m.state = StateDone
		}
	case StateError:
		if m.state == StateStreaming || m.state == StateReconnecting || m.state == StateStarting {
			m.state = StateError
		}
	default:
		m.state = s
	}
}

// Start begins a new assistant turn and opens the SSE stream.
func (m *Manager) Start(req *endpoints.StartChatRequest, listener sse.EventListener) (*models.StreamStartResponse, error) {
	m.mu.Lock()
	if m.state != StateIdle && m.state != StateDone && m.state != StateError && m.state != StateCancelled {
		m.mu.Unlock()
		return nil, fmt.Errorf("stream: cannot start from state %s", m.state)
	}
	m.state = StateStarting
	m.listener = listener
	m.mu.Unlock()

	resp, err := endpoints.StartChat(m.api, req)
	if err != nil {
		m.transition(StateError)
		return nil, err
	}

	m.setState(StateStreaming)
	m.streamID = resp.StreamID
	m.sessionID = req.SessionID
	m.stream = sse.NewStream(m.baseURL, resp.StreamID, m.http)

	go func() {
		if err := m.stream.Subscribe(listener); err != nil {
			m.transition(StateError)
		} else {
			m.transition(StateDone)
		}
	}()

	return resp, nil
}

// Cancel requests the server to cancel the active stream and stops the client loop.
func (m *Manager) Cancel() error {
	m.mu.Lock()
	sid := m.streamID
	state := m.state
	stream := m.stream
	m.state = StateCancelled
	m.mu.Unlock()

	if state != StateStreaming && state != StateStarting && state != StateReconnecting {
		return fmt.Errorf("stream: cannot cancel from state %s", state)
	}

	if stream != nil {
		stream.Cancel()
	}

	if sid != "" {
		if err := endpoints.CancelChat(m.api, &endpoints.CancelChatRequest{StreamID: sid}); err != nil {
			return fmt.Errorf("stream: cancel failed: %w", err)
		}
	}

	return nil
}

// Steer sends a steering message to the active stream.
func (m *Manager) Steer(text string) (*models.SteerResponse, error) {
	m.mu.Lock()
	sid := m.sessionID
	state := m.state
	m.mu.Unlock()

	if state != StateStreaming && state != StateReconnecting {
		return nil, fmt.Errorf("stream: cannot steer from state %s", state)
	}

	if sid == "" {
		return nil, fmt.Errorf("stream: no session to steer")
	}

	return endpoints.SteerChat(m.api, &endpoints.SteerChatRequest{SessionID: sid, Text: text})
}

// Branch forks the current session at the current message boundary.
func (m *Manager) Branch(title string) (*endpoints.ForkSessionResponse, error) {
	m.mu.Lock()
	sid := m.sessionID
	m.mu.Unlock()

	if sid == "" {
		return nil, fmt.Errorf("stream: no session to branch")
	}

	return endpoints.ForkSession(m.api, &endpoints.ForkSessionRequest{SessionID: sid, Title: title})
}

// Truncate compacts the current session as the server-side truncate equivalent.
func (m *Manager) Truncate() error {
	m.mu.Lock()
	sid := m.sessionID
	m.mu.Unlock()

	if sid == "" {
		return fmt.Errorf("stream: no session to truncate")
	}

	return endpoints.CompactSession(m.api, &endpoints.CompactSessionRequest{SessionID: sid})
}

// StreamID returns the active stream id, if any.
func (m *Manager) StreamID() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.streamID
}

// SessionID returns the session id bound to the active stream.
func (m *Manager) SessionID() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessionID
}

// Wait blocks until the stream reaches a terminal state or the timeout expires.
func (m *Manager) Wait(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		s := m.State()
		if s == StateDone || s == StateError || s == StateCancelled {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}
