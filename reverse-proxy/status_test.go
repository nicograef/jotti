package main

import "testing"

func TestDecideStatus(t *testing.T) {
	const green = "https://192-168-1-50.sub-1.lokal.jotti.rocks"
	const fallback = "https://192.168.1.50"

	tests := []struct {
		name        string
		in          statusInputs
		wantPrimary string
		wantNotice  notice
		wantRefresh bool
	}{
		{
			name:        "gültiges Zertifikat, Rebind ok → grüne Adresse mit QR, kein Refresh",
			in:          statusInputs{cert: certValid, rebindOK: true, greenURL: green, fallbackURL: fallback},
			wantPrimary: green, wantNotice: noticeGreen, wantRefresh: false,
		},
		{
			name:        "kein Zertifikat → Fallback, Ausstellung läuft, Selbst-Refresh",
			in:          statusInputs{cert: certNone, rebindOK: true, greenURL: green, fallbackURL: fallback},
			wantPrimary: fallback, wantNotice: noticeIssuing, wantRefresh: true,
		},
		{
			name:        "abgelaufenes Zertifikat → Fallback, Erneuerung läuft",
			in:          statusInputs{cert: certExpired, rebindOK: true, greenURL: green, fallbackURL: fallback},
			wantPrimary: fallback, wantNotice: noticeRenewing, wantRefresh: true,
		},
		{
			name:        "Rebind blockiert trotz gültigem Zertifikat → Fallback + Anleitung",
			in:          statusInputs{cert: certValid, rebindOK: false, greenURL: green, fallbackURL: fallback},
			wantPrimary: fallback, wantNotice: noticeRebind, wantRefresh: true,
		},
		{
			name:        "keine grüne Adresse möglich (kein State/keine IP) → Fallback ohne Refresh",
			in:          statusInputs{greenURL: "", fallbackURL: fallback},
			wantPrimary: fallback, wantNotice: noticeNoGreen, wantRefresh: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decideStatus(tt.in)
			if got.primaryURL != tt.wantPrimary {
				t.Errorf("primaryURL = %q, want %q", got.primaryURL, tt.wantPrimary)
			}
			if got.notice != tt.wantNotice {
				t.Errorf("notice = %d, want %d", got.notice, tt.wantNotice)
			}
			if got.refresh != tt.wantRefresh {
				t.Errorf("refresh = %v, want %v", got.refresh, tt.wantRefresh)
			}
		})
	}
}
