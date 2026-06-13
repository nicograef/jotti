package main

// certState beschreibt das Ergebnis der Zertifikats-Probe gegen den eigenen
// Caddy: noch kein vertrauenswürdiges Zertifikat (interne CA / noch nicht
// ausgestellt), eine gültige öffentlich vertrauenswürdige Let's-Encrypt-Kette
// oder ein abgelaufenes Zertifikat.
type certState int

const (
	certNone    certState = iota // noch kein vertrauenswürdiges Zertifikat
	certValid                    // gültige, öffentlich vertrauenswürdige Kette (LE)
	certExpired                  // Zertifikat vorhanden, aber abgelaufen
)

// notice ist der Hinweis, den die Status-Seite je nach Startzustand anzeigt.
type notice int

const (
	noticeGreen    notice = iota // grüne Adresse ist aktiv — alles gut
	noticeIssuing                // Zertifikat wird (noch) ausgestellt → Fallback
	noticeRenewing               // Zertifikat abgelaufen, wird erneuert → Fallback
	noticeRebind                 // Router-Rebind-Schutz blockiert den Namen → Anleitung
	noticeNoGreen                // keine grüne Adresse möglich (kein State / keine LAN-IP)
)

// statusInputs sind die beim Seitenaufruf beobachteten Eingaben der
// Start-Zustandslogik.
type statusInputs struct {
	cert        certState
	rebindOK    bool
	greenURL    string // "" ⇒ keine vertrauenswürdige Adresse möglich (kein State/keine IP)
	fallbackURL string // "" ⇒ keine LAN-IP bekannt
}

// statusView ist die reine Anzeige-Entscheidung: welche Adresse prominent ist,
// ob ein QR-Code erscheint, ob sich die Seite selbst aktualisiert und welcher
// Hinweis gilt.
type statusView struct {
	primaryURL  string // prominent angezeigte Adresse
	greenActive bool   // grüne Adresse erreichbar & vertrauenswürdig
	showQR      bool   // QR-Code für die grüne Adresse anzeigen
	refresh     bool   // Seite aktualisiert sich selbst (bis „grün")
	notice      notice
}

// decideStatus bildet die beobachteten Eingaben auf die Anzeige-Entscheidung ab.
// Reine Funktion ohne I/O — über alle Startzustände unit-getestet (kein
// Zertifikat / gültig / abgelaufen / Rebind blockiert). Die Reihenfolge ist
// bewusst: ohne grünen Namen gibt es nur den Fallback; ein blockierender
// Rebind-Schutz macht die grüne Adresse auch mit gültigem Zertifikat
// unerreichbar und hat darum Vorrang vor der Zertifikatslage.
func decideStatus(in statusInputs) statusView {
	switch {
	case in.greenURL == "":
		return statusView{primaryURL: in.fallbackURL, refresh: true, notice: noticeNoGreen}
	case !in.rebindOK:
		return statusView{primaryURL: in.fallbackURL, refresh: true, notice: noticeRebind}
	case in.cert == certValid:
		return statusView{primaryURL: in.greenURL, greenActive: true, showQR: true, notice: noticeGreen}
	case in.cert == certExpired:
		return statusView{primaryURL: in.fallbackURL, refresh: true, notice: noticeRenewing}
	default: // certNone
		return statusView{primaryURL: in.fallbackURL, refresh: true, notice: noticeIssuing}
	}
}
