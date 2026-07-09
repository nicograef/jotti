package dsfinvkpruefung

import (
	"fmt"
	"sort"
	"strings"
)

// Regel-Kennungen der Inhaltsprüfung. Sie ergänzen die Strukturregeln (csv.go,
// indexxml.go, paket.go) um fachliche Konsistenzregeln, die nicht aus der reinen
// Dateiform, sondern aus dem Zusammenspiel der DSFinV-K-Tabellen folgen.
const (
	regelStornoReferenz      = "storno-referenz"
	regelStornoBonStorno     = "storno-bon-storno-kennzeichen"
	regelKombiSteuer         = "kombi-steueraufteilung"
	regelBedienerLeer        = "bediener-feld-leer"
	regelBedienerIDNumerisch = "bediener-id-nicht-numerisch"
	regelTagesabschlussName  = "tagesabschluss-bon-name"
	regelTSEStammdaten       = "tse-stammdaten-unvollstaendig"
	regelAbrechnungskreis    = "abrechnungskreis-fehlt"
)

// Feste DSFinV-K-Werte, gegen die die Inhaltsregeln prüfen (Anhang B/E, Anlage 2).
const (
	bonTypBeleg             = "Beleg"          // Anhang B: abgeschlossener Kassenvorgang (Zahlung/Warenrücknahme)
	bonTypSonstige          = "AVSonstige"     // Anhang B: sonstiger anderer Vorgang (Tagesabschluss)
	refTypTransaktion       = "Transaktion"    // Anhang E: Referenz innerhalb der DSFinV-K
	gvTypUmsatz             = "Umsatz"         // Anhang C: realisierter Umsatz (grenzt Storno von Bargeldbewegung ab)
	bonStornoKein           = "0"              // BON_STORNO: keine Vorgangsaufhebung (jotti-Negativdarstellung)
	tagesabschlussName      = "Tagesabschluss" // amtlich verpflichtender BON_NAME des AVSonstige-Abschlussbons
	ustSchluesselRegel      = "1"              // Anlage 2: 19 % Regelsteuersatz
	ustSchluesselErmaessigt = "2"              // Anlage 2: 7 % ermäßigter Steuersatz
)

// pruefeInhalt wendet die fachlichen Inhaltsregeln auf die bereits geparsten
// Tabellen an. Es setzt auf denselben index.xml-getriebenen Tabellenschnitt wie
// die Strukturprüfung: nur deklarierte-und-vorhandene CSVs mit passender Kopfzeile
// werden inhaltlich betrachtet (die Strukturprüfung hat Format-/Kopfzeilenfehler
// bereits gemeldet). Ein leeres Ergebnis bedeutet: inhaltlich konsistent.
//
// Referenz: DSFinV-K 2.4 (Anhang B Vorgangstypen, Anhang E Referenzen, Anlage 2
// USt-Schlüssel) sowie docs/compliance.md Abschnitt 6.4 (Bediener) und 6.6 (Storno).
func pruefeInhalt(dateien map[string][]byte, tabellen []indexTabelle) []Befund {
	daten := ladeTabellendaten(dateien, tabellen)

	var befunde []Befund
	befunde = append(befunde, pruefeStornoReferenzen(daten)...)
	befunde = append(befunde, pruefeKombiSteueraufteilung(daten)...)
	befunde = append(befunde, pruefeBedienerFelder(daten)...)
	befunde = append(befunde, pruefeTagesabschlussZeile(daten)...)
	befunde = append(befunde, pruefeTSEStammdaten(daten)...)
	befunde = append(befunde, pruefeAbrechnungskreise(daten)...)
	return befunde
}

// tabellendaten ist die zeilenweise, spaltenadressierbare Sicht einer geparsten
// CSV: die Spaltennamen aus der index.xml auf ihren Index abgebildet und die
// Datenzeilen (ohne Kopfzeile) als Felder. Fehlt eine Tabelle oder weicht ihre
// Kopfzeile ab, ist sie hier schlicht nicht enthalten (die Strukturprüfung hat
// das bereits gemeldet).
type tabellendaten struct {
	spalten map[string]int
	zeilen  [][]string
}

// wert liefert den Feldwert einer Datenzeile für den gegebenen Spaltennamen.
// Existiert die Spalte nicht oder ist der Index außerhalb der Zeile, liefert es
// den leeren String — die Aufrufer prüfen nur vorhandene Tabellen.
func (t tabellendaten) wert(zeile []string, spalte string) string {
	idx, ok := t.spalten[spalte]
	if !ok || idx >= len(zeile) {
		return ""
	}
	return zeile[idx]
}

