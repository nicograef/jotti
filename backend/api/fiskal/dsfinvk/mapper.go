package dsfinvk

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// ErrKeineVorgaenge marks a Kassensitzung without a single Bon; the caller
// reports it instead of shipping an empty archive.
var ErrKeineVorgaenge = errors.New("kassensitzung enthält keine vorgänge")

const (
	bonTypBeleg      = "Beleg"        // Anhang B: completed Kassenvorgang (payment)
	bonTypBestellung = "AVBestellung" // Anhang B: order as "anderer Vorgang", cash-neutral
	bonTypSonstige   = "AVSonstige"   // Anhang B: other "anderer Vorgang" (Tagesabschluss), cash-neutral
	gvTypUmsatz      = "Umsatz"       // Anhang C: realised revenue at line level
	// Anhang C: the Pfand line of a sale and its Warenrücknahme. Only the v2 sale events carry the
	// Pfand flag per position; v1 sale lines stay "Umsatz" (docs/decisions.md D20).
	gvTypPfand = "Pfand"
	// Anhang C: Pfand paid out to the guest, a negative Entgeltminderung at the Pfand
	// product's VAT rate (Tz. 4.2.5).
	gvTypPfandRueckzahlung = "PfandRueckzahlung"
	// Anhang C GV types that only move the cash balance; jotti records them as
	// BON_TYP "Beleg" with one non-taxable line (see docs/compliance.md §6.8).
	gvTypAnfangsbestand   = "Anfangsbestand"   // cash at session start (opening event)
	gvTypGeldtransit      = "Geldtransit"      // cash deposit or withdrawal (e.g. bank, safe)
	gvTypDifferenzSollIst = "DifferenzSollIst" // booked cash difference from the Kassensturz
	zahlartBar            = "Bar"              // Anhang D: jotti takes cash only
	refTypTransaktion     = "Transaktion"      // Anhang E: REF_TYP for a reference inside the DSFinV-K (Storno → origin)
	tseReferenzID         = "1"                // one TSS per Kasse, referenced as ID 1 in the closing
	tseFehlerAusfall      = "TSE-Ausfall"      // TSE_TA_FEHLER of an unsigned outage Vorgang (not yet re-signed)
	land                  = "DEU"              // ISO 3166 ALPHA-3
	basiswaehrung         = "EUR"              // ISO 4217
	tsePDEncoding         = "UTF-8"            // encoding of the ProcessData
	zertifikatChunk       = 1000               // max characters per TSE_ZERTIFIKAT field (the spec has two)
	zertifikatSpalten     = 2                  // TSE_ZERTIFIKAT_I/_II of the official DSFinV-K schema
	// maxLengthAbrechnungskreis is the official ABRECHNUNGSKREIS length in
	// allocation_groups.csv (index.xml), counted in characters.
	maxLengthAbrechnungskreis = 50
	defaultTSEZeitformat      = "unixTime" // fiskaly logs unixTime; fallback without TSE master data
	kasseBrand                = "jotti"
	kasseModell               = "jotti mPOS"
	kasseSoftware             = "jotti"
)

// Archive holds a DSFinV-K export as one Table per CSV file, in the order of
// the official index.xml.
type Archive struct {
	tables []Table
}

func (a Archive) Tables() []Table { return a.tables }

