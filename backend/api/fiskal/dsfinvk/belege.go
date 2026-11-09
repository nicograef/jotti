package dsfinvk

import (
	"fmt"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// beleg is the per-Bon view that several tables derive from (Bonkopf, its VAT,
// payment and line details, and the TSE row).
type beleg struct {
	bonID            string
	bonNr            int
	bonTyp           string   // BON_TYP: "Beleg", "AVBestellung" or "AVSonstige"
	gvTyp            string   // GV_TYP of the lines: "Umsatz", "PfandRueckzahlung" or a cash GV type; pfand overrides it per line
	zahlart          string   // ZAHLART_TYP: "Bar"; empty for the cash-neutral AVBestellung
	abrechnungskreis string   // ABRECHNUNGSKREIS (table name); empty without a table (e.g. Direktverkauf)
	storno           bool     // negative Beleg (Warenrücknahme, Korrektur, Umbuchung Abgang, Pfand return): flips the sign; BON_STORNO stays 0
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
	tseStatus        string        // the Signaturauftrag status
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

// belegeFromEvents derives the Belege from events ordered by id, so an origin Bon
// always precedes its Storno. BON_NR is the kassenjournal event ID, the Bon-Nr the printed
// Kassenbeleg shows; BON_ID is the Vorgang ID.
func belegeFromEvents(events []event.Event, tischnamen map[int]string, signaturen map[int]tse.EventSignatur) ([]beleg, error) {
	a := belegAbleitung{
		tischnamen:    tischnamen,
		herkunft:      map[string]string{},
		verkaufsPfand: map[string]map[string]bool{},
	}
	var belege []beleg
	for _, ev := range events {
		b, ok, err := a.beleg(ev)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		// A Beleg takes its event's ID as BON_NR, its time and operator, and its TSE state from the
		// Signaturauftrag table, the single source for signature duty, signature and TSE_TA_VORGANGSART.
		sig, pflichtig := signaturen[ev.ID]
		b.bonNr = ev.ID
		b.zeitpunkt = zeit(ev)
		b.bedienerID = ev.UserID
		b.bedienerName = ev.UserName
		b.tsePflichtig = pflichtig
		b.tseStatus = sig.Status
		b.tse = sig.Signatur
		b.processType = sig.ProcessType
		belege = append(belege, b)
	}
	return belege, nil
}

// belegAbleitung holds what one event's Beleg needs from earlier events.
type belegAbleitung struct {
	tischnamen map[int]string
	// herkunft maps each PositionID to its order's BON_ID. PositionIDs are unique per
	// order, so a Korrektur resolves its origin orders via its positions.
	herkunft map[string]string
	// verkaufsPfand maps a sale's BON_ID to its Pfand positions; its Warenrücknahme takes the
	// flags from there, since the frozen Storno events carry none.
	verkaufsPfand map[string]map[string]bool
}

// beleg derives the event's Beleg; false means the event yields none.
func (a *belegAbleitung) beleg(ev event.Event) (beleg, bool, error) {
	switch kasse.EventType(ev.Type) {
	case kasse.EventTypeBestellungAufgenommenV1:
		return mitBeleg(a.bestellung(ev))
	case kasse.EventTypeZahlungKassiertV1, kasse.EventTypeZahlungKassiertV2:
		return mitBeleg(a.zahlung(ev))
	case kasse.EventTypeStornierungErteiltV1:
		return mitBeleg(a.warenruecknahme(ev))
	case kasse.EventTypeBestellungKorrigiertV1:
		return mitBeleg(a.korrektur(ev))
	case kasse.EventTypeBestellungUmgebuchtV1:
		return mitBeleg(a.umbuchung(ev))
	case kasse.EventTypeDirektverkaufGetaetigtV1, kasse.EventTypeDirektverkaufGetaetigtV2:
		return mitBeleg(a.direktverkauf(ev))
	case kasse.EventTypeDirektverkaufStorniertV1:
		return mitBeleg(a.direktverkaufStorno(ev))
	case kasse.EventTypePfandZurueckgegebenV1:
		return mitBeleg(a.pfandrueckgabe(ev))
	case kasse.EventTypePfandrueckgabeStorniertV1:
		return mitBeleg(a.pfandrueckgabeStorno(ev))
	case kasse.EventTypeKassensitzungEroeffnetV1:
		return anfangsbestand(ev)
	case kasse.EventTypeGeldtransitGebuchtV1:
		return mitBeleg(geldtransit(ev))
	case kasse.EventTypeDifferenzSollIstGebuchtV1:
		return mitBeleg(differenzSollIst(ev))
	case kasse.EventTypeKassensturzDurchgefuehrtV1:
		// The count itself moves no cash; its Differenz is booked as its own event.
		return beleg{}, false, nil
	case kasse.EventTypeTagesabschlussErstelltV1:
		return mitBeleg(tagesabschluss(ev))
	}
	return beleg{}, false, nil
}

func mitBeleg(b beleg, err error) (beleg, bool, error) {
	return b, err == nil, err
}

func (a *belegAbleitung) bestellung(ev event.Event) (beleg, error) {
	data, err := event.Decode[kasse.BestellungAufgenommenV1Data](ev)
	if err != nil {
		return beleg{}, err
	}
	for _, p := range data.Positionen {
		a.herkunft[p.PositionID] = data.BestellungID
	}
	return beleg{
		bonID:            data.BestellungID,
		bonTyp:           bonTypBestellung,
		gvTyp:            gvTypUmsatz,
		geldneutral:      true,
		abrechnungskreis: abrechnungskreis(ev.Subject, a.tischnamen),
		positionen:       kasse.PositionenOhneKommentar(data.Positionen),
		bruttoCents:      data.GesamtPreisCents,
		notiz:            data.Kommentar,
	}, nil
}

func (a *belegAbleitung) zahlung(ev event.Event) (beleg, error) {
	// v1 has no pfand key, so its positions read as no Pfand.
	data, err := event.Decode[kasse.ZahlungKassiertV2Data](ev)
	if err != nil {
		return beleg{}, err
	}
	positionen, pfand := verkaufsPositionen(data.Positionen)
	a.verkaufsPfand[data.ZahlungID] = pfandJePosition(data.Positionen)
	return beleg{
		bonID:            data.ZahlungID,
		bonTyp:           bonTypBeleg,
		gvTyp:            gvTypUmsatz,
		zahlart:          zahlartBar,
		abrechnungskreis: abrechnungskreis(ev.Subject, a.tischnamen),
		positionen:       positionen,
		pfand:            pfand,
		bruttoCents:      data.GesamtZahlungCents,
		notiz:            data.Kommentar,
	}, nil
}

// warenruecknahme is the Storno of paid lines: negative revenue at the original VAT rate with cash
// refund (DSFinV-K Tz. 4.2.5), referencing the Zahlung (Tz. 4.2.2, docs/compliance.md §6.6).
func (a *belegAbleitung) warenruecknahme(ev event.Event) (beleg, error) {
	data, err := event.Decode[kasse.StornierungErteiltV1Data](ev)
	if err != nil {
		return beleg{}, err
	}
	return beleg{
		bonID:            data.StornierungID,
		bonTyp:           bonTypBeleg,
		gvTyp:            gvTypUmsatz,
		zahlart:          zahlartBar,
		storno:           true,
		abrechnungskreis: abrechnungskreis(ev.Subject, a.tischnamen),
		refBonIDs:        []string{data.ZahlungID},
		positionen:       data.Positionen,
		pfand:            ruecknahmePfand(data.Positionen, a.verkaufsPfand[data.ZahlungID]),
		bruttoCents:      data.GesamtStornierungCents,
		notiz:            data.Kommentar,
	}, nil
}

// korrektur is the cash-neutral cancellation of unpaid lines: an AVBestellung without revenue,
// Zahlart or cash effect, referencing the origin order.
func (a *belegAbleitung) korrektur(ev event.Event) (beleg, error) {
	data, err := event.Decode[kasse.BestellungKorrigiertV1Data](ev)
	if err != nil {
		return beleg{}, err
	}
	return beleg{
		bonID:            data.KorrekturID,
		bonTyp:           bonTypBestellung,
		gvTyp:            gvTypUmsatz,
		geldneutral:      true,
		abrechnungskreis: abrechnungskreis(ev.Subject, a.tischnamen),
		storno:           true,
		refBonIDs:        ursprungsbons(data.Positionen, a.herkunft),
		positionen:       data.Positionen,
		bruttoCents:      data.GesamtCents,
		notiz:            data.Kommentar,
	}, nil
}

// umbuchung is a cash-neutral AVBestellung. Source and target share the UmbuchungID:
// the Abgang carries it as BON_ID, the Zugang references it. The Abgang negates MENGE
// (DSFinV-K 3.0 Tz. 4.2.8), matching the signed TSE processData of the fiscal projection.
func (a *belegAbleitung) umbuchung(ev event.Event) (beleg, error) {
	data, err := event.Decode[kasse.BestellungUmgebuchtV1Data](ev)
	if err != nil {
		return beleg{}, err
	}
	tischID, err := kasse.ParseTischIDFromSubject(ev.Subject)
	if err != nil {
		return beleg{}, fmt.Errorf("parse tisch from bestellung-umgebucht (event %d): %w", ev.ID, err)
	}
	bon := beleg{
		bonTyp:           bonTypBestellung,
		gvTyp:            gvTypUmsatz,
		geldneutral:      true,
		abrechnungskreis: abrechnungskreis(ev.Subject, a.tischnamen),
		positionen:       data.Positionen,
		bruttoCents:      data.GesamtCents,
		notiz:            umbuchungNotiz(data.Kommentar, data.BenutzerKommentar),
	}
	if tischID == data.QuellTischID {
		bon.bonID = data.UmbuchungID
		bon.storno = true
		return bon, nil
	}
	bon.bonID = fmt.Sprintf("umbuchung-%d", ev.ID)
	bon.refBonIDs = []string{data.UmbuchungID}
	for _, p := range data.Positionen {
		a.herkunft[p.PositionID] = bon.bonID
	}
	return bon, nil
}

func (a *belegAbleitung) direktverkauf(ev event.Event) (beleg, error) {
	// v1 has no pfand key, so its positions read as no Pfand.
	data, err := event.Decode[kasse.DirektverkaufGetaetigtV2Data](ev)
	if err != nil {
		return beleg{}, err
	}
	basis := kasse.PositionenV2OhneKommentar(data.Positionen)
	positionen, pfand := verkaufsPositionen(basis)
	a.verkaufsPfand[data.VerkaufID] = pfandJePosition(basis)
	return beleg{
		bonID:       data.VerkaufID,
		bonTyp:      bonTypBeleg,
		gvTyp:       gvTypUmsatz,
		zahlart:     zahlartBar,
		positionen:  positionen,
		pfand:       pfand,
		bruttoCents: data.GesamtbetragCents,
		notiz:       data.Kommentar,
	}, nil
}

func (a *belegAbleitung) direktverkaufStorno(ev event.Event) (beleg, error) {
	data, err := event.Decode[kasse.DirektverkaufStorniertV1Data](ev)
	if err != nil {
		return beleg{}, err
	}
	return beleg{
		bonID:       data.StornierungID,
		bonTyp:      bonTypBeleg,
		gvTyp:       gvTypUmsatz,
		zahlart:     zahlartBar,
		storno:      true,
		refBonIDs:   []string{data.VerkaufID},
		positionen:  data.Positionen,
		pfand:       ruecknahmePfand(data.Positionen, a.verkaufsPfand[data.VerkaufID]),
		bruttoCents: data.GesamtStornierungCents,
		notiz:       data.Kommentar,
	}, nil
}

// pfandrueckgabe is Pfand paid out in cash: a negative Beleg without reference, since a Pfand return
// reduces the consideration independently of the sale (Abschn. 10.1 Abs. 8 Satz 3 UStAE).
func (a *belegAbleitung) pfandrueckgabe(ev event.Event) (beleg, error) {
	data, err := event.Decode[kasse.PfandZurueckgegebenV1Data](ev)
	if err != nil {
		return beleg{}, err
	}
	return beleg{
		bonID:            data.RueckgabeID,
		bonTyp:           bonTypBeleg,
		gvTyp:            gvTypPfandRueckzahlung,
		zahlart:          zahlartBar,
		storno:           true,
		abrechnungskreis: abrechnungskreisFuerTisch(data.TischID, a.tischnamen),
		positionen:       data.Positionen,
		bruttoCents:      data.GesamtbetragCents,
		notiz:            data.Kommentar,
	}, nil
}

// pfandrueckgabeStorno is a positive Beleg referencing the Rückgabe.
func (a *belegAbleitung) pfandrueckgabeStorno(ev event.Event) (beleg, error) {
	data, err := event.Decode[kasse.PfandrueckgabeStorniertV1Data](ev)
	if err != nil {
		return beleg{}, err
	}
	return beleg{
		bonID:            data.StornierungID,
		bonTyp:           bonTypBeleg,
		gvTyp:            gvTypPfandRueckzahlung,
		zahlart:          zahlartBar,
		abrechnungskreis: abrechnungskreisFuerTisch(data.TischID, a.tischnamen),
		refBonIDs:        []string{data.RueckgabeID},
		positionen:       data.Positionen,
		bruttoCents:      data.GesamtStornierungCents,
		notiz:            data.Kommentar,
	}, nil
}

// anfangsbestand yields no Beleg without cash at session start (Anhang C: recording not mandatory).
func anfangsbestand(ev event.Event) (beleg, bool, error) {
	data, err := event.Decode[kasse.KassensitzungEroeffnetV1Data](ev)
	if err != nil {
		return beleg{}, false, err
	}
	if data.BetragCents == 0 {
		return beleg{}, false, nil
	}
	return geldbewegung(fmt.Sprintf("anfangsbestand-%d", ev.ID), gvTypAnfangsbestand, data.BetragCents, false, data.Bezeichnung), true, nil
}

func geldtransit(ev event.Event) (beleg, error) {
	data, err := event.Decode[kasse.GeldtransitGebuchtV1Data](ev)
	if err != nil {
		return beleg{}, err
	}
	return geldbewegung(data.GeldtransitID, gvTypGeldtransit, data.BetragCents, data.Richtung == kasse.GeldtransitRichtungEntnahme, data.Kommentar), nil
}

// differenzSollIst books BetragCents = Soll − Ist: positive is a shortfall (cash missing, reduce
// the balance), negative a surplus.
func differenzSollIst(ev event.Event) (beleg, error) {
	data, err := event.Decode[kasse.DifferenzSollIstGebuchtV1Data](ev)
	if err != nil {
		return beleg{}, err
	}
	return geldbewegung(fmt.Sprintf("differenz-soll-ist-%d", ev.ID), gvTypDifferenzSollIst, abs(data.BetragCents), data.BetragCents > 0, ""), nil
}

// tagesabschluss is a TSE-signed, cash-neutral AVSonstige without lines. It exists in the export
// so its signature gets a transactions_tse.csv row (docs/compliance.md §6.8).
func tagesabschluss(ev event.Event) (beleg, error) {
	if _, err := event.Decode[kasse.TagesabschlussErstelltV1Data](ev); err != nil {
		return beleg{}, err
	}
	return beleg{
		bonID:       fmt.Sprintf("tagesabschluss-%d", ev.ID),
		bonTyp:      bonTypSonstige,
		geldneutral: true,
	}, nil
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
