package cache

import (
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

	for i := 0; i < 6; i++ {
		sid := "s%d"
		_ = c.PutSession(Session{ID: sid, Title: "Chat", LastMessageAt: time.Now().Add(time.Duration(i) * time.Hour)})
		m := Message{ID: "m%d", SessionID: sid, Role: "user", Content: "x", Timestamp: time.Now().Add(time.Duration(i) * time.Hour)}
		_ = c.PutMessage(m)
	}

	if err := c.Evict(); err != nil {
		t.Fatalf("evict: %v", err)
	}

	sessions, _ := c.ListSessions()
	if len(sessions) > 5 {
		t.Errorf("sessions = %d, want <= 5", len(sessions))
	}
}
