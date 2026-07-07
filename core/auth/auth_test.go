package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	c, err := NewClient("https://hermes.example.com")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/status" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"auth_enabled":true,"logged_in":false}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	status, err := c.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !status.AuthEnabled {
		t.Error("expected auth_enabled=true")
	}
	if status.LoggedIn {
		t.Error("expected logged_in=false")
	}
}

func TestLogin(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/login" && r.Method == "POST" {
			// Set a session cookie
			http.SetCookie(w, &http.Cookie{
				Name:  "hermes_session",
				Value: "test-session-token",
				Path:  "/",
			})
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if err := c.Login("test-password"); err != nil {
		t.Fatalf("Login: %v", err)
	}

	if !c.IsAuthenticated() {
		t.Error("expected IsAuthenticated=true after login")
	}
}

func TestLogout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/login" && r.Method == "POST" {
			http.SetCookie(w, &http.Cookie{
				Name:  "hermes_session",
				Value: "test-session-token",
				Path:  "/",
			})
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/auth/logout" && r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if err := c.Login("test-password"); err != nil {
		t.Fatalf("Login: %v", err)
	}
	if err := c.Logout(); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if c.IsAuthenticated() {
		t.Error("expected IsAuthenticated=false after logout")
	}
}
