package cache

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func newTestCache(t *testing.T) (*Cache, func()) {
	dir := t.TempDir()
	c, err := Open(filepath.Join(dir, "cache.db"), Config{MaxMessages: 5, MaxBytes: 1_000_000})
	if err != nil {
		t.Fatalf("open cache: %v", err)
	}
	return c, func() {
		_ = c.Close()
	}
}

func TestPutAndGetSession(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()

	s := Session{ID: "s1", Title: "Test", LastMessageAt: time.Now()}
	if err := c.PutSession(s); err != nil {
		t.Fatalf("put session: %v", err)
	}

	got, err := c.GetSession("s1")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if got.Title != "Test" {
		t.Errorf("title = %q, want Test", got.Title)
	}
}

func TestPutMessageUpdatesSession(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()

	s := Session{ID: "s1", Title: "Chat", LastMessageAt: time.Now()}
	_ = c.PutSession(s)

	m := Message{ID: "m1", SessionID: "s1", Role: "user", Content: "hi", Timestamp: time.Now()}
	if err := c.PutMessage(m); err != nil {
		t.Fatalf("put message: %v", err)
	}

	got, _ := c.GetSession("s1")
	if got.MessageCount != 1 {
		t.Errorf("message count = %d, want 1", got.MessageCount)
	}
	if got.LastMessageAt.IsZero() {
		t.Error("expected last message time updated")
	}

	msgs, _ := c.GetMessagesBySession("s1")
	if len(msgs) != 1 || msgs[0].Content != "hi" {
		t.Errorf("messages = %v, want 1 hi", msgs)
	}
}

func TestEviction(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()

	// Insert 6 distinct sessions, each with a message. MaxMessages=5, so
	// eviction should drop at least one session (the oldest by LastMessageAt).
	for i := 0; i < 6; i++ {
		sid := fmt.Sprintf("s%d", i)
		ts := time.Now().Add(time.Duration(i) * time.Hour)
		if err := c.PutSession(Session{ID: sid, Title: "Chat", LastMessageAt: ts}); err != nil {
			t.Fatalf("put session %s: %v", sid, err)
		}
		if err := c.PutMessage(Message{
			ID:        fmt.Sprintf("m%d", i),
			SessionID: sid,
			Role:      "user",
			Content:   "x",
			Timestamp: ts,
		}); err != nil {
			t.Fatalf("put message %s: %v", sid, err)
		}
	}

	if err := c.Evict(); err != nil {
		t.Fatalf("evict: %v", err)
	}

	sessions, err := c.ListSessions()
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) > 5 {
		t.Errorf("sessions = %d, want <= 5 after eviction", len(sessions))
	}
	if len(sessions) >= 6 {
		t.Errorf("eviction did not remove any session; still %d", len(sessions))
	}

	// Verify at least one of the original 6 was evicted.
	surviving := make(map[string]bool)
	for _, s := range sessions {
		surviving[s.ID] = true
	}
	evicted := 0
	for i := 0; i < 6; i++ {
		if !surviving[fmt.Sprintf("s%d", i)] {
			evicted++
		}
	}
	if evicted == 0 {
		t.Error("expected at least one session evicted, got 0")
	}
}