// ladeTabellendaten parst alle deklarierten-und-vorhandenen CSVs mit passender
// Kopfzeile in die zeilenweise Sicht. Der Parser ist derselbe wie in der
// Strukturprüfung (splitFelder, zerlegeCRLF); Tabellen mit abweichender Kopfzeile
// bleiben außen vor, weil eine spaltenweise Inhaltsprüfung dort nicht greift.
func ladeTabellendaten(dateien map[string][]byte, tabellen []indexTabelle) map[string]tabellendaten {
	out := make(map[string]tabellendaten, len(tabellen))
	for _, tab := range tabellen {
		inhalt, ok := dateien[tab.URL]
		if !ok || len(inhalt) == 0 {
			continue
		}
		zeilen := zerlegeCRLF(string(inhalt))
		if len(zeilen) == 0 {
			continue
		}
		erwartet := spaltenNamen(tab)
		header := splitFelder(zeilen[0])
		if !gleicheReihenfolge(header, erwartet) {
			continue
		}
		spalten := make(map[string]int, len(erwartet))
		for i, name := range erwartet {
			spalten[name] = i
		}
		daten := tabellendaten{spalten: spalten}
		for i := 1; i < len(zeilen); i++ {
			felder := splitFelder(zeilen[i])
			if len(felder) != len(erwartet) {
				continue // Feldanzahl-Fehler meldet bereits die Strukturprüfung.
			}
			daten.zeilen = append(daten.zeilen, felder)
		}
		out[tab.URL] = daten
	}
	return out
}

// pruefeStornoReferenzen prüft die DSFinV-K-konforme Abbildung eines Stornos als
// eigenen Geschäftsvorfall: Ein geldwirksamer Storno (Warenrücknahme, negativer
// UMS_BRUTTO auf einem "Beleg"-Bon mit GV_TYP "Umsatz") muss (a) BON_STORNO = "0"
// führen — jotti nutzt die zulässige Negativdarstellung statt der Vorgangsaufhebung
// (BON_STORNO = "1") — und (b) in references.csv über REF_TYP "Transaktion" und ein
// gefülltes REF_BON_ID auf den Ursprungsbeleg verweisen.
//
// Nicht-steuerbare Bargeldabflüsse (Geldtransit-Entnahme, Kassenfehlbetrag) sind
// zwar ebenfalls negative "Beleg"-Bons, aber keine Stornos: sie tragen GV_TYP
// "Geldtransit"/"DifferenzSollIst" (nicht "Umsatz") und referenzieren zulässig
// keinen Ursprungsbeleg. Die Regel grenzt darüber ab.
//
// Referenz: DSFinV-K 2.4 Feldbeschreibung BON_STORNO (Tz. 3.2.1, "zweiter Datensatz
// mit umgekehrtem Vorzeichen"), Tz. 4.2.2/4.2.5 (Warenrücknahme als Negativbeleg),
// Anhang E / Tz. 4.2.2 (Auflösung einer Forderung: REF_TYP "Transaktion",
// REF_Z_NR, REF_Z_KASSE_ID, REF_BON_ID) und docs/compliance.md Abschnitt 6.6
// (BON_STORNO bleibt in allen Fällen 0, jotti kennt keine Vorgangsaufhebung).
func pruefeStornoReferenzen(daten map[string]tabellendaten) []Befund {
	transactions, ok := daten["transactions.csv"]
	if !ok {
		return nil
	}
	refDaten := daten["references.csv"]
	referenzen := referenzenNachBonID(refDaten)
	umsatzBons := umsatzBonsAusLines(daten["lines.csv"])

	var befunde []Befund
	for _, zeile := range transactions.zeilen {
		if transactions.wert(zeile, "BON_TYP") != bonTypBeleg {
			continue
		}
		if !istNegativerBetrag(transactions.wert(zeile, "UMS_BRUTTO")) {
			continue
		}
		bonID := transactions.wert(zeile, "BON_ID")
		// Nur Umsatz-Belege sind Stornos; Bargeldabflüsse (Geldtransit,
		// Kassenfehlbetrag) sind negative Belege ohne Referenzpflicht.
		if !umsatzBons[bonID] {
			continue
		}

		// (a) BON_STORNO muss "0" sein: jotti nutzt die Negativdarstellung, nie die
		// Vorgangsaufhebung (compliance.md 6.6).
		if bonStorno := transactions.wert(zeile, "BON_STORNO"); bonStorno != bonStornoKein {
			befunde = append(befunde, Befund{
				Datei:   "transactions.csv",
				Regel:   regelStornoBonStorno,
				Meldung: fmt.Sprintf("Storno-Beleg %q: BON_STORNO = %q, erwartet %q (jotti nutzt die Negativdarstellung, keine Vorgangsaufhebung)", bonID, bonStorno, bonStornoKein),
			})
		}

		// (b) Es muss eine references.csv-Zeile mit REF_TYP "Transaktion" und
		// gefülltem REF_BON_ID auf den Ursprungsbeleg geben.
		if !hatTransaktionsReferenz(refDaten, referenzen[bonID]) {
			befunde = append(befunde, Befund{
				Datei:   "references.csv",
				Regel:   regelStornoReferenz,
				Meldung: fmt.Sprintf("Storno-Beleg %q hat keine Referenz mit REF_TYP %q und gefülltem REF_BON_ID auf den Ursprungsbeleg", bonID, refTypTransaktion),
			})
		}
	}
	return befunde
}

