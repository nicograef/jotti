package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/nicograef/jotti/backend/api/druck/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/internal/logging"
)

// vorgangSuche finds the event of one Vorgang type by the ID in its data.
type vorgangSuche[D any] struct {
	istTyp   func(eventType string) bool
	id       func(D) string
	notFound error
}

func typ(t kasse.EventType) func(string) bool {
	return func(eventType string) bool { return eventType == string(t) }
}

var (
	zahlungSuche = vorgangSuche[kasse.ZahlungKassiertV1Data]{
		istTyp: kasse.IstZahlungKassiert, id: func(d kasse.ZahlungKassiertV1Data) string { return d.ZahlungID }, notFound: ErrZahlungNichtGefunden,
	}
	stornierungSuche = vorgangSuche[kasse.StornierungErteiltV1Data]{
		istTyp: typ(kasse.EventTypeStornierungErteiltV1), id: func(d kasse.StornierungErteiltV1Data) string { return d.StornierungID }, notFound: ErrStornierungNichtGefunden,
	}
	direktverkaufSuche = vorgangSuche[kasse.DirektverkaufGetaetigtV1Data]{
		istTyp: kasse.IstDirektverkaufGetaetigt, id: func(d kasse.DirektverkaufGetaetigtV1Data) string { return d.VerkaufID }, notFound: ErrVerkaufNichtGefunden,
	}
	direktverkaufStornoSuche = vorgangSuche[kasse.DirektverkaufStorniertV1Data]{
		istTyp: typ(kasse.EventTypeDirektverkaufStorniertV1), id: func(d kasse.DirektverkaufStorniertV1Data) string { return d.StornierungID }, notFound: ErrStornierungNichtGefunden,
	}
	pfandrueckgabeSuche = vorgangSuche[kasse.PfandZurueckgegebenV1Data]{
		istTyp: typ(kasse.EventTypePfandZurueckgegebenV1), id: func(d kasse.PfandZurueckgegebenV1Data) string { return d.RueckgabeID }, notFound: ErrRueckgabeNichtGefunden,
	}
	pfandStornoSuche = vorgangSuche[kasse.PfandrueckgabeStorniertV1Data]{
		istTyp: typ(kasse.EventTypePfandrueckgabeStorniertV1), id: func(d kasse.PfandrueckgabeStorniertV1Data) string { return d.StornierungID }, notFound: ErrStornierungNichtGefunden,
	}
)

// finde returns notFound when no event of the type carries want; a decode failure is returned as is.
func (s vorgangSuche[D]) finde(events []event.Event, want string) (event.Event, D, error) {
	var leer D
	for _, evt := range events {
		if !s.istTyp(evt.Type) {
			continue
		}
		var data D
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return event.Event{}, leer, err
		}
		if s.id(data) == want {
			return evt, data, nil
		}
	}
	return event.Event{}, leer, s.notFound
}

// fehler logs a finde error: a missing event warns and keeps its sentinel, a decode failure is ErrDatabase.
func (s vorgangSuche[D]) fehler(log *slog.Logger, err error, notFoundMsg, decodeMsg string, logArgs ...any) error {
	if errors.Is(err, s.notFound) {
		log.Warn(notFoundMsg, logArgs...)
		return s.notFound
	}
	log.Error(decodeMsg, append([]any{"error", err}, logArgs...)...)
	return ErrDatabase
}

func TSEAbschnittAusSignatur(signatur *tse.Signatur) *escpos.TSEAbschnitt {
	return &escpos.TSEAbschnitt{
		TransaktionNr:   signatur.TransaktionNummer,
		Signaturzaehler: signatur.SignaturZaehler,
		TSESeriennummer: signatur.TSESeriennummer,
		ZeitpunktBeginn: signatur.LogTimeStart,
		ZeitpunktEnde:   signatur.LogTimeEnd,
		Signatur:        signatur.Signatur,
		QRCodeData:      signatur.QRCodeData,
	}
}

// BelegStatus ist die Beleg-Abruf-Antwort: eingereiht (Druckauftrag angelegt)
// oder ausstehend (TSE-Signatur liegt noch nicht vor, kein Druckauftrag — die
// UI ruft denselben Endpunkt erneut auf).
type BelegStatus string

