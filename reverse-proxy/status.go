package main

type certState int

const (
	certNone    certState = iota // noch kein vertrauenswürdiges Zertifikat
	certValid                    // gültige, öffentlich vertrauenswürdige Kette (LE)
	certExpired                  // Zertifikat vorhanden, aber abgelaufen
)

type notice int

const (
	noticeGreen    notice = iota // grüne Adresse ist aktiv — alles gut
	noticeIssuing                // Zertifikat wird (noch) ausgestellt → Fallback
	noticeRenewing               // Zertifikat abgelaufen, wird erneuert → Fallback
	noticeRebind                 // Router-Rebind-Schutz blockiert den Namen → Anleitung
	noticeNoGreen                // keine grüne Adresse möglich (kein State / keine LAN-IP)
)

type statusInputs struct {
	cert        certState
	rebindOK    bool
	greenURL    string // "" ⇒ keine vertrauenswürdige Adresse möglich (kein State/keine IP)
	fallbackURL string // "" ⇒ keine LAN-IP bekannt
}

type statusView struct {
	primaryURL  string // prominent angezeigte Adresse
	greenActive bool   // grüne Adresse erreichbar & vertrauenswürdig
	showQR      bool   // QR-Code für die grüne Adresse anzeigen
	refresh     bool   // Seite aktualisiert sich selbst (bis „grün")
	notice      notice
}

// decideStatus bildet die beobachteten Eingaben auf die Anzeige ab. Die Reihenfolge
// ist bewusst: ohne grünen Namen gibt es nur den Fallback, und ein blockierender
// Rebind-Schutz macht die grüne Adresse auch mit gültigem Zertifikat unerreichbar —
// er hat darum Vorrang vor der Zertifikatslage. Ohne grünen Namen aktualisiert sich
// die Seite nicht selbst: State und LAN-IP entstehen nur beim Start (runLANMode),
// ein Refresh verspräche eine Änderung, die erst ein Neustart bringt.
func decideStatus(in statusInputs) statusView {
	switch {
	case in.greenURL == "":
		return statusView{primaryURL: in.fallbackURL, notice: noticeNoGreen}
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
