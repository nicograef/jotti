package dsfinvk

import (
	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// The tables of the DSFinV-K Stammdatenmodul.

var cashpointclosingColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"Z_BUCHUNGSTAG", "TAXONOMIE_VERSION",
	"Z_START_ID", "Z_ENDE_ID",
	"NAME", "STRASSE", "PLZ", "ORT", "LAND",
	"STNR", "USTID",
	"Z_SE_ZAHLUNGEN", "Z_SE_BARZAHLUNGEN",
}

func buildCashpointclosing(s Snapshot, erstellung string, belege []beleg) Table {
	bar := barbestand(belege)

	// Z_BUCHUNGSTAG stays empty: the spec uses it only for a booking day that differs
	// from Z_ERSTELLUNG, and jotti books on the creation day.
	record := []string{
		s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
		"", Version,
		belege[0].bonID, belege[len(belege)-1].bonID,
		truncateRunes(s.Betreiber.Vereinsname, betreiber.MaxLengthVereinsname), truncateRunes(s.Betreiber.Strasse, betreiber.MaxLengthStrasse),
		truncateRunes(s.Betreiber.Plz, betreiber.MaxLengthPlz), truncateRunes(s.Betreiber.Ort, betreiber.MaxLengthOrt), land,
		derefOrEmpty(s.Betreiber.Steuernummer), derefOrEmpty(s.Betreiber.UstID),
		formatAmount(bar), formatAmount(bar),
	}

	return Table{
		File:        "cashpointclosing.csv",
		LogicalName: "Stamm_Abschluss",
		Description: "Metadaten zum Kassenabschluss (Z-Bon)",
		Columns:     cashpointclosingColumns,
		Records:     [][]string{record},
	}
}

var locationColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"LOC_NAME", "LOC_STRASSE", "LOC_PLZ", "LOC_ORT",
	"LOC_LAND", "LOC_USTID",
}

func buildLocation(s Snapshot, erstellung string) Table {
	record := []string{
		s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
		truncateRunes(s.Betreiber.Vereinsname, betreiber.MaxLengthVereinsname), truncateRunes(s.Betreiber.Strasse, betreiber.MaxLengthStrasse),
		truncateRunes(s.Betreiber.Plz, betreiber.MaxLengthPlz), truncateRunes(s.Betreiber.Ort, betreiber.MaxLengthOrt),
		land, derefOrEmpty(s.Betreiber.UstID),
	}

	return Table{
		File:        "location.csv",
		LogicalName: "Stamm_Orte",
		Description: "Betriebsstätte des Betreibers",
		Columns:     locationColumns,
		Records:     [][]string{record},
	}
}

var cashregisterColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"KASSE_BRAND", "KASSE_MODELL", "KASSE_SERIENNR",
	"KASSE_SW_BRAND", "KASSE_SW_VERSION",
	"KASSE_BASISWAEH_CODE", "KEINE_UST_ZUORDNUNG",
}

func buildCashregister(s Snapshot, erstellung string) Table {
	record := []string{
		s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
		kasseBrand, kasseModell, s.KasseSeriennummer,
		kasseSoftware, s.SoftwareVersion,
		basiswaehrung, "",
	}

	return Table{
		File:        "cashregister.csv",
		LogicalName: "Stamm_Kassen",
		Description: "Seriennummer, Software-Typ und -Version der Kasse",
		Columns:     cashregisterColumns,
		Records:     [][]string{record},
	}
}

var vatColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"UST_SCHLUESSEL", "UST_SATZ", "UST_BESCHR",
}

