package dsfinvk

import (
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// The tables of the DSFinV-K Einzelaufzeichnungsmodul.

var transactionsColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"BON_ID", "BON_NR", "BON_TYP", "BON_NAME",
	"TERMINAL_ID", "BON_STORNO", "BON_START", "BON_ENDE",
	"BEDIENER_ID", "BEDIENER_NAME", "UMS_BRUTTO",
	"KUNDE_NAME", "KUNDE_ID", "KUNDE_TYP", "KUNDE_STRASSE",
	"KUNDE_PLZ", "KUNDE_ORT", "KUNDE_LAND", "KUNDE_USTID",
	"BON_NOTIZ",
}

func buildTransactions(s Snapshot, erstellung string, belege []beleg) Table {
	records := make([][]string, 0, len(belege))
	for bi := range belege {
		b := &belege[bi]
		// The cash-neutral AVBestellung carries no revenue (UMS_BRUTTO 0,00); its gross
		// appears only informatively in lines.csv.
		umsBrutto := b.sign() * b.bruttoCents
		if b.geldneutral {
			umsBrutto = 0
		}
		// BON_STORNO stays 0: jotti never voids a whole Beleg, the negative sign carries
		// the correction (docs/compliance.md §6.6).
		records = append(records, []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			b.bonID, itoa(b.bonNr), b.bonTyp, bonName(b),
			"", stornoNein, b.zeitpunkt, b.zeitpunkt,
			itoa(b.bedienerID), b.bedienerName, formatAmount(umsBrutto),
			"", "", "", "",
			"", "", "", "",
			b.notiz,
		})
	}

	return Table{
		File:        "transactions.csv",
		LogicalName: "Bonkopf",
		Description: "Ein Datensatz je Kassenbon",
		Columns:     transactionsColumns,
		Records:     records,
	}
}

var allocationGroupsColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"BON_ID", "ABRECHNUNGSKREIS",
}

// buildAllocationGroups assigns each Bon with a table to its ABRECHNUNGSKREIS
// (F-06); Belege without one (Direktverkauf) are left out.
func buildAllocationGroups(s Snapshot, erstellung string, belege []beleg) Table {
	var records [][]string
	for bi := range belege {
		b := &belege[bi]
		if b.abrechnungskreis == "" {
			continue
		}
		records = append(records, []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			b.bonID, b.abrechnungskreis,
		})
	}

	return Table{
		File:        "allocation_groups.csv",
		LogicalName: "Bonkopf_AbrKreis",
		Description: "Zuordnung Bon zu Abrechnungskreis (Tisch)",
		Columns:     allocationGroupsColumns,
		Records:     records,
	}
}

var transactionsVatColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"BON_ID", "UST_SCHLUESSEL",
	"BON_BRUTTO", "BON_NETTO", "BON_UST",
}

func buildTransactionsVat(s Snapshot, erstellung string, belege []beleg) Table {
	var records [][]string
	for bi := range belege {
		b := &belege[bi]
		if b.geldneutral {
			continue
		}
		for _, z := range b.ustAufteilung() {
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, itoa(z.schluessel),
				formatAmount(b.sign() * z.brutto), formatAmount(b.sign() * z.netto), formatAmount(b.sign() * z.ust),
			})
		}
	}

	return Table{
		File:        "transactions_vat.csv",
		LogicalName: "Bonkopf_USt",
		Description: "USt-Aufschlüsselung je Bon",
		Columns:     transactionsVatColumns,
		Records:     records,
	}
}

var datapaymentColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"BON_ID", "ZAHLART_TYP", "ZAHLART_NAME",
	"ZAHLWAEH_CODE", "ZAHLWAEH_BETRAG", "BASISWAEH_BETRAG",
}

func buildDatapayment(s Snapshot, erstellung string, belege []beleg) Table {
	var records [][]string
	for bi := range belege {
		b := &belege[bi]
		if b.geldneutral {
			continue
		}
		records = append(records, []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			b.bonID, b.zahlart, b.zahlart,
			basiswaehrung, formatAmount(b.sign() * b.bruttoCents), formatAmount(b.sign() * b.bruttoCents),
		})
	}

	return Table{
		File:        "datapayment.csv",
		LogicalName: "Bonkopf_Zahlarten",
		Description: "Zahlarten je Bon",
		Columns:     datapaymentColumns,
		Records:     records,
	}
}

var referencesColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"BON_ID", "POS_ZEILE", "REF_TYP", "REF_NAME",
	"REF_DATUM", "REF_Z_KASSE_ID", "REF_Z_NR", "REF_BON_ID",
}

// buildReferences links each Storno to its origin (Radierverbot, DSFinV-K Tz. 4.2.2)
// and each Umbuchung Zugang to its Abgang. Both sides lie in this session, so the
// REF_ closing fields repeat this session's values (docs/compliance.md §6.6).
func buildReferences(s Snapshot, erstellung string, belege []beleg) Table {
	var records [][]string
	for bi := range belege {
		b := &belege[bi]
		for _, refBonID := range b.refBonIDs {
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, "", refTypTransaktion, "",
				erstellung, s.KasseSeriennummer, itoa(s.KassensitzungNr), refBonID,
			})
		}
	}

	return Table{
		File:        "references.csv",
		LogicalName: "Bon_Referenzen",
		Description: "Referenzen auf andere Bons (Storno/Umbuchung → Ursprung)",
		Columns:     referencesColumns,
		Records:     records,
	}
}

var linesColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"BON_ID", "POS_ZEILE", "GUTSCHEIN_NR", "ARTIKELTEXT",
	"POS_TERMINAL_ID", "GV_TYP", "GV_NAME", "INHAUS",
	"P_STORNO", "AGENTUR_ID", "ART_NR", "GTIN",
	"WARENGR_ID", "WARENGR", "MENGE", "FAKTOR",
	"EINHEIT", "STK_BR",
}

func buildLines(s Snapshot, erstellung string, belege []beleg) Table {
	var records [][]string
	for bi := range belege {
		b := &belege[bi]
		if b.nichtSteuerbar {
			// Cash movement: one synthetic line (ARTIKELTEXT = GV type); MENGE ±1 carries the
			// sign, the unit price the positive magnitude.
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, "1", "", b.artikeltext,
				"", b.gvTyp, "", "",
				stornoNein, "0", "", "",
				"", "", formatQuantity(b.sign()), "",
				"", formatAmount(b.bruttoCents),
			})
			continue
		}
		for i, p := range b.positionen {
			// Cash-neutral AVBestellungen carry no GV_TYP: "Umsatz" on their lines would make
			// a per-GV_TYP sum of lines exceed the revenue the closing knows.
			posGvTyp := b.positionGvTyp(i)
			if b.geldneutral {
				posGvTyp = ""
			}
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, itoa(i + 1), "", positionText(p),
				"", posGvTyp, "", "",
				stornoNein, "0", itoa(p.VarianteID), "",
				p.Kategorie, p.Kategorie, formatQuantity(b.sign() * p.Menge), "",
				"", formatAmount(p.EinzelpreisCents),
			})
		}
	}

	return Table{
		File:        "lines.csv",
		LogicalName: "Bonpos",
		Description: "Artikelzeilen je Bon",
		Columns:     linesColumns,
		Records:     records,
	}
}

var linesVatColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"BON_ID", "POS_ZEILE", "UST_SCHLUESSEL",
	"POS_BRUTTO", "POS_NETTO", "POS_UST",
}

func buildLinesVat(s Snapshot, erstellung string, belege []beleg) Table {
	var records [][]string
	for bi := range belege {
		b := &belege[bi]
		if b.geldneutral {
			continue
		}
		if b.nichtSteuerbar {
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, "1", itoa(ustNichtSteuerbar),
				formatAmount(b.sign() * b.bruttoCents), formatAmount(b.sign() * b.bruttoCents), formatAmount(0),
			})
			continue
		}
		for i, p := range b.positionen {
			brutto := p.EinzelpreisCents * p.Menge
			for _, aufteilung := range steuer.Aufteilen(brutto, steuer.Steuersatz(p.Steuersatz)) {
				records = append(records, []string{
					s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
					b.bonID, itoa(i + 1), itoa(ustSchluessel(aufteilung.Satz)),
					formatAmount(b.sign() * aufteilung.Brutto), formatAmount(b.sign() * aufteilung.Netto), formatAmount(b.sign() * aufteilung.Steuer),
				})
			}
		}
	}

	return Table{
		File:        "lines_vat.csv",
		LogicalName: "Bonpos_USt",
		Description: "USt-Aufschlüsselung je Artikelzeile",
		Columns:     linesVatColumns,
		Records:     records,
	}
}

