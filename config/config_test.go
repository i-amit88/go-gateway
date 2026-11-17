package config

import (
	"testing"
	"time"
)

func TestLoadConfigUsesPortFromEnvironment(t *testing.T) {
	t.Setenv("PORT", "9090")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "9090" {
		t.Fatalf("expected port 9090, got %q", cfg.Port)
	}
}

func TestLoadConfigUsesDefaultPort(t *testing.T) {
	t.Setenv("PORT", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %q", cfg.Port)
	}
}

func TestLoadConfigRejectsInvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port string
	}{
		{name: "not a number", port: "abc"},
		{name: "below minimum", port: "0"},
		{name: "above maximum", port: "70000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PORT", tt.port)

			_, err := LoadConfig()
			if err == nil {
				t.Fatalf("expected error for PORT=%q", tt.port)
			}
		})
	}
}

func TestLoadConfigUsesDefaultProxyTimeout(t *testing.T) {
	t.Setenv("PROXY_TIMEOUT", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ProxyTimeout != 2*time.Second {
		t.Fatalf("expected default timeout 2s, got %v", cfg.ProxyTimeout)
	}
}

func TestLoadConfigUsesCustomProxyTimeout(t *testing.T) {
	t.Setenv("PROXY_TIMEOUT", "500ms")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ProxyTimeout != 500*time.Millisecond {
		t.Fatalf("expected timeout 500ms, got %v", cfg.ProxyTimeout)
	}
}

func TestLoadConfigRejectsInvalidProxyTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout string
	}{
		{name: "not a duration", timeout: "abc"},
		{name: "zero", timeout: "0s"},
		{name: "negative", timeout: "-1s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PROXY_TIMEOUT", tt.timeout)

			_, err := LoadConfig()
			if err == nil {
				t.Fatalf("expected error for PROXY_TIMEOUT=%q", tt.timeout)
			}
		})
	}
}