// beleg is the per-Bon view that several tables derive from (Bonkopf, its VAT,
// payment and line details, and the TSE row).
type beleg struct {
	bonID            string
	bonNr            int
	bonTyp           string   // BON_TYP: "Beleg", "AVBestellung" or "AVSonstige"
	gvTyp            string   // GV_TYP of the lines: "Umsatz", "PfandRueckzahlung" or a cash GV type; pfand overrides it per line
	zahlart          string   // ZAHLART_TYP: "Bar"; empty for the cash-neutral AVBestellung
	abrechnungskreis string   // ABRECHNUNGSKREIS (table name); empty without a table (e.g. Direktverkauf)
	storno           bool     // negative Beleg (Warenrücknahme, Korrektur, Pfand return): flips the sign; BON_STORNO stays 0
	barabfluss       bool     // reduces the cash balance (Geldtransit withdrawal, shortfall); flips the sign like storno
	geldneutral      bool     // AVBestellung or AVSonstige: TSE-signed, lines informative only; no revenue, VAT, Zahlart or cash effect
	nichtSteuerbar   bool     // cash movement without VAT: one line with UST_SCHLUESSEL 5 instead of a tax split
	artikeltext      string   // ARTIKELTEXT of the synthetic line (nichtSteuerbar only)
	refBonIDs        []string // REF_BON_ID per origin Bon (Warenrücknahme → Zahlung, Korrektur → Bestellung, Umbuchung Zugang → Abgang, Pfand-Storno → Rückgabe)
	zeitpunkt        string   // BON_START and BON_ENDE: a Vorgang is one event, so both are its time
	bedienerID       int
	bedienerName     string
	positionen       []kasse.PositionEventData
	pfand            []bool // per position: GV_TYP "Pfand" instead of gvTyp; nil when no line is Pfand
	bruttoCents      int
	tsePflichtig     bool          // signature required (a Signaturauftrag exists for the event)
	tse              *tse.Signatur // signature from the Auftrag; nil while unsigned
	processType      string        // TSE_TA_VORGANGSART, the Auftrag's process_type snapshot
	notiz            string
}

// sign is -1 for a negative Beleg (DSFinV-K Tz. 4.2.5) or a cash outflow, else +1;
// beleg amounts stay positive because the tax split needs non-negative input.
func (b *beleg) sign() int {
	if b.storno || b.barabfluss {
		return -1
	}
	return 1
}

// ustBetrag is a Beleg's VAT split for one VAT key as a positive magnitude; the
// caller applies beleg.sign().
type ustBetrag struct {
	schluessel int
	brutto     int
	netto      int
	ust        int
}

// ustAufteilung splits the Beleg per VAT key; a non-taxable cash movement yields
// one UST_SCHLUESSEL 5 row with net = gross. transactions_vat.csv and
// businesscases.csv both derive from it, so their sums agree.
func (b *beleg) ustAufteilung() []ustBetrag {
	if b.nichtSteuerbar {
		return []ustBetrag{{schluessel: ustNichtSteuerbar, brutto: b.bruttoCents, netto: b.bruttoCents, ust: 0}}
	}
	return ustBetraege(b.positionen)
}

// positionGvTyp is the GV_TYP of line i.
func (b *beleg) positionGvTyp(i int) string {
	if i < len(b.pfand) && b.pfand[i] {
		return gvTypPfand
	}
	return b.gvTyp
}

// gvUstBetrag is one GV type's VAT split within a Beleg.
type gvUstBetrag struct {
	gvTyp    string
	betraege []ustBetrag
}

// ustAufteilungJeGvTyp splits the Beleg per line GV type, then per VAT key. Steuermatrix sums
// per position, so the parts add up to ustAufteilung.
func (b *beleg) ustAufteilungJeGvTyp() []gvUstBetrag {
	if b.nichtSteuerbar || b.pfand == nil {
		return []gvUstBetrag{{gvTyp: b.gvTyp, betraege: b.ustAufteilung()}}
	}
	var out []gvUstBetrag
	for _, gvTyp := range []string{b.gvTyp, gvTypPfand} {
		var positionen []kasse.PositionEventData
		for i, p := range b.positionen {
			if b.positionGvTyp(i) == gvTyp {
				positionen = append(positionen, p)
			}
		}
		if len(positionen) > 0 {
			out = append(out, gvUstBetrag{gvTyp: gvTyp, betraege: ustBetraege(positionen)})
		}
	}
	return out
}

