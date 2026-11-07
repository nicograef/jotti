package dsfinvk

import (
	"sort"
)

// The tables of the DSFinV-K Kassenabschlussmodul.

var businesscasesColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"GV_TYP", "GV_NAME", "AGENTUR_ID", "UST_SCHLUESSEL",
	"Z_UMS_BRUTTO", "Z_UMS_NETTO", "Z_UST",
}

// gvTypReihenfolge orders GV types for a stable businesscases.csv (Umsatz, Pfand and
// PfandRueckzahlung before the cash movements).
var gvTypReihenfolge = map[string]int{
	gvTypUmsatz:            0,
	gvTypPfand:             1,
	gvTypPfandRueckzahlung: 2,
	gvTypAnfangsbestand:    3,
	gvTypGeldtransit:       4,
	gvTypDifferenzSollIst:  5,
}

// gvUstSchluessel is the businesscases.csv aggregation key: one GV type per VAT key.
type gvUstSchluessel struct {
	gvTyp      string
	schluessel int
}

// buildBusinesscases sums the session per GV type and VAT key (DSFinV-K Anhang C).
// It reads the same Belege as the Einzelaufzeichnungsmodul, so the day total reconciles.
func buildBusinesscases(s Snapshot, erstellung string, belege []beleg) Table {
	summen := map[gvUstSchluessel]ustBetrag{}
	for bi := range belege {
		b := &belege[bi]
		if b.geldneutral {
			continue
		}
		for _, teil := range b.ustAufteilungJeGvTyp() {
			for _, z := range teil.betraege {
				key := gvUstSchluessel{gvTyp: teil.gvTyp, schluessel: z.schluessel}
				cur := summen[key]
				cur.brutto += b.sign() * z.brutto
				cur.netto += b.sign() * z.netto
				cur.ust += b.sign() * z.ust
				summen[key] = cur
			}
		}
	}

	keys := make([]gvUstSchluessel, 0, len(summen))
	for k := range summen {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		oi, oj := ordnung(gvTypReihenfolge, keys[i].gvTyp), ordnung(gvTypReihenfolge, keys[j].gvTyp)
		if oi != oj {
			return oi < oj
		}
		if keys[i].gvTyp != keys[j].gvTyp {
			return keys[i].gvTyp < keys[j].gvTyp
		}
		return keys[i].schluessel < keys[j].schluessel
	})

	records := make([][]string, 0, len(keys))
	for _, k := range keys {
		summe := summen[k]
		records = append(records, []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			k.gvTyp, "", "0", itoa(k.schluessel),
			formatAmount(summe.brutto), formatAmount(summe.netto), formatAmount(summe.ust),
		})
	}

	return Table{
		File:        "businesscases.csv",
		LogicalName: "Z_GV_Typ",
		Description: "Aggregierte Beträge je Geschäftsvorfalltyp und Steuersatz",
		Columns:     businesscasesColumns,
		Records:     records,
	}
}

var paymentColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"ZAHLART_TYP", "ZAHLART_NAME", "Z_ZAHLART_BETRAG",
}

// buildPayment sums amounts per Zahlart (DSFinV-K Anhang D); jotti knows only Bar,
// and the cash-neutral AVBestellung contributes none.
func buildPayment(s Snapshot, erstellung string, belege []beleg) Table {
	summen := map[string]int{}
	for bi := range belege {
		b := &belege[bi]
		if b.geldneutral {
			continue
		}
		summen[b.zahlart] += b.sign() * b.bruttoCents
	}

	// Sorted for a reproducible payment.csv; with only Bar, no ranking by Zahlart is needed.
	zahlarten := make([]string, 0, len(summen))
	for z := range summen {
		zahlarten = append(zahlarten, z)
	}
	sort.Strings(zahlarten)

	records := make([][]string, 0, len(zahlarten))
	for _, z := range zahlarten {
		records = append(records, []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			z, z, formatAmount(summen[z]),
		})
	}

	return Table{
		File:        "payment.csv",
		LogicalName: "Z_Zahlart",
		Description: "Aggregierte Summen je Zahlart",
		Columns:     paymentColumns,
		Records:     records,
	}
}

var cashPerCurrencyColumns = []string{
	"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR",
	"ZAHLART_WAEH", "ZAHLART_BETRAG_WAEH",
}

// buildCashPerCurrency reports the closing cash balance per currency; jotti uses EUR only.
func buildCashPerCurrency(s Snapshot, erstellung string, belege []beleg) Table {
	record := []string{
		s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
		basiswaehrung, formatAmount(barbestand(belege)),
	}

	return Table{
		File:        "cash_per_currency.csv",
		LogicalName: "Z_Waehrungen",
		Description: "Bargeldbestand je Währung zum Abschluss",
		Columns:     cashPerCurrencyColumns,
		Records:     [][]string{record},
	}
}
