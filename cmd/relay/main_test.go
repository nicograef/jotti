package main

import (
	"strings"
	"testing"
)

func TestLoadConfigFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		wantConfig  RelayConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "uses defaults when optional values are missing",
			env: map[string]string{
				"RELAY_AUTH_TOKEN": "devrelay",
			},
			wantConfig: RelayConfig{
				Token:         "devrelay",
				BackendURL:    defaultBackendURL,
				PollSeconds:   defaultPollSeconds,
				TLSSkipVerify: true,
			},
		},
		{
			name: "uses overridden backend url and poll interval",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":   "devrelay",
				"RELAY_BACKEND_URL":  "https://example.org/api",
				"RELAY_POLL_SECONDS": "5",
			},
			wantConfig: RelayConfig{
				Token:         "devrelay",
				BackendURL:    "https://example.org/api",
				PollSeconds:   5,
				TLSSkipVerify: false,
			},
		},
		{
			name: "trims trailing slash from backend url",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":  "devrelay",
				"RELAY_BACKEND_URL": "https://example.org/api/",
			},
			wantConfig: RelayConfig{
				Token:         "devrelay",
				BackendURL:    "https://example.org/api",
				PollSeconds:   defaultPollSeconds,
				TLSSkipVerify: false,
			},
		},
		{
			name: "accepts explicit TLS skip true",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":       "devrelay",
				"RELAY_BACKEND_URL":      "https://example.org/api",
				"RELAY_TLS_SKIP_VERIFY": "true",
			},
			wantConfig: RelayConfig{
				Token:         "devrelay",
				BackendURL:    "https://example.org/api",
				PollSeconds:   defaultPollSeconds,
				TLSSkipVerify: true,
			},
		},
		{
			name: "accepts explicit TLS skip false on localhost",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":       "devrelay",
				"RELAY_TLS_SKIP_VERIFY": "0",
			},
			wantConfig: RelayConfig{
				Token:         "devrelay",
				BackendURL:    defaultBackendURL,
				PollSeconds:   defaultPollSeconds,
				TLSSkipVerify: false,
			},
		},
		{
			name: "fails on invalid TLS skip value",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":       "devrelay",
				"RELAY_TLS_SKIP_VERIFY": "ja",
			},
			wantErr:     true,
			errContains: "RELAY_TLS_SKIP_VERIFY",
		},
		{
			name: "fails when token is missing",
			env: map[string]string{
				"RELAY_BACKEND_URL": "https://example.org/api",
			},
			wantErr:     true,
			errContains: "RELAY_AUTH_TOKEN",
		},
		{
			name: "fails when poll is not a positive integer",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":   "devrelay",
				"RELAY_POLL_SECONDS": "0",
			},
			wantErr:     true,
			errContains: "RELAY_POLL_SECONDS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := loadConfigFromEnv(func(key string) string {
				return tt.env[key]
			})

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if config.Token != tt.wantConfig.Token {
				t.Fatalf("token mismatch: got %q, want %q", config.Token, tt.wantConfig.Token)
			}
			if config.BackendURL != tt.wantConfig.BackendURL {
				t.Fatalf("backend url mismatch: got %q, want %q", config.BackendURL, tt.wantConfig.BackendURL)
			}
			if config.PollSeconds != tt.wantConfig.PollSeconds {
				t.Fatalf("poll seconds mismatch: got %d, want %d", config.PollSeconds, tt.wantConfig.PollSeconds)
			}
			if config.TLSSkipVerify != tt.wantConfig.TLSSkipVerify {
				t.Fatalf("tls skip verify mismatch: got %t, want %t", config.TLSSkipVerify, tt.wantConfig.TLSSkipVerify)
			}
		})
	}
}