func ustBetraege(positionen []kasse.PositionEventData) []ustBetrag {
	matrix := steuer.Steuermatrix(kasse.SteuermatrixPositionen(kasse.PositionenFromEventData(positionen)))
	out := make([]ustBetrag, 0, len(matrix))
	for _, a := range matrix {
		out = append(out, ustBetrag{schluessel: ustSchluessel(a.Satz), brutto: a.Brutto, netto: a.Netto, ust: a.Steuer})
	}
	return out
}

// pfandJePosition maps each PositionID of a v2 sale to its Pfand flag; a v1 sale yields none.
func pfandJePosition(positionen []kasse.PositionV2EventData) map[string]bool {
	out := make(map[string]bool, len(positionen))
	for _, p := range positionen {
		if p.Pfand {
			out[p.PositionID] = true
		}
	}
	return out
}

// verkaufsPositionen splits v2 sale positions into the v1 form and the Pfand flags (nil if none).
func verkaufsPositionen(positionen []kasse.PositionV2EventData) ([]kasse.PositionEventData, []bool) {
	basis := make([]kasse.PositionEventData, len(positionen))
	var pfand []bool
	for i, p := range positionen {
		basis[i] = p.PositionEventData
		if p.Pfand {
			if pfand == nil {
				pfand = make([]bool, len(positionen))
			}
			pfand[i] = true
		}
	}
	return basis, pfand
}

// ruecknahmePfand derives a Warenrücknahme's Pfand flags from its sale's positions (nil if none).
func ruecknahmePfand(positionen []kasse.PositionEventData, verkauf map[string]bool) []bool {
	var pfand []bool
	for i, p := range positionen {
		if verkauf[p.PositionID] {
			if pfand == nil {
				pfand = make([]bool, len(positionen))
			}
			pfand[i] = true
		}
	}
	return pfand
}

// Map turns a Kassensitzung's snapshot and events into the archive; revenue arises
// at payment (DSFinV-K Tz. 2.7.2, see docs/compliance.md §6.8). signaturen maps
// event ID to signature state, and an event without entry needs no signature.
func Map(snapshot Snapshot, events []event.Event, signaturen map[int]tse.EventSignatur) (Archive, error) {
	erstellung := snapshot.Erstellung.UTC().Format(time.RFC3339)

	belege, err := belegeFromEvents(events, snapshot.Tischnamen, signaturen)
	if err != nil {
		return Archive{}, err
	}
	if len(belege) == 0 {
		return Archive{}, ErrKeineVorgaenge
	}

	// All 20 files the official index.xml declares, in its order; unfilled ones as
	// header-only CSV.
	tables := []Table{
		buildCashpointclosing(snapshot, erstellung, belege),
		buildLocation(snapshot, erstellung),
		buildCashregister(snapshot, erstellung),
		headerOnlyTable("slaves.csv", "Stamm_Terminals", "Terminal-Kassen (in jotti nicht vorhanden)", slavesColumns),
		headerOnlyTable("pa.csv", "Stamm_Agenturen", "Agenturgeschäft (in jotti nicht vorhanden)", paColumns),
		buildTSE(snapshot, erstellung),
		buildVat(snapshot, erstellung),
		buildBusinesscases(snapshot, erstellung, belege),
		buildPayment(snapshot, erstellung, belege),
		buildCashPerCurrency(snapshot, erstellung, belege),
		buildTransactions(snapshot, erstellung, belege),
		buildDatapayment(snapshot, erstellung, belege),
		buildLines(snapshot, erstellung, belege),
		headerOnlyTable("itemamounts.csv", "Bonpos_Preisfindung", "Preisfindung je Position (in jotti nicht vorhanden)", itemamountsColumns),
		headerOnlyTable("subitems.csv", "Bonpos_Zusatzinfo", "Zusatzinformationen je Position (in jotti nicht vorhanden)", subitemsColumns),
		buildTransactionsTSE(snapshot, erstellung, belege),
		buildTransactionsVat(snapshot, erstellung, belege),
		buildLinesVat(snapshot, erstellung, belege),
		buildAllocationGroups(snapshot, erstellung, belege),
		buildReferences(snapshot, erstellung, belege),
	}

	return Archive{tables: tables}, nil
}

