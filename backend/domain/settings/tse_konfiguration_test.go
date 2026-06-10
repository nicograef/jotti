//go:build unit

package settings

import "testing"

func TestNewTSEKonfiguration(t *testing.T) {
	conf, err := NewTSEKonfiguration(" api-key ", " api-secret ", " tss-1 ", " client-1 ")
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

func TestTSEKonfigurationIstKonfiguriert(t *testing.T) {
	tests := []struct {
		name string
		conf TSEKonfiguration
		want bool
	}{
		{
			name: "all fields set",
			conf: TSEKonfiguration{ApiKey: "a", ApiSecret: "b", TssID: "c", ClientID: "d"},
			want: true,
		},
		{
			name: "api key missing",
			conf: TSEKonfiguration{ApiSecret: "b", TssID: "c", ClientID: "d"},
			want: false,
		},
		{
			name: "whitespace is treated as missing",
			conf: TSEKonfiguration{ApiKey: " ", ApiSecret: "b", TssID: "c", ClientID: "d"},
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

func TestNewTSEKonfiguration_PartialValuesRejected(t *testing.T) {
	_, err := NewTSEKonfiguration("api-key", "", "tss-1", "client-1")
	if err == nil {
		t.Fatal("expected error for partial config, got nil")
	}
}
