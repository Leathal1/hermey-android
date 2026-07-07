package core

import "testing"

func TestNewHermeyClient(t *testing.T) {
	c := NewHermeyClient("https://hermes.example.com")
	if c.BaseURL() != "https://hermes.example.com" {
		t.Errorf("BaseURL = %q, want %q", c.BaseURL(), "https://hermes.example.com")
	}
	if c.Version() != Version {
		t.Errorf("Version = %q, want %q", c.Version(), Version)
	}
}
