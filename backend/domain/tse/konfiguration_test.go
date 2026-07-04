//go:build unit

package tse

import "testing"

func TestNewKonfiguration(t *testing.T) {
	conf, err := NewKonfiguration(" api-key ", " api-secret ", " tss-1 ", " client-1 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conf.ApiKey != "api-key" {
		t.Fatalf("expected trimmed api key, got %q", conf.ApiKey)
	}
	if conf.ApiSecret != "api-secret" {
		t.Fatalf("expected trimmed api secret, got %q", conf.ApiSecret)
	}
	if conf.TssID != "tss-1" {
		t.Fatalf("expected trimmed tss id, got %q", conf.TssID)
	}
	if conf.ClientID != "client-1" {
		t.Fatalf("expected trimmed client id, got %q", conf.ClientID)
	}
	if conf.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be set")
	}
}

func TestKonfigurationIstKonfiguriert(t *testing.T) {
	tests := []struct {
		name string
		conf Konfiguration
		want bool
	}{
		{
			name: "all fields set",
			conf: Konfiguration{ApiKey: "a", ApiSecret: "b", TssID: "c", ClientID: "d"},
			want: true,
		},
		{
			name: "api key missing",
			conf: Konfiguration{ApiSecret: "b", TssID: "c", ClientID: "d"},
			want: false,
		},
		{
			name: "whitespace is treated as missing",
			conf: Konfiguration{ApiKey: " ", ApiSecret: "b", TssID: "c", ClientID: "d"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.conf.IstKonfiguriert()
			if got != tt.want {
				t.Fatalf("IstKonfiguriert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewKonfiguration_PartialValuesRejected(t *testing.T) {
	_, err := NewKonfiguration("api-key", "", "tss-1", "client-1")
	if err == nil {
		t.Fatal("expected error for partial config, got nil")
	}
}