const (
	BelegStatusEingereiht BelegStatus = "eingereiht"
	BelegStatusAusstehend BelegStatus = "ausstehend"
)

// tseAbschnittFuerBeleg löst den TSE-Abschnitt über tse.DetermineSignaturstatus
// auf — die einzige Implementierung des Ausfallbegriffs. Kein Signaturauftrag
// heisst: nicht signaturpflichtig, Beleg ohne TSE-Abschnitt und ohne Vermerk.
func (c Command) tseAbschnittFuerBeleg(ctx context.Context, eventID int) (abschnitt *escpos.TSEAbschnitt, vermerk escpos.TSEBelegvermerk, ausstehend bool, err error) {
	stand, err := c.TSERepo.GetSignaturauftragZuEvent(ctx, eventID)
	if errors.Is(err, db.ErrNotFound) {
		return nil, escpos.KeinTSEVermerk, false, nil
	}
	if err != nil {
		return nil, escpos.KeinTSEVermerk, false, err
	}

	aktiveStoerung, err := c.TSERepo.GetAktiveTSEStoerung(ctx)
	if err != nil {
		return nil, escpos.KeinTSEVermerk, false, err
	}

	ergebnis := tse.DetermineSignaturstatus(stand, aktiveStoerung)
	switch ergebnis.Status {
	case tse.SignaturstatusVorhanden:
		return TSEAbschnittAusSignatur(ergebnis.Signatur), escpos.KeinTSEVermerk, false, nil
	case tse.SignaturstatusNachsigniert:
		abschnitt := TSEAbschnittAusSignatur(ergebnis.Signatur)
		abschnitt.Nachsigniert = true
		return abschnitt, escpos.KeinTSEVermerk, false, nil
	case tse.SignaturstatusAusfall:
		logging.FromContext(ctx).Warn("Kassenbeleg ohne TSE-Signatur: dokumentierter Ausfall", "event_id", eventID, "ausfall_grund", ergebnis.AusfallGrund)
		return nil, vermerkFuerAusfall(ergebnis), false, nil
	case tse.SignaturstatusAusstehend:
		fallthrough
	default:
		return nil, escpos.KeinTSEVermerk, true, nil
	}
}

// vermerkFuerAusfall: only an Ausfall during a Störungszeitraum is re-signed once the TSE is back.
// Without a configured TSE, and after a final fehlgeschlagen, the Vorgang stays unsigned.
func vermerkFuerAusfall(ergebnis tse.SignaturstatusErgebnis) escpos.TSEBelegvermerk {
	switch {
	case ergebnis.OhneKonfiguration():
		return escpos.TSEVermerkKeineKonfiguration
	case ergebnis.AusfallGrund == tse.StatusFehlgeschlagen:
		return escpos.TSEVermerkSignaturFehlgeschlagen
	default:
		return escpos.TSEVermerkVoruebergehend
	}
}

func negierePositionen(positionen []kasse.Position) []kasse.Position {
	out := make([]kasse.Position, 0, len(positionen))
	for _, pos := range positionen {
		pos.EinzelpreisCents = -pos.EinzelpreisCents
		out = append(out, pos)
	}
	return out
}

// steuer.Aufteilen ignores negative Brutto: the matrix is computed from the
// positive amounts and negated afterwards (like the faktor in the TSE-processData).
func negiereAufteilungen(aufteilungen []steuer.Aufteilung) []steuer.Aufteilung {
	out := make([]steuer.Aufteilung, 0, len(aufteilungen))
	for _, aufteilung := range aufteilungen {
		aufteilung.Brutto = -aufteilung.Brutto
		aufteilung.Netto = -aufteilung.Netto
		aufteilung.Steuer = -aufteilung.Steuer
		out = append(out, aufteilung)
	}
	return out
}

type KassenbelegDruckenCommand struct {
	TischID       int
	ZahlungID     string
	VerkaufID     string
	RueckgabeID   string
	StornierungID string
}

