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
		wantQR      bool
		wantGreen   bool
		wantRefresh bool
	}{
		{
			name:        "gültiges Zertifikat, Rebind ok → grüne Adresse mit QR, kein Refresh",
			in:          statusInputs{cert: certValid, rebindOK: true, greenURL: green, fallbackURL: fallback},
			wantPrimary: green, wantNotice: noticeGreen, wantQR: true, wantGreen: true, wantRefresh: false,
		},
		{
			name:        "kein Zertifikat → Fallback, Ausstellung läuft, Selbst-Refresh",
			in:          statusInputs{cert: certNone, rebindOK: true, greenURL: green, fallbackURL: fallback},
			wantPrimary: fallback, wantNotice: noticeIssuing, wantQR: false, wantGreen: false, wantRefresh: true,
		},
		{
			name:        "abgelaufenes Zertifikat → Fallback, Erneuerung läuft",
			in:          statusInputs{cert: certExpired, rebindOK: true, greenURL: green, fallbackURL: fallback},
			wantPrimary: fallback, wantNotice: noticeRenewing, wantQR: false, wantGreen: false, wantRefresh: true,
		},
		{
			name:        "Rebind blockiert trotz gültigem Zertifikat → Fallback + Anleitung",
			in:          statusInputs{cert: certValid, rebindOK: false, greenURL: green, fallbackURL: fallback},
			wantPrimary: fallback, wantNotice: noticeRebind, wantQR: false, wantGreen: false, wantRefresh: true,
		},
		{
			name:        "keine grüne Adresse möglich (kein State/keine IP) → Fallback",
			in:          statusInputs{greenURL: "", fallbackURL: fallback},
			wantPrimary: fallback, wantNotice: noticeNoGreen, wantQR: false, wantGreen: false, wantRefresh: true,
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
			if got.showQR != tt.wantQR {
				t.Errorf("showQR = %v, want %v", got.showQR, tt.wantQR)
			}
			if got.greenActive != tt.wantGreen {
				t.Errorf("greenActive = %v, want %v", got.greenActive, tt.wantGreen)
			}
			if got.refresh != tt.wantRefresh {
				t.Errorf("refresh = %v, want %v", got.refresh, tt.wantRefresh)
			}
		})
	}
}
