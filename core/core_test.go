package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHermeyClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c, err := NewHermeyClient(ts.URL)
	if err != nil {
		t.Fatalf("NewHermeyClient: %v", err)
	}
	if c.BaseURL() != ts.URL {
		t.Errorf("BaseURL = %q, want %q", c.BaseURL(), ts.URL)
	}
	if c.Version() != Version {
		t.Errorf("Version = %q, want %q", c.Version(), Version)
	}
}

func TestHermeyClientHealth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	c, err := NewHermeyClient(ts.URL)
	if err != nil {
		t.Fatalf("NewHermeyClient: %v", err)
	}
	if err := c.Health(); err != nil {
		t.Fatalf("Health: %v", err)
	}
}

func TestHermeyClientAuthStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/status" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"auth_enabled":true,"logged_in":false}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	c, err := NewHermeyClient(ts.URL)
	if err != nil {
		t.Fatalf("NewHermeyClient: %v", err)
	}
	status, err := c.AuthStatus()
	if err != nil {
		t.Fatalf("AuthStatus: %v", err)
	}
	if !status.AuthEnabled {
		t.Error("expected auth_enabled=true")
	}
	if status.LoggedIn {
		t.Error("expected logged_in=false")
	}
}