func buildVat(s Snapshot, erstellung string) Table {
	// vat.csv lists DSFinV-K Anlage 2 keys 1-7, not only those used, as audit software
	// expects (docs/compliance.md §6.7). UST_SATZ is the rate valid at recording for
	// keys 1-4 and the fixed 0,00 for 5-7.
	amtlicheSchluessel := [][2]string{
		{"19,00", "Allgemeiner Steuersatz"},
		{"7,00", "Ermäßigter Steuersatz"},
		{"7,80", "Durchschnittsatz (§ 24 Abs. 1 Nr. 3 UStG)"},
		{"5,50", "Durchschnittsatz (§ 24 Abs. 1 Nr. 1 UStG)"},
		{"0,00", "Nicht Steuerbar"},
		{"0,00", "Umsatzsteuerfrei"},
		{"0,00", "UmsatzsteuerNichtErmittelbar"},
	}

	records := make([][]string, 0, len(amtlicheSchluessel))
	for i, eintrag := range amtlicheSchluessel {
		records = append(records, []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			itoa(i + 1), eintrag[0], eintrag[1],
		})
	}

	return Table{
		File:        "vat.csv",
		LogicalName: "Stamm_USt",
		Description: "Verwendete Umsatzsteuersätze",
		Columns:     vatColumns,
		Records:     records,
	}
}

var tseColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"TSE_ID", "TSE_SERIAL", "TSE_SIG_ALGO",
	"TSE_ZEITFORMAT", "TSE_PD_ENCODING", "TSE_PUBLIC_KEY",
	"TSE_ZERTIFIKAT_I", "TSE_ZERTIFIKAT_II",
}

// tseVerzeichnis lists the TSS of one export: tse.csv gets one row per TSS, and TSE_ID is
// the row's position.
type tseVerzeichnis struct {
	stammdaten []tse.Stammdaten
	ids        map[string]string // TSS serial -> TSE_ID
	aktuelleID string            // TSE_ID of the unsigned Vorgänge, which the current TSS will sign
}

// tseVerzeichnisAus lists each TSS that signed a Beleg, in order of first signature, and the
// current TSS when a Vorgang is still unsigned or nothing is signed. A serial without stored
// master data keeps its serial and leaves the rest empty, which the export check reports.
func tseVerzeichnisAus(s Snapshot, belege []beleg) tseVerzeichnis {
	v := tseVerzeichnis{ids: map[string]string{}}
	aufnehmen := func(stammdaten tse.Stammdaten) string {
		if id, ok := v.ids[stammdaten.Seriennummer]; ok {
			return id
		}
		v.stammdaten = append(v.stammdaten, stammdaten)
		id := itoa(len(v.stammdaten))
		v.ids[stammdaten.Seriennummer] = id
		return id
	}

	unsigniert := false
	for bi := range belege {
		b := &belege[bi]
		switch {
		case b.tse != nil:
			stammdaten, ok := s.TSEJeSeriennummer[b.tse.TSESeriennummer]
			if !ok {
				stammdaten = tse.Stammdaten{Seriennummer: b.tse.TSESeriennummer}
			}
			aufnehmen(stammdaten)
		case b.tsePflichtig:
			unsigniert = true
		}
	}
	if unsigniert || len(v.stammdaten) == 0 {
		v.aktuelleID = aufnehmen(s.AktuelleTSE)
	}
	return v
}

func buildTSE(s Snapshot, erstellung string, tses tseVerzeichnis) Table {
	records := make([][]string, 0, len(tses.stammdaten))
	for _, stammdaten := range tses.stammdaten {
		// TSE_ZEITFORMAT declares the TSE's own log time format (fiskaly: unixTime) from the
		// master data stored at setup; TSE_TA_START/ENDE are ISO 8601 regardless.
		zeitformat := stammdaten.LogTimeFormat
		if zeitformat == "" {
			zeitformat = defaultTSEZeitformat
		}

		record := []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			tses.ids[stammdaten.Seriennummer], stammdaten.Seriennummer, stammdaten.SignaturAlgorithmus,
			zeitformat, tsePDEncoding, stammdaten.PublicKey,
		}
		for i := range zertifikatSpalten {
			record = append(record, certChunk(stammdaten.Zertifikat, i))
		}
		records = append(records, record)
	}

	return Table{
		File:        "tse.csv",
		LogicalName: "Stamm_TSE",
		Description: "Stammdaten der technischen Sicherheitseinrichtung",
		Columns:     tseColumns,
		Records:     records,
	}
}
