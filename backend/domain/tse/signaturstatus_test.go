//go:build unit

package tse

import (
	"testing"
	"time"
)

var statusTestErstellt = time.Date(2026, 6, 10, 18, 0, 0, 0, time.UTC)

// signaturNach liefert eine quittierte Signatur, deren TSE-logTime um delay
// nach der Auftragserstellung liegt.
func signaturNach(delay time.Duration) *Signatur {
	return &Signatur{
		TransaktionNummer: 41,
		SignaturZaehler:   700,
		TSESeriennummer:   "TSE-SN-1",
		LogTimeStart:      statusTestErstellt.Add(delay),
		LogTimeEnd:        statusTestErstellt.Add(delay),
		Signatur:          "SIG-1",
		QRCodeData:        "V0;QR",
	}
}

func TestBestimmeSignaturstatus(t *testing.T) {
	aktiveRueckstandStoerung := &Stoerung{Beginn: statusTestErstellt, GrundArt: StoerungGrundRueckstand, Fehlertext: "Rueckstand"}
	aktiveTSEFehlerStoerung := &Stoerung{Beginn: statusTestErstellt, GrundArt: StoerungGrundTSEFehler, Fehlertext: "HTTP 503"}

	tests := []struct {
		name             string
		auftrag          SignaturauftragStand
		aktiveStoerung   *Stoerung
		wantStatus       Signaturstatus
		wantAusfallGrund string
		wantSignatur     bool
	}{
		{
			name:         "erledigt kurz nach Erstellung -> vorhanden",
			auftrag:      SignaturauftragStand{Status: StatusErledigt, ErstelltAm: statusTestErstellt, Signatur: signaturNach(3 * time.Second)},
			wantStatus:   SignaturstatusVorhanden,
			wantSignatur: true,
		},
		{
			name:         "erledigt genau an der Schwelle -> vorhanden (Kriterium: spaeter als)",
			auftrag:      SignaturauftragStand{Status: StatusErledigt, ErstelltAm: statusTestErstellt, Signatur: signaturNach(NachsigniertSchwelle)},
			wantStatus:   SignaturstatusVorhanden,
			wantSignatur: true,
		},
		{
			name:         "erledigt verspaetet -> nachsigniert",
			auftrag:      SignaturauftragStand{Status: StatusErledigt, ErstelltAm: statusTestErstellt, Signatur: signaturNach(5 * time.Minute)},
			wantStatus:   SignaturstatusNachsigniert,
			wantSignatur: true,
		},
		{
			name: "erledigt verspaetet trotz aktiver Stoerung -> nachsigniert (Aufholphase)",
			auftrag: SignaturauftragStand{
				Status: StatusErledigt, ErstelltAm: statusTestErstellt, Signatur: signaturNach(5 * time.Minute),
			},
			aktiveStoerung: aktiveRueckstandStoerung,
			wantStatus:     SignaturstatusNachsigniert,
			wantSignatur:   true,
		},
		{
			name:             "fehlgeschlagen -> Ausfall mit Endstatus als Grund",
			auftrag:          SignaturauftragStand{Status: StatusFehlgeschlagen, ErstelltAm: statusTestErstellt},
			wantStatus:       SignaturstatusAusfall,
			wantAusfallGrund: StatusFehlgeschlagen,
		},
		{
			name:             "tse_nicht_konfiguriert -> Ausfall",
			auftrag:          SignaturauftragStand{Status: StatusTSENichtKonfiguriert, ErstelltAm: statusTestErstellt},
			wantStatus:       SignaturstatusAusfall,
			wantAusfallGrund: StatusTSENichtKonfiguriert,
		},
		{
			// Fehlversuche unterhalb der Maximalzahl zaehlen nicht: Ein
			// Gift-Auftrag ist bis zum endgueltigen Fehlschlag ausstehend.
			name:       "offen ohne Stoerung -> ausstehend (bloße Latenz ist kein Ausfall)",
			auftrag:    SignaturauftragStand{Status: StatusOffen, ErstelltAm: statusTestErstellt},
			wantStatus: SignaturstatusAusstehend,
		},
		{
			// Geschlossene Zeitraeume zaehlen nicht: Der Aufrufer reicht nur
			// den aktiven Zeitraum herein; ohne aktiven bleibt es ausstehend.
			name:       "offen nach geschlossener Stoerung -> ausstehend",
			auftrag:    SignaturauftragStand{Status: StatusOffen, ErstelltAm: statusTestErstellt.Add(-10 * time.Minute)},
			wantStatus: SignaturstatusAusstehend,
		},
		{
			// Am Watchdog-Tick oeffnet die Schwellen-Ueberschreitung den
			// Rueckstands-Zeitraum — dasselbe offene Event kippt von
			// ausstehend in Ausfall.
			name:             "offen bei aktivem Rueckstands-Zeitraum -> Ausfall mit Grund-Art",
			auftrag:          SignaturauftragStand{Status: StatusOffen, ErstelltAm: statusTestErstellt},
			aktiveStoerung:   aktiveRueckstandStoerung,
			wantStatus:       SignaturstatusAusfall,
			wantAusfallGrund: StoerungGrundRueckstand,
		},
		{
			name:             "offen bei aktivem TSE-Fehler-Zeitraum -> Ausfall mit Grund-Art",
			auftrag:          SignaturauftragStand{Status: StatusOffen, ErstelltAm: statusTestErstellt},
			aktiveStoerung:   aktiveTSEFehlerStoerung,
			wantStatus:       SignaturstatusAusfall,
			wantAusfallGrund: StoerungGrundTSEFehler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BestimmeSignaturstatus(tt.auftrag, tt.aktiveStoerung)

			if got.Status != tt.wantStatus {
				t.Errorf("Status = %q, erwartet %q", got.Status, tt.wantStatus)
			}
			if got.AusfallGrund != tt.wantAusfallGrund {
				t.Errorf("AusfallGrund = %q, erwartet %q", got.AusfallGrund, tt.wantAusfallGrund)
			}
			if (got.Signatur != nil) != tt.wantSignatur {
				t.Errorf("Signatur gesetzt = %v, erwartet %v", got.Signatur != nil, tt.wantSignatur)
			}
			if tt.wantSignatur && got.Signatur != tt.auftrag.Signatur {
				t.Errorf("Signatur wird nicht unveraendert durchgereicht")
			}
		})
	}
}
