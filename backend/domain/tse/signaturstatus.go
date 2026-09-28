package tse

import "time"

const (
	// NachsigniertSchwelle: a later signature gets the Nachsigniert mark, as its TSE times visibly differ from the receipt date.
	NachsigniertSchwelle = time.Minute
	// RueckstandSchwelle is the age of the oldest open job at which the watchdog opens a rueckstand period.
	RueckstandSchwelle = 2 * time.Minute
	// WatchdogTickIntervall bounds how late past RueckstandSchwelle a period opens.
	WatchdogTickIntervall = 10 * time.Second
)

type Signaturstatus string

const (
	SignaturstatusVorhanden Signaturstatus = "vorhanden"
	// SignaturstatusNachsigniert: signed later than NachsigniertSchwelle after the job was created.
	SignaturstatusNachsigniert Signaturstatus = "nachsigniert"
	SignaturstatusAusfall      Signaturstatus = "ausfall"
	SignaturstatusAusstehend   Signaturstatus = "ausstehend"
)

type SignaturstatusErgebnis struct {
	Status Signaturstatus
	// Signatur is set for Vorhanden and Nachsigniert.
	Signatur *Signatur
	// AusfallGrund is the job's final status or the active period's Grund-Art.
	AusfallGrund string
}

// DetermineSignaturstatus is the only implementation of Ausfall; retries below the maximum and closed periods never count.
// See docs/handbuch.md §3.13 (Störungsprotokoll und Signaturstatus).
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
