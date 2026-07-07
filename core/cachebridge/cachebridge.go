// Package cachebridge exposes the bbolt cache to Android via gomobile bind
// or to desktop JVM via JNI through a C-shared library.
package cachebridge

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Leathal1/hermey-android/core/cache"
)

var (
	mu      sync.Mutex
	handles = make(map[int]*cache.Cache)
	next    int
)

// CacheHandle wraps an integer handle so gomobile can pass it back to us.
type CacheHandle struct {
	Value int
}

// Open creates or opens a cache at path and returns a handle.
func Open(path string, maxMessages int, maxBytes int64) (*CacheHandle, error) {
	c, err := cache.Open(path, cache.Config{MaxMessages: maxMessages, MaxBytes: maxBytes})
	if err != nil {
		return nil, err
	}
	mu.Lock()
	h := next
	next++
	handles[h] = c
	mu.Unlock()
	return &CacheHandle{Value: h}, nil
}

// Close closes a cache handle.
func Close(h *CacheHandle) error {
	if h == nil {
		return fmt.Errorf("nil handle")
	}
	mu.Lock()
	c, ok := handles[h.Value]
	delete(handles, h.Value)
	mu.Unlock()
	if !ok {
		return fmt.Errorf("invalid handle %d", h.Value)
	}
	return c.Close()
}

// PutSession stores a session.
func PutSession(h *CacheHandle, id, title string, lastMessageAtMs int64, count int) error {
	c, err := lookup(h)
	if err != nil {
		return err
	}
	return c.PutSession(cache.Session{
		ID:            id,
		Title:         title,
		LastMessageAt: time.UnixMilli(lastMessageAtMs),
		MessageCount:  count,
	})
}

// ListSessionsJson returns sessions as a JSON array sorted by recency.
func ListSessionsJson(h *CacheHandle) (string, error) {
	c, err := lookup(h)
	if err != nil {
		return "", err
	}
	sessions, err := c.ListSessions()
	if err != nil {
		return "", err
	}
	out := make([]map[string]interface{}, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, map[string]interface{}{
			"id":            s.ID,
			"title":         s.Title,
			"lastMessageAt": s.LastMessageAt.UnixMilli(),
			"messageCount":  s.MessageCount,
		})
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}

// GetMessagesJson returns messages for a session as a JSON array.
func GetMessagesJson(h *CacheHandle, sessionID string) (string, error) {
	c, err := lookup(h)
	if err != nil {
		return "", err
	}
	msgs, err := c.GetMessagesBySession(sessionID)
	if err != nil {
		return "", err
	}
	out := make([]map[string]interface{}, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, map[string]interface{}{
			"id":        m.ID,
			"sessionId": m.SessionID,
			"role":      m.Role,
			"content":   m.Content,
			"timestamp": m.Timestamp.UnixMilli(),
		})
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}

// PutMessage stores a message and updates its session metadata.
func PutMessage(h *CacheHandle, id, sessionID, role, content string, timestampMs int64) error {
	c, err := lookup(h)
	if err != nil {
		return err
	}
	return c.PutMessage(cache.Message{
		ID:        id,
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		Timestamp: time.UnixMilli(timestampMs),
	})
}

// DeleteSession removes a session and its messages.
func DeleteSession(h *CacheHandle, sessionID string) error {
	c, err := lookup(h)
	if err != nil {
		return err
	}
	return c.DeleteSessionAndMessages(sessionID)
}

// Evict removes oldest sessions until limits are satisfied.
func Evict(h *CacheHandle) error {
	c, err := lookup(h)
	if err != nil {
		return err
	}
	return c.Evict()
}

func lookup(h *CacheHandle) (*cache.Cache, error) {
	if h == nil {
		return nil, fmt.Errorf("nil handle")
	}
	mu.Lock()
	c, ok := handles[h.Value]
	mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("invalid handle %d", h.Value)
	}
	return c, nil
}