// BelegQuelle is what a Kassenbeleg prints from one event. Positionen and GesamtbetragCents are positive; Negativ negates them with the
// Steuermatrix on the Beleg. Stornobeleg only sets the title and the "Storno zu Bon-Nr" line.
type BelegQuelle struct {
	Event                    event.Event
	Positionen               []kasse.Position
	GesamtbetragCents        int
	Referenz                 string
	ErsteBestellungZeitpunkt *time.Time
	Negativ                  bool
	Stornobeleg              bool
	StornoZuBelegnummer      string
}

func (c Command) KassenbelegDrucken(ctx context.Context, cmd KassenbelegDruckenCommand) (BelegStatus, error) {
	log := logging.FromContext(ctx)

	ks, err := sentinels.OffeneKassensitzung(ctx, c.KassensitzungenRepo)
	if err != nil {
		return "", err
	}

	quelle, err := c.resolveBelegQuelle(ctx, ks, cmd)
	if err != nil {
		return "", err
	}

	tseAbschnitt, tseVermerk, ausstehend, err := c.tseAbschnittFuerBeleg(ctx, quelle.Event.ID)
	if err != nil {
		log.Error("Failed to resolve TSE section for kassenbeleg", "error", err, "event_id", quelle.Event.ID)
		return "", ErrDatabase
	}
	if ausstehend {
		log.Info("Kassenbeleg zurückgestellt: TSE-Signatur ausstehend", "event_id", quelle.Event.ID)
		return BelegStatusAusstehend, nil
	}

	stationen, err := c.DruckstationRepo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		log.Error("Failed to load druckstationen for kassenbeleg", "error", err)
		return "", ErrDatabase
	}
	kassenbelegStation, ok := stationen[string(druckstation.KategorieKassenbeleg)]
	if !ok || kassenbelegStation.DruckerIP == "" {
		return "", ErrKassenbelegDruckerNichtKonfiguriert
	}

	betreiber, err := c.BetreiberRepo.GetBetreiber(ctx)
	if err != nil {
		log.Error("Failed to load betreiber for kassenbeleg", "error", err)
		return "", ErrDatabase
	}

	kassenidentitaet, err := c.TSERepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error("Failed to load kassenidentitaet for kassenbeleg", "error", err)
		return "", ErrDatabase
	}

	payload := escpos.FormatKassenbeleg(KassenbelegDaten(betreiber, kassenidentitaet.Seriennummer.String(), quelle, tseAbschnitt, tseVermerk))

	auftrag := druckstation.NeuerDruckauftrag{
		ZielIP:   kassenbelegStation.DruckerIP,
		Payload:  base64.StdEncoding.EncodeToString(payload),
		BonArt:   druckstation.BonArtKassenbeleg,
		Referenz: quelle.Referenz,
	}

	if err := c.DruckauftragRepo.EnqueueDruckauftraege(ctx, []druckstation.NeuerDruckauftrag{auftrag}); err != nil {
		log.Error("Failed to enqueue kassenbeleg", "error", err, "tisch_id", cmd.TischID, "zahlung_id", cmd.ZahlungID, "verkauf_id", cmd.VerkaufID, "rueckgabe_id", cmd.RueckgabeID)
		return "", ErrDatabase
	}

	log.Info("Kassenbeleg queued", "tisch_id", cmd.TischID, "verkauf_id", cmd.VerkaufID, "rueckgabe_id", cmd.RueckgabeID, "event_id", quelle.Event.ID)
	return BelegStatusEingereiht, nil
}

