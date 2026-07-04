package tse

import "time"

// Zeitschwellen des Signaturstatus und des Stoerungsprotokolls — die einzige
// Stelle dieser Konstanten; Signaturstatus-Funktion, Rueckstands-Watchdog und
// Dashboard-Warnung referenzieren sie.
const (
	// NachsigniertSchwelle: Eine Signatur, die spaeter als diese Spanne nach
	// der Auftragserstellung entsteht, traegt das Nachsigniert-Kennzeichen —
	// ihre TSE-Zeitpunkte weichen sichtbar vom Belegdatum ab.
	NachsigniertSchwelle = time.Minute
	// RueckstandSchwelle: Ab diesem Alter des aeltesten offenen Auftrags
	// dokumentiert der Rueckstands-Watchdog einen Rueckstands-Stoerungszeitraum.
	RueckstandSchwelle = 2 * time.Minute
	// WatchdogTickIntervall ist der Pruef-Takt des Rueckstands-Watchdogs; die
	// Rueckstands-Schwelle materialisiert nur am Tick.
	WatchdogTickIntervall = 10 * time.Second
)

// Signaturstatus ist das Urteil der Signaturstatus-Funktion ueber einen
// Signaturauftrag — genau eine von vier Ergebnisarten.
type Signaturstatus string

const (
	// SignaturstatusVorhanden: Die Signatur liegt vor.
	SignaturstatusVorhanden Signaturstatus = "vorhanden"
	// SignaturstatusNachsigniert: Die Signatur liegt vor, entstand aber
	// verspaetet (spaeter als NachsigniertSchwelle nach der Auftragserstellung);
	// der Beleg traegt das Nachsigniert-Kennzeichen.
	SignaturstatusNachsigniert Signaturstatus = "nachsigniert"
	// SignaturstatusAusfall: Keine Signatur, mit belegbarem Grund — Endstatus
	// des Auftrags oder offener Auftrag bei aktivem Stoerungszeitraum.
	SignaturstatusAusfall Signaturstatus = "ausfall"
	// SignaturstatusAusstehend: Der Auftrag ist offen und keine Stoerung ist
	// dokumentiert; die Signatur wird in Kuerze erwartet.
	SignaturstatusAusstehend Signaturstatus = "ausstehend"
)

// SignaturstatusErgebnis ist das Ergebnis von BestimmeSignaturstatus.
type SignaturstatusErgebnis struct {
	Status Signaturstatus
	// Signatur ist bei Vorhanden und Nachsigniert gesetzt.
	Signatur *Signatur
	// AusfallGrund ist bei Ausfall gesetzt: der Endstatus des Auftrags
	// (fehlgeschlagen, tse_nicht_konfiguriert) oder die Grund-Art des aktiven
	// Stoerungszeitraums.
	AusfallGrund string
}

// BestimmeSignaturstatus ist die einzige Implementierung des Ausfallbegriffs:
// Beleg-Abruf und Kassenabschluss-Gate urteilen ueber diese Funktion. Der
// Ausfallbegriff ist rein status- und zeitraumbasiert: Endstatus des Auftrags
// (fehlgeschlagen, tse_nicht_konfiguriert) oder offener Auftrag bei aktivem
// Stoerungszeitraum. Fehlversuche unterhalb der Maximalzahl und geschlossene
// Zeitraeume zaehlen nicht — ein offener Auftrag ohne aktive Stoerung ist
// ausstehend, nie Ausfall.
func BestimmeSignaturstatus(auftrag SignaturauftragStand, aktiveStoerung *Stoerung) SignaturstatusErgebnis {
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