// Declared tables jotti never fills ship as header-only CSV, since the official
// index.xml declares all 20 files and audit software expects them (docs/compliance.md §6.3).
var slavesColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"TERMINAL_ID", "TERMINAL_BRAND", "TERMINAL_MODELL",
	"TERMINAL_SERIENNR", "TERMINAL_SW_BRAND", "TERMINAL_SW_VERSION",
}

var paColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"AGENTUR_ID", "AGENTUR_NAME", "AGENTUR_STRASSE",
	"AGENTUR_PLZ", "AGENTUR_ORT", "AGENTUR_LAND",
	"AGENTUR_STNR", "AGENTUR_USTID",
}

var itemamountsColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"BON_ID", "POS_ZEILE", "TYP",
	"UST_SCHLUESSEL", "PF_BRUTTO", "PF_NETTO", "PF_UST",
}

var subitemsColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"BON_ID", "POS_ZEILE", "ZI_ART_NR",
	"ZI_GTIN", "ZI_NAME", "ZI_WARENGR_ID",
	"ZI_WARENGR", "ZI_MENGE", "ZI_FAKTOR",
	"ZI_EINHEIT", "ZI_UST_SCHLUESSEL",
	"ZI_BASISPREIS_BRUTTO", "ZI_BASISPREIS_NETTO", "ZI_BASISPREIS_UST",
}

func headerOnlyTable(file, logicalName, description string, columns []string) Table {
	return Table{File: file, LogicalName: logicalName, Description: description, Columns: columns, Records: nil}
}

var transactionsTSEColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"BON_ID", "TSE_ID", "TSE_TANR",
	"TSE_TA_START", "TSE_TA_ENDE", "TSE_TA_VORGANGSART",
	"TSE_TA_SIGZ", "TSE_TA_SIG", "TSE_TA_FEHLER",
	"TSE_VORGANGSDATEN",
}

func buildTransactionsTSE(s Snapshot, erstellung string, belege []beleg) Table {
	var records [][]string
	for bi := range belege {
		b := &belege[bi]
		switch {
		case b.tse != nil:
			// TSE_VORGANGSDATEN stays empty: optional in the spec, and the signed processData
			// is not reconstructed here.
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, tseReferenzID, itoa(b.tse.TransaktionNummer),
				isoZeit(b.tse.LogTimeStart), isoZeit(b.tse.LogTimeEnd), b.processType,
				itoa(b.tse.SignaturZaehler), b.tse.Signatur, "",
				"",
			})
		case b.tsePflichtig:
			// An unsigned Vorgang that requires a signature gets a TSE_TA_FEHLER row naming its
			// Auftrag status (docs/compliance.md §3.8). Vorgänge without signature duty get no row.
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, tseReferenzID, "",
				"", "", "",
				"", "", tseFehlertext(b.tseStatus),
				"",
			})
		}
	}

	return Table{
		File:        "transactions_tse.csv",
		LogicalName: "TSE_Transaktionen",
		Description: "TSE-Transaktionsdaten je Bon",
		Columns:     transactionsTSEColumns,
		Records:     records,
	}
}

// tseFehlertext names why an unsigned Vorgang has no signature; fehlgeschlagen is the outage case.
func tseFehlertext(status string) string {
	switch status {
	case tse.StatusOffen:
		return tseFehlerAusstehend
	case tse.StatusTSENichtKonfiguriert:
		return tseFehlerKeineTSE
	default:
		return tseFehlerAusfall
	}
}