// referenzenNachBonID gruppiert die references.csv-Zeilen nach dem referenzierenden
// BON_ID.
func referenzenNachBonID(refs tabellendaten) map[string][][]string {
	out := map[string][][]string{}
	for _, zeile := range refs.zeilen {
		bonID := refs.wert(zeile, "BON_ID")
		out[bonID] = append(out[bonID], zeile)
	}
	return out
}

// umsatzBonsAusLines liefert die Menge der BON_IDs, deren Positionen in lines.csv
// den GV_TYP "Umsatz" tragen — also die umsatzwirksamen Belege (Verkauf, Storno).
// Bargeldbewegungen (Anfangsbestand, Geldtransit, Kassendifferenz) tragen einen
// anderen GV_TYP und sind hier nicht enthalten.
func umsatzBonsAusLines(lines tabellendaten) map[string]bool {
	out := map[string]bool{}
	for _, zeile := range lines.zeilen {
		if lines.wert(zeile, "GV_TYP") == gvTypUmsatz {
			out[lines.wert(zeile, "BON_ID")] = true
		}
	}
	return out
}

// hatTransaktionsReferenz meldet, ob unter den Referenzzeilen eine mit REF_TYP
// "Transaktion" und nicht-leerem REF_BON_ID ist. Der Zugriff auf die Felder nutzt
// die feste Spaltenordnung der references.csv (BON_ID, POS_ZEILE, REF_TYP,
// REF_NAME, REF_DATUM, REF_Z_KASSE_ID, REF_Z_NR, REF_BON_ID).
func hatTransaktionsReferenz(refs tabellendaten, zeilen [][]string) bool {
	for _, zeile := range zeilen {
		if refs.wert(zeile, "REF_TYP") == refTypTransaktion && refs.wert(zeile, "REF_BON_ID") != "" {
			return true
		}
	}
	return false
}

// pruefeKombiSteueraufteilung stellt sicher, dass ein Bon mit gemischten Steuersätzen
// (z. B. Essen 7 % + Getränk 19 % in einer Bestellung) tatsächlich getrennte
// USt-Zeilen je Schlüssel führt — sowohl auf Positionsebene (lines_vat.csv) als
// auch auf Bonkopfebene (transactions_vat.csv). Konkret: existiert in lines_vat.csv
// für einen Bon sowohl der Schlüssel 7 % als auch 19 %, muss dieselbe Aufteilung
// in transactions_vat.csv erscheinen (der Bonkopf darf die Positionsaufteilung nicht
// zu einer Zeile verschmelzen).
//
// Referenz: DSFinV-K 2.4 Tz. 3.2.5/3.2.6 (Bonkopf_USt/Bonpos_USt: USt-Aufschlüsselung
// je Schlüssel), Anlage 2 (USt-Schlüssel 1 = 19 %, 2 = 7 %) und docs/steuerrecht.md
// (Kombi-Splitting Gastronomie).
func pruefeKombiSteueraufteilung(daten map[string]tabellendaten) []Befund {
	linesVat, okL := daten["lines_vat.csv"]
	if !okL {
		return nil
	}
	transVat, okT := daten["transactions_vat.csv"]
	if !okT {
		return nil
	}

	linesSchluessel := schluesselNachBonID(linesVat)
	transSchluessel := schluesselNachBonID(transVat)

	var befunde []Befund
	for _, bonID := range sortierteSchluessel(linesSchluessel) {
		sk := linesSchluessel[bonID]
		if !sk[ustSchluesselErmaessigt] || !sk[ustSchluesselRegel] {
			continue // kein Kombi-Bon
		}
		ziel := transSchluessel[bonID]
		if !ziel[ustSchluesselErmaessigt] || !ziel[ustSchluesselRegel] {
			befunde = append(befunde, Befund{
				Datei:   "transactions_vat.csv",
				Regel:   regelKombiSteuer,
				Meldung: fmt.Sprintf("Bon %q hat in lines_vat.csv sowohl 7 %% als auch 19 %% (Schlüssel %s und %s), aber transactions_vat.csv teilt den Bonkopf nicht entsprechend auf", bonID, ustSchluesselErmaessigt, ustSchluesselRegel),
			})
		}
	}
	return befunde
}