// KassenbelegDaten assembles the printed Beleg; the demo seed builds its receipts through it too.
func KassenbelegDaten(b betreiber.Betreiber, kassenSeriennummer string, quelle BelegQuelle, tseAbschnitt *escpos.TSEAbschnitt, tseVermerk escpos.TSEBelegvermerk) escpos.KassenbelegData {
	positionen := quelle.Positionen
	gesamtbetragCents := quelle.GesamtbetragCents
	steuermatrix := steuer.Steuermatrix(kasse.SteuermatrixPositionen(positionen))
	if quelle.Negativ {
		steuermatrix = negiereAufteilungen(steuermatrix)
		positionen = negierePositionen(positionen)
		gesamtbetragCents = -gesamtbetragCents
	}

	return escpos.KassenbelegData{
		Vereinsname:              b.Vereinsname,
		Strasse:                  b.Strasse,
		Plz:                      b.Plz,
		Ort:                      b.Ort,
		KassenSeriennummer:       kassenSeriennummer,
		Belegnummer:              fmt.Sprintf("%d", quelle.Event.ID),
		Zeitpunkt:                quelle.Event.Time,
		ErsteBestellungZeitpunkt: quelle.ErsteBestellungZeitpunkt,
		Positionen:               positionen,
		Steuermatrix:             steuermatrix,
		TSE:                      tseAbschnitt,
		TSEVermerk:               tseVermerk,
		GesamtbetragCents:        gesamtbetragCents,
		Zahlungsart:              "bar",
		Stornobeleg:              quelle.Stornobeleg,
		StornoZuBelegnummer:      quelle.StornoZuBelegnummer,
	}
}

func (c Command) resolveBelegQuelle(ctx context.Context, ks *kasse.Kassensitzung, cmd KassenbelegDruckenCommand) (BelegQuelle, error) {
	switch {
	case cmd.RueckgabeID != "" && cmd.StornierungID != "":
		return c.resolvePfandStornobeleg(ctx, ks, cmd.RueckgabeID, cmd.StornierungID)
	case cmd.RueckgabeID != "":
		return c.resolvePfandbeleg(ctx, ks, cmd.RueckgabeID)
	case cmd.VerkaufID != "" && cmd.StornierungID != "":
		return c.resolveDirektverkaufStornobeleg(ctx, ks, cmd.VerkaufID, cmd.StornierungID)
	case cmd.StornierungID != "":
		return c.resolveTischStornobeleg(ctx, ks, cmd.TischID, cmd.StornierungID)
	case cmd.VerkaufID != "":
		return c.resolveDirektverkaufBeleg(ctx, ks, cmd.VerkaufID)
	default:
		return c.resolveTischZahlungsbeleg(ctx, ks, cmd.TischID, cmd.ZahlungID)
	}
}

func (c Command) resolveDirektverkaufStornobeleg(ctx context.Context, ks *kasse.Kassensitzung, verkaufID, stornierungID string) (BelegQuelle, error) {
	log := logging.FromContext(ctx)

	subject := kasse.DirektverkaufSubject(ks.ZNr, verkaufID)
	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error("Failed to read direktverkauf events for stornobeleg", "error", err, "verkauf_id", verkaufID)
		return BelegQuelle{}, ErrDatabase
	}

	verkaufEvent, _, err := direktverkaufSuche.finde(events, verkaufID)
	if err != nil {
		return BelegQuelle{}, direktverkaufSuche.fehler(log, err, "Direktverkauf not found for stornobeleg", "Failed to decode direktverkauf event data", "verkauf_id", verkaufID)
	}

	stornoEvent, stornoData, err := direktverkaufStornoSuche.finde(events, stornierungID)
	if err != nil {
		return BelegQuelle{}, direktverkaufStornoSuche.fehler(log, err, "Stornierung not found for stornobeleg", "Failed to decode direktverkauf storno event data", "verkauf_id", verkaufID, "stornierung_id", stornierungID)
	}

	return BelegQuelle{
		Event:               stornoEvent,
		Positionen:          kasse.PositionenFromEventData(stornoData.Positionen),
		GesamtbetragCents:   stornoData.GesamtStornierungCents,
		Referenz:            druckstation.EventReferenz(druckstation.ReferenzDirektverkaufStorno, stornoEvent.ID),
		Negativ:             true,
		Stornobeleg:         true,
		StornoZuBelegnummer: fmt.Sprintf("%d", verkaufEvent.ID),
	}, nil
}

