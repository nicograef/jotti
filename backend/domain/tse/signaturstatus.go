package tse

import "time"

// Zeitschwellen des Signaturstatus und des Störungsprotokolls — die einzige
// Stelle dieser Konstanten; Signaturstatus-Funktion, Rückstands-Watchdog und
// Dashboard-Warnung referenzieren sie.
const (
	// NachsigniertSchwelle: Eine Signatur, die später als diese Spanne nach
	// der Auftragserstellung entsteht, trägt das Nachsigniert-Kennzeichen —
	// ihre TSE-Zeitpunkte weichen sichtbar vom Belegdatum ab.
	NachsigniertSchwelle = time.Minute
	// RueckstandSchwelle: Ab diesem Alter des ältesten offenen Auftrags
	// dokumentiert der Rückstands-Watchdog einen Rückstands-Störungszeitraum.
	RueckstandSchwelle = 2 * time.Minute
	// WatchdogTickIntervall ist der Prüf-Takt des Rückstands-Watchdogs; die
	// Rückstands-Schwelle materialisiert nur am Tick.
	WatchdogTickIntervall = 10 * time.Second
)

// Signaturstatus ist das Urteil der Signaturstatus-Funktion über einen
// Signaturauftrag — genau eine von vier Ergebnisarten.
type Signaturstatus string

const (
	// SignaturstatusVorhanden: Die Signatur liegt vor.
	SignaturstatusVorhanden Signaturstatus = "vorhanden"
	// SignaturstatusNachsigniert: Die Signatur liegt vor, entstand aber
	// verspätet (später als NachsigniertSchwelle nach der Auftragserstellung);
	// der Beleg trägt das Nachsigniert-Kennzeichen.
	SignaturstatusNachsigniert Signaturstatus = "nachsigniert"
	// SignaturstatusAusfall: Keine Signatur, mit belegbarem Grund — Endstatus
	// des Auftrags oder offener Auftrag bei aktivem Störungszeitraum.
	SignaturstatusAusfall Signaturstatus = "ausfall"
	// SignaturstatusAusstehend: Der Auftrag ist offen und keine Störung ist
	// dokumentiert; die Signatur wird in Kürze erwartet.
	SignaturstatusAusstehend Signaturstatus = "ausstehend"
)

// SignaturstatusErgebnis ist das Ergebnis von DetermineSignaturstatus.
type SignaturstatusErgebnis struct {
	Status Signaturstatus
	// Signatur ist bei Vorhanden und Nachsigniert gesetzt.
	Signatur *Signatur
	// AusfallGrund ist bei Ausfall gesetzt: der Endstatus des Auftrags
	// (fehlgeschlagen, tse_nicht_konfiguriert) oder die Grund-Art des aktiven
	// Störungszeitraums.
	AusfallGrund string
}

// DetermineSignaturstatus ist die einzige Implementierung des Ausfallbegriffs:
// Beleg-Abruf und Kassenabschluss-Gate urteilen über diese Funktion. Der
// Ausfallbegriff ist rein status- und zeitraumbasiert: Endstatus des Auftrags
// (fehlgeschlagen, tse_nicht_konfiguriert) oder offener Auftrag bei aktivem
// Störungszeitraum. Fehlversuche unterhalb der Maximalzahl und geschlossene
// Zeiträume zählen nicht — ein offener Auftrag ohne aktive Störung ist
// ausstehend, nie Ausfall.
func DetermineSignaturstatus(auftrag SignaturauftragStand, aktiveStoerung *Stoerung) SignaturstatusErgebnis {
	switch auftrag.Status {
	case StatusErledigt:
		if auftrag.Signatur.LogTimeEnd.Sub(auftrag.ErstelltAm) > NachsigniertSchwelle {
			return SignaturstatusErgebnis{Status: SignaturstatusNachsigniert, Signatur: auftrag.Signatur}
		}
		return SignaturstatusErgebnis{Status: SignaturstatusVorhanden, Signatur: auftrag.Signatur}
	case StatusFehlgeschlagen, StatusTSENichtKonfiguriert:
		return SignaturstatusErgebnis{Status: SignaturstatusAusfall, AusfallGrund: auftrag.Status}
	default: // offen
		if aktiveStoerung != nil {
			return SignaturstatusErgebnis{Status: SignaturstatusAusfall, AusfallGrund: aktiveStoerung.GrundArt}
		}
		return SignaturstatusErgebnis{Status: SignaturstatusAusstehend}
	}
}