// belegeFromEvents derives the Belege from events ordered by id, so an origin Bon
// always precedes its Storno. BON_NR is the kassenjournal event ID, the Bon-Nr the printed
// Kassenbeleg shows; BON_ID is the Vorgang ID.
func belegeFromEvents(events []event.Event, tischnamen map[int]string, signaturen map[int]tse.EventSignatur) ([]beleg, error) {
	var belege []beleg
	// herkunft maps each PositionID to its order's BON_ID. PositionIDs are unique per
	// order, so a Korrektur resolves its origin orders via its positions.
	herkunft := map[string]string{}
	// verkaufsPfand maps a sale's BON_ID to its Pfand positions; its Warenrücknahme takes the
	// flags from there, since the frozen Storno events carry none.
	verkaufsPfand := map[string]map[string]bool{}

	for _, ev := range events {
		vorher := len(belege)
		switch kasse.EventType(ev.Type) {
		case kasse.EventTypeBestellungAufgenommenV1:
			data, err := event.Decode[kasse.BestellungAufgenommenV1Data](ev)
			if err != nil {
				return nil, err
			}
			for _, p := range data.Positionen {
				herkunft[p.PositionID] = data.BestellungID
			}
			belege = append(belege, beleg{
				bonID:            data.BestellungID,
				bonTyp:           bonTypBestellung,
				gvTyp:            gvTypUmsatz,
				geldneutral:      true,
				abrechnungskreis: abrechnungskreis(ev.Subject, tischnamen),
				positionen:       data.Positionen,
				bruttoCents:      data.GesamtPreisCents,
				notiz:            data.Kommentar,
			})

		case kasse.EventTypeZahlungKassiertV1, kasse.EventTypeZahlungKassiertV2:
			// v1 has no pfand key, so its positions read as no Pfand.
			data, err := event.Decode[kasse.ZahlungKassiertV2Data](ev)
			if err != nil {
				return nil, err
			}
			positionen, pfand := verkaufsPositionen(data.Positionen)
			verkaufsPfand[data.ZahlungID] = pfandJePosition(data.Positionen)
			belege = append(belege, beleg{
				bonID:            data.ZahlungID,
				bonTyp:           bonTypBeleg,
				gvTyp:            gvTypUmsatz,
				zahlart:          zahlartBar,
				abrechnungskreis: abrechnungskreis(ev.Subject, tischnamen),
				positionen:       positionen,
				pfand:            pfand,
				bruttoCents:      data.GesamtZahlungCents,
				notiz:            data.Kommentar,
			})

		case kasse.EventTypeStornierungErteiltV1:
			data, err := event.Decode[kasse.StornierungErteiltV1Data](ev)
			if err != nil {
				return nil, err
			}
			// Warenrücknahme of paid lines: negative revenue at the original VAT rate with cash
			// refund (DSFinV-K Tz. 4.2.5), referencing the Zahlung (Tz. 4.2.2, docs/compliance.md §6.6).
			belege = append(belege, beleg{
				bonID:            data.StornierungID,
				bonTyp:           bonTypBeleg,
				gvTyp:            gvTypUmsatz,
				zahlart:          zahlartBar,
				storno:           true,
				abrechnungskreis: abrechnungskreis(ev.Subject, tischnamen),
				refBonIDs:        []string{data.ZahlungID},
				positionen:       data.Positionen,
				pfand:            ruecknahmePfand(data.Positionen, verkaufsPfand[data.ZahlungID]),
				bruttoCents:      data.GesamtStornierungCents,
				notiz:            data.Kommentar,
			})

		case kasse.EventTypeBestellungKorrigiertV1:
			data, err := event.Decode[kasse.BestellungKorrigiertV1Data](ev)
			if err != nil {
				return nil, err
			}
			// Cash-neutral cancellation of unpaid lines: an AVBestellung without revenue,
			// Zahlart or cash effect, referencing the origin order.
			belege = append(belege, beleg{
				bonID:            data.KorrekturID,
				bonTyp:           bonTypBestellung,
				gvTyp:            gvTypUmsatz,
				geldneutral:      true,
				abrechnungskreis: abrechnungskreis(ev.Subject, tischnamen),
				storno:           true,
				refBonIDs:        ursprungsbons(data.Positionen, herkunft),
				positionen:       data.Positionen,
				bruttoCents:      data.GesamtCents,
				notiz:            data.Kommentar,
			})

		case kasse.EventTypeBestellungUmgebuchtV1:
			data, err := event.Decode[kasse.BestellungUmgebuchtV1Data](ev)
			if err != nil {
				return nil, err
			}
			tischID, err := kasse.ParseTischIDFromSubject(ev.Subject)
			if err != nil {
				return nil, fmt.Errorf("parse tisch from bestellung-umgebucht (event %d): %w", ev.ID, err)
			}
			// An Umbuchung is a cash-neutral AVBestellung. Source and target share the
			// UmbuchungID: the Abgang carries it as BON_ID, the Zugang references it.
			bon := beleg{
				bonTyp:           bonTypBestellung,
				gvTyp:            gvTypUmsatz,
				geldneutral:      true,
				abrechnungskreis: abrechnungskreis(ev.Subject, tischnamen),
				positionen:       data.Positionen,
				bruttoCents:      data.GesamtCents,
				notiz:            umbuchungNotiz(data.Kommentar, data.BenutzerKommentar),
			}
			if tischID == data.QuellTischID {
				bon.bonID = data.UmbuchungID
			} else {
				bon.bonID = fmt.Sprintf("umbuchung-%d", ev.ID)
				bon.refBonIDs = []string{data.UmbuchungID}
				for _, p := range data.Positionen {
					herkunft[p.PositionID] = bon.bonID
				}
			}
			belege = append(belege, bon)

		case kasse.EventTypeDirektverkaufGetaetigtV1, kasse.EventTypeDirektverkaufGetaetigtV2:
			// v1 has no pfand key, so its positions read as no Pfand.
			data, err := event.Decode[kasse.DirektverkaufGetaetigtV2Data](ev)
			if err != nil {
				return nil, err
			}
			positionen, pfand := verkaufsPositionen(data.Positionen)
			verkaufsPfand[data.VerkaufID] = pfandJePosition(data.Positionen)
			belege = append(belege, beleg{
				bonID:       data.VerkaufID,
				bonTyp:      bonTypBeleg,
				gvTyp:       gvTypUmsatz,
				zahlart:     zahlartBar,
				positionen:  positionen,
				pfand:       pfand,
				bruttoCents: data.GesamtbetragCents,
				notiz:       data.Kommentar,
			})

		case kasse.EventTypeDirektverkaufStorniertV1:
			data, err := event.Decode[kasse.DirektverkaufStorniertV1Data](ev)
			if err != nil {
				return nil, err
			}
			belege = append(belege, beleg{
				bonID:       data.StornierungID,
				bonTyp:      bonTypBeleg,
				gvTyp:       gvTypUmsatz,
				zahlart:     zahlartBar,
				storno:      true,
				refBonIDs:   []string{data.VerkaufID},
				positionen:  data.Positionen,
				pfand:       ruecknahmePfand(data.Positionen, verkaufsPfand[data.VerkaufID]),
				bruttoCents: data.GesamtStornierungCents,
				notiz:       data.Kommentar,
			})

		case kasse.EventTypePfandZurueckgegebenV1:
			data, err := event.Decode[kasse.PfandZurueckgegebenV1Data](ev)
			if err != nil {
				return nil, err
			}
			// Pfand paid out in cash: a negative Beleg without reference, since a Pfand return
			// reduces the consideration independently of the sale (Abschn. 10.1 Abs. 8 Satz 3 UStAE).
			belege = append(belege, beleg{
				bonID:            data.RueckgabeID,
				bonTyp:           bonTypBeleg,
				gvTyp:            gvTypPfandRueckzahlung,
				zahlart:          zahlartBar,
				storno:           true,
				abrechnungskreis: abrechnungskreisFuerTisch(data.TischID, tischnamen),
				positionen:       data.Positionen,
				bruttoCents:      data.GesamtbetragCents,
				notiz:            data.Kommentar,
			})

		case kasse.EventTypePfandrueckgabeStorniertV1:
			data, err := event.Decode[kasse.PfandrueckgabeStorniertV1Data](ev)
			if err != nil {
				return nil, err
			}
			// Cancelling a Pfand return is a positive Beleg referencing the Rückgabe.
			belege = append(belege, beleg{
				bonID:            data.StornierungID,
				bonTyp:           bonTypBeleg,
				gvTyp:            gvTypPfandRueckzahlung,
				zahlart:          zahlartBar,
				abrechnungskreis: abrechnungskreisFuerTisch(data.TischID, tischnamen),
				refBonIDs:        []string{data.RueckgabeID},
				positionen:       data.Positionen,
				bruttoCents:      data.GesamtStornierungCents,
				notiz:            data.Kommentar,
			})

		case kasse.EventTypeKassensitzungEroeffnetV1:
			data, err := event.Decode[kasse.KassensitzungEroeffnetV1Data](ev)
			if err != nil {
				return nil, err
			}
			// Without cash at session start there is no Anfangsbestand to record
			// (Anhang C: recording not mandatory).
			if data.BetragCents == 0 {
				continue
			}
			belege = append(belege, geldbewegung(fmt.Sprintf("anfangsbestand-%d", ev.ID), gvTypAnfangsbestand, data.BetragCents, false, data.Bezeichnung))

		case kasse.EventTypeGeldtransitGebuchtV1:
			data, err := event.Decode[kasse.GeldtransitGebuchtV1Data](ev)
			if err != nil {
				return nil, err
			}
			belege = append(belege, geldbewegung(data.GeldtransitID, gvTypGeldtransit, data.BetragCents, data.Richtung == "entnahme", data.Kommentar))

		case kasse.EventTypeDifferenzSollIstGebuchtV1:
			data, err := event.Decode[kasse.DifferenzSollIstGebuchtV1Data](ev)
			if err != nil {
				return nil, err
			}
			// BetragCents = Soll − Ist: positive is a shortfall (cash missing, reduce the
			// balance), negative a surplus.
			belege = append(belege, geldbewegung(fmt.Sprintf("differenz-soll-ist-%d", ev.ID), gvTypDifferenzSollIst, abs(data.BetragCents), data.BetragCents > 0, ""))

		case kasse.EventTypeKassensturzDurchgefuehrtV1:
			// The count itself moves no cash; its Differenz is booked as its own event.

		case kasse.EventTypeTagesabschlussErstelltV1:
			if _, err := event.Decode[kasse.TagesabschlussErstelltV1Data](ev); err != nil {
				return nil, err
			}
			// The Tagesabschluss is a TSE-signed, cash-neutral AVSonstige without lines. It
			// exists in the export so its signature gets a transactions_tse.csv row (docs/compliance.md §6.8).
			belege = append(belege, beleg{
				bonID:       fmt.Sprintf("tagesabschluss-%d", ev.ID),
				bonTyp:      bonTypSonstige,
				geldneutral: true,
			})
		}
		// Each Beleg of this iteration (at most one per event) takes its event's ID as BON_NR, carries
		// its event's time and operator, and takes its event's TSE state from the Signaturauftrag
		// table, the single source for signature duty, signature and TSE_TA_VORGANGSART.
		sig, pflichtig := signaturen[ev.ID]
		for i := vorher; i < len(belege); i++ {
			belege[i].bonNr = ev.ID
			belege[i].zeitpunkt = zeit(ev)
			belege[i].bedienerID = ev.UserID
			belege[i].bedienerName = ev.UserName
			belege[i].tsePflichtig = pflichtig
			belege[i].tse = sig.Signatur
			belege[i].processType = sig.ProcessType
		}
	}

	return belege, nil
}