// resolvePfandbeleg: the Pfandbeleg is a Kassenbeleg with negative amounts, without Storno title.
func (c Command) resolvePfandbeleg(ctx context.Context, ks *kasse.Kassensitzung, rueckgabeID string) (BelegQuelle, error) {
	log := logging.FromContext(ctx)

	events, err := c.EventRepo.ReadEventsBySubject(ctx, kasse.PfandrueckgabeSubject(ks.ZNr, rueckgabeID))
	if err != nil {
		log.Error("Failed to read pfandrueckgabe events for kassenbeleg", "error", err, "rueckgabe_id", rueckgabeID)
		return BelegQuelle{}, ErrDatabase
	}

	rueckgabeEvent, rueckgabeData, err := pfandrueckgabeSuche.finde(events, rueckgabeID)
	if err != nil {
		return BelegQuelle{}, pfandrueckgabeSuche.fehler(log, err, "Pfandrückgabe not found for kassenbeleg", "Failed to decode pfandrueckgabe event data", "rueckgabe_id", rueckgabeID)
	}

	return BelegQuelle{
		Event:             rueckgabeEvent,
		Positionen:        kasse.PositionenFromEventData(rueckgabeData.Positionen),
		GesamtbetragCents: rueckgabeData.GesamtbetragCents,
		Referenz:          druckstation.EventReferenz(druckstation.ReferenzPfandrueckgabe, rueckgabeEvent.ID),
		Negativ:           true,
	}, nil
}

// resolvePfandStornobeleg: cancelling a Rückgabe books money back in, so its Stornobeleg carries
// positive amounts and points to the Rückgabe's Bon-Nr.
func (c Command) resolvePfandStornobeleg(ctx context.Context, ks *kasse.Kassensitzung, rueckgabeID, stornierungID string) (BelegQuelle, error) {
	log := logging.FromContext(ctx)

	events, err := c.EventRepo.ReadEventsBySubject(ctx, kasse.PfandrueckgabeSubject(ks.ZNr, rueckgabeID))
	if err != nil {
		log.Error("Failed to read pfandrueckgabe events for stornobeleg", "error", err, "rueckgabe_id", rueckgabeID)
		return BelegQuelle{}, ErrDatabase
	}

	rueckgabeEvent, _, err := pfandrueckgabeSuche.finde(events, rueckgabeID)
	if err != nil {
		return BelegQuelle{}, pfandrueckgabeSuche.fehler(log, err, "Pfandrückgabe not found for stornobeleg", "Failed to decode pfandrueckgabe event data", "rueckgabe_id", rueckgabeID)
	}

	stornoEvent, stornoData, err := pfandStornoSuche.finde(events, stornierungID)
	if err != nil {
		return BelegQuelle{}, pfandStornoSuche.fehler(log, err, "Stornierung not found for pfand stornobeleg", "Failed to decode pfandrueckgabe storno event data", "rueckgabe_id", rueckgabeID, "stornierung_id", stornierungID)
	}

	return BelegQuelle{
		Event:               stornoEvent,
		Positionen:          kasse.PositionenFromEventData(stornoData.Positionen),
		GesamtbetragCents:   stornoData.GesamtStornierungCents,
		Referenz:            druckstation.EventReferenz(druckstation.ReferenzPfandrueckgabeStorno, stornoEvent.ID),
		Stornobeleg:         true,
		StornoZuBelegnummer: fmt.Sprintf("%d", rueckgabeEvent.ID),
	}, nil
}

func (c Command) resolveTischStornobeleg(ctx context.Context, ks *kasse.Kassensitzung, tischID int, stornierungID string) (BelegQuelle, error) {
	log := logging.FromContext(ctx)

	subject := kasse.TischSessionSubject(ks.ZNr, tischID)
	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error("Failed to read events for stornobeleg", "error", err, "tisch_id", tischID)
		return BelegQuelle{}, ErrDatabase
	}

	stornoEvent, stornoData, err := stornierungSuche.finde(events, stornierungID)
	if err != nil {
		return BelegQuelle{}, stornierungSuche.fehler(log, err, "Stornierung not found for stornobeleg", "Failed to decode stornierung event data", "tisch_id", tischID, "stornierung_id", stornierungID)
	}

	zahlungEvent, _, err := zahlungSuche.finde(events, stornoData.ZahlungID)
	if err != nil {
		return BelegQuelle{}, zahlungSuche.fehler(log, err, "Referenzierte Zahlung not found for stornobeleg", "Failed to decode referenzierte zahlung event data", "tisch_id", tischID, "zahlung_id", stornoData.ZahlungID)
	}

	return BelegQuelle{
		Event:               stornoEvent,
		Positionen:          kasse.PositionenFromEventData(stornoData.Positionen),
		GesamtbetragCents:   stornoData.GesamtStornierungCents,
		Referenz:            druckstation.EventReferenz(druckstation.ReferenzStornierung, stornoEvent.ID),
		Negativ:             true,
		Stornobeleg:         true,
		StornoZuBelegnummer: fmt.Sprintf("%d", zahlungEvent.ID),
	}, nil
}

