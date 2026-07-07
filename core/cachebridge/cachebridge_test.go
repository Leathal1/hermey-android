package cachebridge

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBridgeRoundTrip(t *testing.T) {
	dir := t.TempDir()
	h, err := Open(filepath.Join(dir, "bridge.db"), 100, 1_000_000)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = Close(h) }()

	now := time.Now().UnixMilli()
	if err := PutSession(h, "s1", "Test", now, 0); err != nil {
		t.Fatalf("put session: %v", err)
	}
	if err := PutMessage(h, "m1", "s1", "user", "hello", now); err != nil {
		t.Fatalf("put message: %v", err)
	}
	json, err := ListSessionsJson(h)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	t.Logf("sessions=%s", json)
	if !strings.Contains(json, "s1") {
		t.Fatal("missing s1")
	}
}
