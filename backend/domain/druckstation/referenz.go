package druckstation

import (
	"fmt"
	"strings"
)

// ReferenzArt names the source of a print job. A Referenz is "<art>:<rest>": the event ID, or the
// Kategorie for a Testbon. Stored druckauftraege keep these strings, so a value never changes.
type ReferenzArt string

const (
	ReferenzBestellung           ReferenzArt = "bestellung-aufgenommen"
	ReferenzZahlung              ReferenzArt = "zahlung-kassiert"
	ReferenzDirektverkauf        ReferenzArt = "direktverkauf-getaetigt"
	ReferenzDirektverkaufStorno  ReferenzArt = "direktverkauf-storniert"
	ReferenzStornierung          ReferenzArt = "stornierung-erteilt"
	ReferenzPfandrueckgabe       ReferenzArt = "pfand-zurueckgegeben"
	ReferenzPfandrueckgabeStorno ReferenzArt = "pfandrueckgabe-storniert"
	referenzTestdruck            ReferenzArt = "testdruck"
)

var referenzLabel = map[ReferenzArt]string{
	ReferenzBestellung:           "Bestellung",
	ReferenzZahlung:              "Zahlung",
	ReferenzDirektverkauf:        "Direktverkauf",
	ReferenzDirektverkaufStorno:  "Direktverkauf-Storno",
	ReferenzStornierung:          "Stornierung",
	ReferenzPfandrueckgabe:       "Pfandrückgabe",
	ReferenzPfandrueckgabeStorno: "Pfandrückgabe-Storno",
}

func EventReferenz(art ReferenzArt, eventID int) string {
	return fmt.Sprintf("%s:%d", art, eventID)
}

func TestdruckReferenz(kategorie Kategorie) string {
	return string(referenzTestdruck) + ":" + string(kategorie)
}

// ReferenzAnzeige names a Referenz for the admin; an unknown format stays as it is.
func ReferenzAnzeige(referenz string) string {
	art, rest, ok := strings.Cut(referenz, ":")
	if !ok || rest == "" {
		return referenz
	}
	if ReferenzArt(art) == referenzTestdruck {
		return "Testbon " + Kategorie(rest).Anzeigename()
	}
	label, ok := referenzLabel[ReferenzArt(art)]
	if !ok {
		return referenz
	}
	return label + " Nr. " + rest
}

// druckfehlerMeldungen matches the error texts of the print relay (windows/relay) and the demo seed's
// cover-open text, in priority order.
var druckfehlerMeldungen = []struct{ schluessel, meldung string }{
	{"papier", "Papier leer"},
	{"abdeckung", "Abdeckung offen"},
	{"nicht erreichbar", "Drucker nicht erreichbar"},
	{"senden fehlgeschlagen", "Übertragung fehlgeschlagen"},
}

// DruckfehlerMeldung turns a raw relay error chain into a short message; an unknown text becomes "Druckfehler".
func DruckfehlerMeldung(fehler string) string {
	text := strings.ToLower(fehler)
	for _, m := range druckfehlerMeldungen {
		if strings.Contains(text, m.schluessel) {
			return m.meldung
		}
	}
	return "Druckfehler"
}