// schluesselNachBonID sammelt je BON_ID die vorkommenden UST_SCHLUESSEL-Werte.
func schluesselNachBonID(daten tabellendaten) map[string]map[string]bool {
	out := map[string]map[string]bool{}
	for _, zeile := range daten.zeilen {
		bonID := daten.wert(zeile, "BON_ID")
		if out[bonID] == nil {
			out[bonID] = map[string]bool{}
		}
		out[bonID][daten.wert(zeile, "UST_SCHLUESSEL")] = true
	}
	return out
}

// pruefeBedienerFelder prüft die Bedienerfelder jedes Bonkopfs: BEDIENER_NAME muss
// gefüllt sein (der zum Buchungszeitpunkt eingefrorene unternehmensinterne
// Benutzername) und BEDIENER_ID muss die numerische, stabile Benutzer-ID (user_id)
// tragen.
//
// Referenz: DSFinV-K 2.4 Feldbeschreibung BEDIENER_ID ("unternehmensinterne
// Kennung") und BEDIENER_NAME ("unternehmensinterner Name der Person, die den
// Vorgang erfasst"), Tz. 3.2.1, sowie docs/compliance.md Abschnitt 6.4 (BEDIENER_ID
// = user_id, BEDIENER_NAME = kassenjournal.user_name).
func pruefeBedienerFelder(daten map[string]tabellendaten) []Befund {
	transactions, ok := daten["transactions.csv"]
	if !ok {
		return nil
	}

	var befunde []Befund
	for _, zeile := range transactions.zeilen {
		bonID := transactions.wert(zeile, "BON_ID")
		if transactions.wert(zeile, "BEDIENER_NAME") == "" {
			befunde = append(befunde, Befund{
				Datei:   "transactions.csv",
				Regel:   regelBedienerLeer,
				Meldung: fmt.Sprintf("Bon %q: BEDIENER_NAME ist leer (der eingefrorene Bedienername ist verpflichtend)", bonID),
			})
		}
		if id := transactions.wert(zeile, "BEDIENER_ID"); !istNumerisch(id) {
			befunde = append(befunde, Befund{
				Datei:   "transactions.csv",
				Regel:   regelBedienerIDNumerisch,
				Meldung: fmt.Sprintf("Bon %q: BEDIENER_ID = %q ist keine numerische Benutzer-ID (user_id)", bonID, id),
			})
		}
	}
	return befunde
}

// pruefeTagesabschlussZeile prüft die Abschluss-Sonderzeile: Ein AVSonstige-Bon
// (Tagesabschluss) muss den amtlich verpflichtenden BON_NAME "Tagesabschluss"
// tragen. Umgekehrt darf nur ein AVSonstige-Bon diesen Namen führen.
//
// Referenz: DSFinV-K 2.4 Tz. 4.1.1/4.1.2 (BON_TYP AVSonstige "zwingend zu erläutern
// über BON_NAME") und Anhang B ("AVSonstige … Zusätzlich ist zwingend das Feld
// BON_NAME mit einer individuellen Beschreibung zu füllen").
func pruefeTagesabschlussZeile(daten map[string]tabellendaten) []Befund {
	transactions, ok := daten["transactions.csv"]
	if !ok {
		return nil
	}

	var befunde []Befund
	for _, zeile := range transactions.zeilen {
		bonTyp := transactions.wert(zeile, "BON_TYP")
		bonName := transactions.wert(zeile, "BON_NAME")
		bonID := transactions.wert(zeile, "BON_ID")
		if bonTyp == bonTypSonstige && bonName != tagesabschlussName {
			befunde = append(befunde, Befund{
				Datei:   "transactions.csv",
				Regel:   regelTagesabschlussName,
				Meldung: fmt.Sprintf("AVSonstige-Bon %q hat BON_NAME %q, erwartet %q (Abschlussbon)", bonID, bonName, tagesabschlussName),
			})
		}
	}
	return befunde
}

