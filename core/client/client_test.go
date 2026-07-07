package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Leathal1/hermey-android/core/auth"
)

func newTestAPIClient(t *testing.T, handler http.HandlerFunc) (*APIClient, *httptest.Server) {
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	authClient, err := auth.NewClient(srv.URL)
	if err != nil {
		t.Fatalf("auth.NewClient: %v", err)
	}
	c := NewAPIClient(authClient)
	c.SetBaseURL(srv.URL)
	return c, srv
}

func TestHealth(t *testing.T) {
	called := false
	api, _ := newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != "/health" {
			t.Errorf("path = %q, want /health", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	if err := api.Health(); err != nil {
		t.Fatalf("Health: %v", err)
	}
	if !called {
		t.Error("handler not called")
	}
}

func TestListSessions(t *testing.T) {
	api, _ := newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sessions" {
			t.Errorf("path = %q, want /api/sessions", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sessions": []map[string]string{
				{"id": "s1", "title": "First"},
				{"id": "s2", "title": "Second"},
			},
		})
	})
	resp, err := api.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(resp.Sessions) != 2 {
		t.Fatalf("sessions = %d, want 2", len(resp.Sessions))
	}
	if resp.Sessions[0].ID != "s1" {
		t.Errorf("sessions[0].ID = %q, want s1", resp.Sessions[0].ID)
	}
}

func TestListModels(t *testing.T) {
	api, _ := newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/models" {
			t.Errorf("path = %q, want /api/models", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]string{
			{"id": "gpt-4", "name": "GPT-4", "provider": "openai"},
		})
	})
	models, err := api.ListModels()
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("models = %d, want 1", len(models))
	}
	if models[0].Provider != "openai" {
		t.Errorf("models[0].Provider = %q, want openai", models[0].Provider)
	}
}

func TestGetSettings(t *testing.T) {
	api, _ := newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/settings" {
			t.Errorf("path = %q, want /api/settings", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"bot_name": "Hermes",
			"version": "1.0",
		})
	})
	settings, err := api.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if settings.BotName != "Hermes" {
		t.Errorf("BotName = %q, want Hermes", settings.BotName)
	}
	if settings.Version != "1.0" {
		t.Errorf("Version = %q, want 1.0", settings.Version)
	}
}

func TestNewSession(t *testing.T) {
	api, _ := newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/session/new" {
			t.Errorf("path = %q, want /api/session/new", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"id":    "new1",
			"title": "New Session",
		})
	})
	session, err := api.NewSession("ws1", "gpt-4", "openai", "default")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if session.ID != "new1" {
		t.Errorf("session.ID = %q, want new1", session.ID)
	}
}

func TestHealthError(t *testing.T) {
	api, _ := newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	err := api.Health()
	if err == nil {
		t.Error("Health() err = nil, want error on 500")
	}
}