// geldbewegung builds a non-taxable cash movement: BON_TYP "Beleg" with one line
// (ARTIKELTEXT = GV type, UST_SCHLUESSEL 5; DSFinV-K Anhang C, docs/compliance.md §6.8).
// betragCents is the positive magnitude; barabfluss sets the sign.
func geldbewegung(bonID string, gvTyp string, betragCents int, barabfluss bool, notiz string) beleg {
	return beleg{
		bonID:          bonID,
		bonTyp:         bonTypBeleg,
		gvTyp:          gvTyp,
		zahlart:        zahlartBar,
		barabfluss:     barabfluss,
		nichtSteuerbar: true,
		artikeltext:    gvTyp,
		bruttoCents:    betragCents,
		notiz:          notiz,
	}
}

// ursprungsbons returns the BON_IDs of the orders the cancelled lines came from,
// deduplicated in order of first appearance. A Korrektur across several order
// rounds references several, one references.csv row each.
func ursprungsbons(positionen []kasse.PositionEventData, herkunft map[string]string) []string {
	var bons []string
	gesehen := map[string]bool{}
	for _, p := range positionen {
		bon, ok := herkunft[p.PositionID]
		if !ok || gesehen[bon] {
			continue
		}
		gesehen[bon] = true
		bons = append(bons, bon)
	}
	return bons
}