func (c Command) resolveDirektverkaufBeleg(ctx context.Context, ks *kasse.Kassensitzung, verkaufID string) (BelegQuelle, error) {
	log := logging.FromContext(ctx)

	subject := kasse.DirektverkaufSubject(ks.ZNr, verkaufID)
	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error("Failed to read direktverkauf events for kassenbeleg", "error", err, "verkauf_id", verkaufID)
		return BelegQuelle{}, ErrDatabase
	}

	verkaufEvent, verkaufData, err := direktverkaufSuche.finde(events, verkaufID)
	if err != nil {
		return BelegQuelle{}, direktverkaufSuche.fehler(log, err, "Direktverkauf not found for kassenbeleg", "Failed to decode direktverkauf event data", "verkauf_id", verkaufID)
	}

	return BelegQuelle{
		Event:             verkaufEvent,
		Positionen:        kasse.PositionenFromEventData(verkaufData.Positionen),
		GesamtbetragCents: verkaufData.GesamtbetragCents,
		Referenz:          druckstation.EventReferenz(druckstation.ReferenzDirektverkauf, verkaufEvent.ID),
	}, nil
}

func (c Command) resolveTischZahlungsbeleg(ctx context.Context, ks *kasse.Kassensitzung, tischID int, zahlungID string) (BelegQuelle, error) {
	log := logging.FromContext(ctx)

	subject := kasse.TischSessionSubject(ks.ZNr, tischID)
	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error("Failed to read events for kassenbeleg", "error", err, "tisch_id", tischID)
		return BelegQuelle{}, ErrDatabase
	}

	zahlungEvent, zahlungData, err := zahlungSuche.finde(events, zahlungID)
	if err != nil {
		return BelegQuelle{}, zahlungSuche.fehler(log, err, "Zahlung not found for kassenbeleg", "Failed to decode zahlung event data", "tisch_id", tischID, "zahlung_id", zahlungID)
	}

	ersteBestellung, err := c.ersteBestellungZeitpunkt(ctx, events)
	if err != nil {
		log.Error("Failed to resolve first bestellung for kassenbeleg", "error", err, "tisch_id", tischID)
		return BelegQuelle{}, ErrDatabase
	}

	return BelegQuelle{
		Event:                    zahlungEvent,
		Positionen:               kasse.PositionenFromEventData(zahlungData.Positionen),
		GesamtbetragCents:        zahlungData.GesamtZahlungCents,
		Referenz:                 druckstation.EventReferenz(druckstation.ReferenzZahlung, zahlungEvent.ID),
		ErsteBestellungZeitpunkt: ersteBestellung,
	}, nil
}

// ersteBestellungZeitpunkt is the TSE start of the session's first Bestellung (DSFinV-K Tz. 2.7.2). An
// unsigned first Bestellung (TSE-Ausfall, no TSE, signature pending) keeps its event time.
func (c Command) ersteBestellungZeitpunkt(ctx context.Context, events []event.Event) (*time.Time, error) {
	erste, found, err := kasse.ErsteBestellung(events)
	if err != nil || !found {
		return nil, err
	}
	zeitpunkt := erste.Time.UTC()
	stand, err := c.TSERepo.GetSignaturauftragZuEvent(ctx, erste.ID)
	if errors.Is(err, db.ErrNotFound) {
		return &zeitpunkt, nil
	}
	if err != nil {
		return nil, err
	}
	if stand.Signatur != nil {
		zeitpunkt = stand.Signatur.LogTimeStart.UTC()
	}
	return &zeitpunkt, nil
}
