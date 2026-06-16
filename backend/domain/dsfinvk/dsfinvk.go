// Package dsfinvk transformiert die Events und Stammdaten einer Kassensitzung
// seiteneffektfrei in ein DSFinV-K-Archiv (CSV-Dateien, index.xml und die
// gdpdu-01-09-2004.dtd). Der fiskalisch heikle Teil — GV-Typ-/Beleg-Mapping,
// Steueraufteilung, TSE-Daten — liegt vollständig hier und ist über
// Golden-File-Tests prüfbar. Das Paket kennt kein I/O: der Orchestrator lädt die
// Daten und reicht sie als Snapshot plus Event-Liste herein.
package dsfinvk

import (
	"fmt"
	"strconv"
	"time"

	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

// Version ist der deklarierte DSFinV-K-Versionsstring. Konfigurierbar gehalten,
// da die Tabellenstruktur seit v2.0 stabil ist; aktuell verbindlich ist v2.5.
const Version = "2.5"

// Snapshot ist der lesende Stammdaten-Eingang des Mappers: alles, was der Export
// neben den Events selbst braucht. Der Orchestrator lädt ihn; der Mapper bleibt
// rein.
type Snapshot struct {
	// KasseSeriennummer speist Z_KASSE_ID und KASSE_SERIENNR (UUID der Kasse).
	KasseSeriennummer string
	// Erstellung ist der Zeitpunkt des Kassenabschlusses (Z_ERSTELLUNG). Bei
	// einer offenen Sitzung der Exportzeitpunkt, bei einer abgeschlossenen der
	// Zeitpunkt des Tagesabschlusses.
	Erstellung time.Time
	// KassensitzungNr ist die Z_NR des Abschlusses.
	KassensitzungNr int
	Betreiber       settings.Betreiber
	TSEStammdaten   settings.TSEStammdaten
	// Tischnamen bildet Tisch-IDs auf ihren Namen ab (Quelle des
	// ABRECHNUNGSKREIS). Fehlt ein Tisch (gelöscht), synthetisiert der Mapper
	// "Tisch N".
	Tischnamen map[int]string
}

// itoa formatiert eine Ganzzahl dezimal; ein kurzer Alias hält die dichten
// Zeilen-Literale der Tabellen lesbar.
func itoa(n int) string { return strconv.Itoa(n) }

// ptr liefert den Wert eines optionalen Strings oder "" bei nil.
func ptr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// storno bildet ein Storno-Kennzeichen als DSFinV-K-Boolean ("0"/"1") ab.
func storno(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// formatAmount stellt einen Cent-Betrag als Dezimalzahl mit Punkt und zwei
// Nachkommastellen dar, z. B. 500 -> "5.00", -150 -> "-1.50". Zwei Stellen sind
// der DSFinV-K-Regelfall (technisch bis fünf zulässig); intern wird durchgehend
// in Cent gerechnet, erst hier dezimal dargestellt.
func formatAmount(cents int) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}

// formatQuantity stellt eine Stückzahl mit drei Nachkommastellen dar (MENGE).
func formatQuantity(menge int) string {
	return fmt.Sprintf("%d.000", menge)
}

// formatPercent stellt einen Steuersatz als Prozentwert mit zwei Nachkommastellen
// dar, z. B. 19 -> "19.00".
func formatPercent(prozent int) string {
	return fmt.Sprintf("%d.00", prozent)
}

// ustSchluessel bildet einen jotti-Steuersatz auf den DSFinV-K-Umsatzsteuer-
// schlüssel (Anlage 2) ab: 1 = Regelsteuersatz (19 %), 2 = ermäßigter Satz
// (7 %), 6 = umsatzsteuerfrei (0 %, z. B. Zweckbetrieb § 67a AO). Die
// Steueraufteilung entfaltet kombi vorab in regel und ermäßigt, daher kommt hier
// nie KombiSteuersatz an.
func ustSchluessel(satz steuer.Steuersatz) int {
	switch satz {
	case steuer.RegelSteuersatz:
		return 1
	case steuer.ErmaessigtSteuersatz:
		return 2
	case steuer.BefreitSteuersatz:
		return 6
	default:
		return 0
	}
}

// ustBeschreibung liefert die UST_BESCHR für die vat.csv je Schlüssel.
func ustBeschreibung(schluessel int) string {
	switch schluessel {
	case 1:
		return "Allgemeiner Steuersatz"
	case 2:
		return "Ermäßigter Steuersatz"
	case 6:
		return "Umsatzsteuerfrei"
	default:
		return ""
	}
}