// pruefeTSEStammdaten stellt sicher, dass die tse.csv die TSS-Stammdaten
// vollständig führt: Seriennummer, Signaturalgorithmus, öffentlicher Schlüssel und
// mindestens der erste Zertifikatsblock müssen gefüllt sein. Ohne diese Angaben ist
// die Signaturprüfung des Exports nicht möglich.
//
// Referenz: DSFinV-K 2.4 Tz. 3.2.7 "Datei: Stamm_TSE" (Felder TSE_SERIAL,
// TSE_SIG_ALGO, TSE_PUBLIC_KEY, TSE_ZERTIFIKAT_I/_II) — die Seriennummer entspricht
// laut BSI TR-03153 der TSE-Seriennummer.
func pruefeTSEStammdaten(daten map[string]tabellendaten) []Befund {
	tse, ok := daten["tse.csv"]
	if !ok {
		return nil
	}

	pflichtfelder := []string{"TSE_SERIAL", "TSE_SIG_ALGO", "TSE_PUBLIC_KEY", "TSE_ZERTIFIKAT_I"}

	var befunde []Befund
	for _, zeile := range tse.zeilen {
		tseID := tse.wert(zeile, "TSE_ID")
		for _, feld := range pflichtfelder {
			if tse.wert(zeile, feld) == "" {
				befunde = append(befunde, Befund{
					Datei:   "tse.csv",
					Regel:   regelTSEStammdaten,
					Meldung: fmt.Sprintf("TSE %q: Pflichtfeld %s ist leer", tseID, feld),
				})
			}
		}
	}
	return befunde
}

// pruefeAbrechnungskreise stellt sicher, dass jeder Tischbon seinem Abrechnungskreis
// zugeordnet ist: In allocation_groups.csv muss zu jeder Zeile ein nicht-leerer
// ABRECHNUNGSKREIS gehören und deren BON_ID auf einen existierenden Bonkopf zeigen.
// (Direktverkäufe ohne Tischbezug erscheinen gar nicht in allocation_groups.csv;
// die Regel prüft nur die vorhandenen Zeilen.)
//
// Referenz: DSFinV-K 2.4 Tz. 3.1.2.2 "Datei: Bonkopf_AbrKreis" (der Abrechnungskreis
// ordnet einen Vorgang einer variablen Einheit — hier dem Tisch — zu; F-06).
func pruefeAbrechnungskreise(daten map[string]tabellendaten) []Befund {
	allocation, ok := daten["allocation_groups.csv"]
	if !ok {
		return nil
	}
	transactions := daten["transactions.csv"]
	bekannteBons := map[string]bool{}
	for _, zeile := range transactions.zeilen {
		bekannteBons[transactions.wert(zeile, "BON_ID")] = true
	}

	var befunde []Befund
	for _, zeile := range allocation.zeilen {
		bonID := allocation.wert(zeile, "BON_ID")
		if allocation.wert(zeile, "ABRECHNUNGSKREIS") == "" {
			befunde = append(befunde, Befund{
				Datei:   "allocation_groups.csv",
				Regel:   regelAbrechnungskreis,
				Meldung: fmt.Sprintf("Bon %q hat einen leeren ABRECHNUNGSKREIS", bonID),
			})
		}
		if len(transactions.zeilen) > 0 && !bekannteBons[bonID] {
			befunde = append(befunde, Befund{
				Datei:   "allocation_groups.csv",
				Regel:   regelAbrechnungskreis,
				Meldung: fmt.Sprintf("ABRECHNUNGSKREIS-Zeile verweist auf BON_ID %q ohne Bonkopf in transactions.csv", bonID),
			})
		}
	}
	return befunde
}

// --- kleine Feldhelfer ---

// istNegativerBetrag meldet, ob ein DSFinV-K-Betragsfeld (Komma-Dezimal) negativ
// ist. Ein führendes Minus reicht als Signal; die Strukturprüfung stellt das
// Zahlenformat sicher.
func istNegativerBetrag(feld string) bool {
	return strings.HasPrefix(strings.TrimSpace(feld), "-")
}

// istNumerisch meldet, ob das Feld eine nicht-leere Folge von ASCII-Ziffern ist
// (BEDIENER_ID = user_id).
func istNumerisch(feld string) bool {
	if feld == "" {
		return false
	}
	for i := 0; i < len(feld); i++ {
		if feld[i] < '0' || feld[i] > '9' {
			return false
		}
	}
	return true
}

// sortierteSchluessel liefert die BON_IDs einer Kombi-Map deterministisch sortiert
// (für stabile Befundreihenfolge).
func sortierteSchluessel(m map[string]map[string]bool) []string {
	namen := make([]string, 0, len(m))
	for k := range m {
		namen = append(namen, k)
	}
	sort.Strings(namen)
	return namen
}
