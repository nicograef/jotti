package tse

import "time"

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

type Signaturstatus string

const (
	SignaturstatusVorhanden Signaturstatus = "vorhanden"
	// SignaturstatusNachsigniert: Signatur liegt vor, entstand aber später als
	// NachsigniertSchwelle; der Beleg trägt das Nachsigniert-Kennzeichen.
	SignaturstatusNachsigniert Signaturstatus = "nachsigniert"
	SignaturstatusAusfall      Signaturstatus = "ausfall"
	SignaturstatusAusstehend   Signaturstatus = "ausstehend"
)

type SignaturstatusErgebnis struct {
	Status Signaturstatus
	// Signatur ist bei Vorhanden und Nachsigniert gesetzt.
	Signatur *Signatur
	// AusfallGrund ist bei Ausfall gesetzt: der Endstatus des Auftrags
	// (fehlgeschlagen, tse_nicht_konfiguriert) oder die Grund-Art des aktiven
	// Störungszeitraums.
	AusfallGrund string
}

// DetermineSignaturstatus ist die einzige Implementierung des Ausfallbegriffs;
// Beleg-Abruf und Kassenabschluss-Gate urteilen über sie. Ausfall ist allein der
// Endstatus des Auftrags (fehlgeschlagen, tse_nicht_konfiguriert) oder ein offener
// Auftrag bei aktivem Störungszeitraum — Fehlversuche unterhalb der Maximalzahl
// und geschlossene Zeiträume zählen nicht.
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