// umbuchungNotiz builds BON_NOTIZ from the direction autotext and the optional user
// comment joined by "; " (at most 202 of the 255 allowed characters, so no truncation).
func umbuchungNotiz(autotext string, benutzerKommentar string) string {
	if benutzerKommentar == "" {
		return autotext
	}
	return autotext + "; " + benutzerKommentar
}

// zeit formats the event time as ISO 8601 UTC for BON_START/BON_ENDE.
func zeit(ev event.Event) string { return ev.Time.UTC().Format(time.RFC3339) }

// isoZeit formats a TSE logTime for TSE_TA_START/ENDE, which the spec requires as
// ISO 8601 with milliseconds. fiskaly logs whole seconds, so milliseconds are .000.
func isoZeit(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05.000Z07:00") }

// Erstellungszeitpunkt returns Z_ERSTELLUNG: the tagesabschluss-erstellt time of a
// closed session, else fallback (the export time of an open session).
func Erstellungszeitpunkt(events []event.Event, fallback time.Time) time.Time {
	for _, ev := range events {
		if kasse.EventType(ev.Type) == kasse.EventTypeTagesabschlussErstelltV1 {
			return ev.Time
		}
	}
	return fallback
}

// abrechnungskreis names the subject's Tisch-Session (F-06); subjects without a table
// get none.
func abrechnungskreis(subject string, tischnamen map[int]string) string {
	tischID, err := kasse.ParseTischIDFromSubject(subject)
	if err != nil {
		return ""
	}
	return abrechnungskreisFuerTisch(tischID, tischnamen)
}

// abrechnungskreisFuerTisch names a table from the master data, truncated to the field
// length; table ID 0 (Direktverkauf) gets none. The fallback "Tisch N" is only right
// while table ID and name happen to coincide.
func abrechnungskreisFuerTisch(tischID int, tischnamen map[int]string) string {
	if tischID == 0 {
		return ""
	}
	if name, ok := tischnamen[tischID]; ok {
		return truncateRunes(name, maxLengthAbrechnungskreis)
	}
	return fmt.Sprintf("Tisch %d", tischID)
}

// --- Stammdatenmodul ---

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
		{"10,70", "Durchschnittsatz (§ 24 Abs. 1 Nr. 3 UStG)"},
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

func buildTSE(s Snapshot, erstellung string) Table {
	// TSE_ZEITFORMAT declares the TSE's own log time format (fiskaly: unixTime) from the
	// master data stored at setup; TSE_TA_START/ENDE are ISO 8601 regardless.
	zeitformat := s.TSEStammdaten.LogTimeFormat
	if zeitformat == "" {
		zeitformat = defaultTSEZeitformat
	}

	record := []string{
		s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
		tseReferenzID, s.TSEStammdaten.Seriennummer, s.TSEStammdaten.SignaturAlgorithmus,
		zeitformat, tsePDEncoding, s.TSEStammdaten.PublicKey,
	}
	for i := range zertifikatSpalten {
		record = append(record, certChunk(s.TSEStammdaten.Zertifikat, i))
	}

	return Table{
		File:        "tse.csv",
		LogicalName: "Stamm_TSE",
		Description: "Stammdaten der technischen Sicherheitseinrichtung",
		Columns:     tseColumns,
		Records:     [][]string{record},
	}
}

// --- Einzelaufzeichnungsmodul ---

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
			// An unsigned Vorgang that requires a signature (open, failed, tse_nicht_konfiguriert)
			// gets a TSE_TA_FEHLER row instead of none (docs/compliance.md §3.8). Vorgänge
			// without signature duty get no row.
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, tseReferenzID, "",
				"", "", "",
				"", "", tseFehlerAusfall,
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

// --- Kassenabschlussmodul ---

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

// --- Helpers ---

// truncateRunes cuts wert to maxLength runes, as the official field lengths count
// characters; only stored values that skipped write validation (TEXT columns) exceed
// them. Steuernummer and USt-IdNr. are never cut: a truncated number is a wrong one.
func truncateRunes(wert string, maxLength int) string {
	runen := []rune(wert)
	if len(runen) <= maxLength {
		return wert
	}

	return string(runen[:maxLength])
}

// barbestand sums the Bar Belege with sign: sales and Anfangsbestand add, withdrawals,
// Warenrücknahmen and Pfand returns subtract, AVBestellungen carry no Bar. It feeds
// Z_SE_(BAR)ZAHLUNGEN and cash_per_currency.csv.
func barbestand(belege []beleg) int {
	bar := 0
	for bi := range belege {
		b := &belege[bi]
		if b.zahlart == zahlartBar {
			bar += b.sign() * b.bruttoCents
		}
	}
	return bar
}

// ordnung returns a key's sort position; unknown keys sort after the known ones,
// alphabetically among themselves.
func ordnung(reihenfolge map[string]int, key string) int {
	if v, ok := reihenfolge[key]; ok {
		return v
	}
	return len(reihenfolge) + 1
}

// ZertifikatZuLang reports whether the TSE certificate exceeds the two official
// TSE_ZERTIFIKAT fields and is exported empty. The caller logs a warning; the archive
// stays valid, as TSE master data and the vendor export hold the full certificate.
func ZertifikatZuLang(cert string) bool {
	return len(cert) > zertifikatSpalten*zertifikatChunk
}

// certChunk returns the index-th 1000-character block of the base64 certificate (ASCII,
// so byte slicing is safe). A certificate longer than both fields (e.g. a whole chain)
// leaves both empty, since a truncated one is worthless.
func certChunk(cert string, index int) string {
	if ZertifikatZuLang(cert) {
		return ""
	}
	start := index * zertifikatChunk
	if start >= len(cert) {
		return ""
	}
	end := min(start+zertifikatChunk, len(cert))
	return cert[start:end]
}

// bonName returns BON_NAME: the spec requires it for the AVSonstige Bon, which jotti
// fills with "Tagesabschluss"; all other Bons leave it empty.
func bonName(b *beleg) string {
	if b.bonTyp == bonTypSonstige {
		return "Tagesabschluss"
	}
	return ""
}

func positionText(p kasse.PositionEventData) string {
	return kasse.PositionFromEventData(p).Bezeichnung()
}
