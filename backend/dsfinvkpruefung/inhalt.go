package dsfinvkpruefung

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// Content rules follow from how the DSFinV-K tables relate, not from the file form.
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

// Fixed DSFinV-K values the content rules check against (Anhang B/C/E, Anlage 2).
const (
	bonTypBeleg             = "Beleg"          // Anhang B: completed transaction (payment, return)
	bonTypSonstige          = "AVSonstige"     // Anhang B: other transaction (Tagesabschluss)
	refTypTransaktion       = "Transaktion"    // Anhang E: reference within the DSFinV-K
	gvTypUmsatz             = "Umsatz"         // Anhang C: realised revenue; separates a Storno from a cash movement
	bonStornoKein           = "0"              // BON_STORNO: no voiding; jotti uses the negative representation
	tagesabschlussName      = "Tagesabschluss" // jotti's BON_NAME for the AVSonstige closing bon
	ustSchluesselRegel      = "1"              // Anlage 2: 19 % standard rate
	ustSchluesselErmaessigt = "2"              // Anlage 2: 7 % reduced rate
)

// pruefeInhalt reads only declared, present CSVs with a matching header; the structure checks report the rest.
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

// tabellendaten is a parsed CSV addressed by column name; zeilen excludes the header.
type tabellendaten struct {
	spalten map[string]int
	zeilen  [][]string
}

// wert returns "" for an unknown column or a short row.
func (t tabellendaten) wert(zeile []string, spalte string) string {
	idx, ok := t.spalten[spalte]
	if !ok || idx >= len(zeile) {
		return ""
	}
	return zeile[idx]
}

// ladeTabellendaten skips tables with a mismatched header and rows with a wrong field count;
// the structure checks report both.
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
		if !slices.Equal(header, erwartet) {
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
				continue
			}
			daten.zeilen = append(daten.zeilen, felder)
		}
		out[tab.URL] = daten
	}
	return out
}

// pruefeStornoReferenzen requires each Storno (negative "Beleg" bon with GV_TYP "Umsatz") to carry
// BON_STORNO "0" and a "Transaktion" reference with REF_BON_ID. Negative cash outflows (Geldtransit,
// DifferenzSollIst) are no Stornos and need no reference; see docs/compliance.md §6.6.
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
		if !umsatzBons[bonID] {
			continue
		}

		if bonStorno := transactions.wert(zeile, "BON_STORNO"); bonStorno != bonStornoKein {
			befunde = append(befunde, Befund{
				Datei:   "transactions.csv",
				Regel:   regelStornoBonStorno,
				Meldung: fmt.Sprintf("Storno-Beleg %q: BON_STORNO = %q, erwartet %q (jotti nutzt die Negativdarstellung, keine Vorgangsaufhebung)", bonID, bonStorno, bonStornoKein),
			})
		}

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

// referenzenNachBonID groups references.csv rows by the referencing BON_ID.
func referenzenNachBonID(refs tabellendaten) map[string][][]string {
	out := map[string][][]string{}
	for _, zeile := range refs.zeilen {
		bonID := refs.wert(zeile, "BON_ID")
		out[bonID] = append(out[bonID], zeile)
	}
	return out
}

// umsatzBonsAusLines returns the BON_IDs with a GV_TYP "Umsatz" line: sales and Stornos, never cash movements.
func umsatzBonsAusLines(lines tabellendaten) map[string]bool {
	out := map[string]bool{}
	for _, zeile := range lines.zeilen {
		if lines.wert(zeile, "GV_TYP") == gvTypUmsatz {
			out[lines.wert(zeile, "BON_ID")] = true
		}
	}
	return out
}

func hatTransaktionsReferenz(refs tabellendaten, zeilen [][]string) bool {
	for _, zeile := range zeilen {
		if refs.wert(zeile, "REF_TYP") == refTypTransaktion && refs.wert(zeile, "REF_BON_ID") != "" {
			return true
		}
	}
	return false
}

// pruefeKombiSteueraufteilung requires a bon with 7 % and 19 % in lines_vat.csv to keep both keys
// in transactions_vat.csv. See docs/compliance.md §6.7.
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
			continue // not a Kombi bon
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

// pruefeBedienerFelder requires a non-empty BEDIENER_NAME and a numeric BEDIENER_ID (user_id) on every bon.
// See docs/compliance.md §6.4.
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

// pruefeTagesabschlussZeile: Anhang B requires BON_NAME on an AVSonstige bon; jotti's closing bon uses "Tagesabschluss".
// See docs/compliance.md §6.3.
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

// pruefeTSEStammdaten requires the tse.csv fields without which the export's signatures cannot be verified.
// See DSFinV-K 2.4 Tz. 3.2.7 (Stamm_TSE) and docs/compliance.md §6.3.
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

// pruefeAbrechnungskreise requires a non-empty ABRECHNUNGSKREIS and a known Bonkopf per allocation_groups.csv row.
// Direktverkauf bons have no row; see docs/compliance.md §6.5.
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

// istNegativerBetrag checks only the leading minus; the structure checks own the number format.
func istNegativerBetrag(feld string) bool {
	return strings.HasPrefix(strings.TrimSpace(feld), "-")
}

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

// sortierteSchluessel sorts the BON_IDs so findings come in a stable order.
func sortierteSchluessel(m map[string]map[string]bool) []string {
	namen := make([]string, 0, len(m))
	for k := range m {
		namen = append(namen, k)
	}
	sort.Strings(namen)
	return namen
}
