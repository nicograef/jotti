package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nicograef/jotti/backend/api/druck/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/internal/logging"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
)

func toSteuermatrixPositionen(positionen []kasse.Position) []steuer.SteuermatrixPosition {
	matrixPositionen := make([]steuer.SteuermatrixPosition, 0, len(positionen))
	for _, position := range positionen {
		matrixPositionen = append(matrixPositionen, steuer.SteuermatrixPosition{
			Brutto:     position.EinzelpreisCents * position.Menge,
			Steuersatz: steuer.Steuersatz(position.Steuersatz),
		})
	}

	return matrixPositionen
}

func findZahlungEvent(events []event.Event, zahlungID string) (event.Event, kasse.ZahlungKassiertV1Data, error) {
	for _, evt := range events {
		if !kasse.IstZahlungKassiert(evt.Type) {
			continue
		}

		var data kasse.ZahlungKassiertV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return event.Event{}, kasse.ZahlungKassiertV1Data{}, err
		}
		if data.ZahlungID == zahlungID {
			return evt, data, nil
		}
	}

	return event.Event{}, kasse.ZahlungKassiertV1Data{}, ErrZahlungNichtGefunden
}

func tseAbschnittAusSignatur(signatur *tse.Signatur) *escpos.TSEAbschnitt {
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

func findStornierungEvent(events []event.Event, stornierungID string) (event.Event, kasse.StornierungErteiltV1Data, error) {
	for _, evt := range events {
		if evt.Type != string(kasse.EventTypeStornierungErteiltV1) {
			continue
		}

		var data kasse.StornierungErteiltV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return event.Event{}, kasse.StornierungErteiltV1Data{}, err
		}
		if data.StornierungID == stornierungID {
			return evt, data, nil
		}
	}

	return event.Event{}, kasse.StornierungErteiltV1Data{}, ErrStornierungNichtGefunden
}

func findDirektverkaufGetaetigtEvent(events []event.Event, verkaufID string) (event.Event, kasse.DirektverkaufGetaetigtV1Data, error) {
	for _, evt := range events {
		if !kasse.IstDirektverkaufGetaetigt(evt.Type) {
			continue
		}

		var data kasse.DirektverkaufGetaetigtV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return event.Event{}, kasse.DirektverkaufGetaetigtV1Data{}, err
		}
		if data.VerkaufID == verkaufID {
			return evt, data, nil
		}
	}

	return event.Event{}, kasse.DirektverkaufGetaetigtV1Data{}, ErrVerkaufNichtGefunden
}

func findDirektverkaufStorniertEvent(events []event.Event, stornierungID string) (event.Event, kasse.DirektverkaufStorniertV1Data, error) {
	for _, evt := range events {
		if evt.Type != string(kasse.EventTypeDirektverkaufStorniertV1) {
			continue
		}

		var data kasse.DirektverkaufStorniertV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return event.Event{}, kasse.DirektverkaufStorniertV1Data{}, err
		}
		if data.StornierungID == stornierungID {
			return evt, data, nil
		}
	}

	return event.Event{}, kasse.DirektverkaufStorniertV1Data{}, ErrStornierungNichtGefunden
}

func findPfandZurueckgegebenEvent(events []event.Event, rueckgabeID string) (event.Event, kasse.PfandZurueckgegebenV1Data, error) {
	for _, evt := range events {
		if evt.Type != string(kasse.EventTypePfandZurueckgegebenV1) {
			continue
		}

		var data kasse.PfandZurueckgegebenV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return event.Event{}, kasse.PfandZurueckgegebenV1Data{}, err
		}
		if data.RueckgabeID == rueckgabeID {
			return evt, data, nil
		}
	}

	return event.Event{}, kasse.PfandZurueckgegebenV1Data{}, ErrRueckgabeNichtGefunden
}

func findPfandrueckgabeStorniertEvent(events []event.Event, stornierungID string) (event.Event, kasse.PfandrueckgabeStorniertV1Data, error) {
	for _, evt := range events {
		if evt.Type != string(kasse.EventTypePfandrueckgabeStorniertV1) {
			continue
		}

		var data kasse.PfandrueckgabeStorniertV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return event.Event{}, kasse.PfandrueckgabeStorniertV1Data{}, err
		}
		if data.StornierungID == stornierungID {
			return evt, data, nil
		}
	}

	return event.Event{}, kasse.PfandrueckgabeStorniertV1Data{}, ErrStornierungNichtGefunden
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
		return tseAbschnittAusSignatur(ergebnis.Signatur), escpos.KeinTSEVermerk, false, nil
	case tse.SignaturstatusNachsigniert:
		abschnitt := tseAbschnittAusSignatur(ergebnis.Signatur)
		abschnitt.Nachsigniert = true
		return abschnitt, escpos.KeinTSEVermerk, false, nil
	case tse.SignaturstatusAusfall:
		logging.FromContext(ctx).Warn("Kassenbeleg ohne TSE-Signatur: dokumentierter Ausfall", "event_id", eventID, "ausfall_grund", ergebnis.AusfallGrund)
		return nil, vermerkFuerAusfall(ergebnis.AusfallGrund), false, nil
	case tse.SignaturstatusAusstehend:
		fallthrough
	default:
		return nil, escpos.KeinTSEVermerk, true, nil
	}
}

// vermerkFuerAusfall: fehlende TSE-Konfiguration wird nicht nachsigniert und
// trägt „keine TSE konfiguriert"; jeder andere Ausfall (vorübergehende
// Nichterreichbarkeit) wird nachsigniert.
func vermerkFuerAusfall(ausfallGrund string) escpos.TSEBelegvermerk {
	if ausfallGrund == tse.StatusTSENichtKonfiguriert || ausfallGrund == tse.StoerungGrundKeineKonfiguration {
		return escpos.TSEVermerkKeineKonfiguration
	}
	return escpos.TSEVermerkVoruebergehend
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

// belegQuelle: Positionen and GesamtbetragCents are positive; Negativ negates them with the
// Steuermatrix on the Beleg. Stornobeleg only sets the title and the "Storno zu Bon-Nr" line.
type belegQuelle struct {
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

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
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

	positionen := quelle.Positionen
	gesamtbetragCents := quelle.GesamtbetragCents
	steuermatrix := steuer.Steuermatrix(toSteuermatrixPositionen(positionen))
	if quelle.Negativ {
		steuermatrix = negiereAufteilungen(steuermatrix)
		positionen = negierePositionen(positionen)
		gesamtbetragCents = -gesamtbetragCents
	}

	payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
		Vereinsname:              betreiber.Vereinsname,
		Strasse:                  betreiber.Strasse,
		Plz:                      betreiber.Plz,
		Ort:                      betreiber.Ort,
		KassenSeriennummer:       kassenidentitaet.Seriennummer.String(),
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
	})

	auftrag := druckauftrag_repo.NeuerDruckauftrag{
		ZielIP:   kassenbelegStation.DruckerIP,
		Payload:  base64.StdEncoding.EncodeToString(payload),
		BonArt:   string(druckstation.KategorieKassenbeleg),
		Referenz: quelle.Referenz,
	}

	if err := c.DruckauftragRepo.EnqueueDruckauftraege(ctx, []druckauftrag_repo.NeuerDruckauftrag{auftrag}); err != nil {
		log.Error("Failed to enqueue kassenbeleg", "error", err, "tisch_id", cmd.TischID, "zahlung_id", cmd.ZahlungID, "verkauf_id", cmd.VerkaufID, "rueckgabe_id", cmd.RueckgabeID)
		return "", ErrDatabase
	}

	log.Info("Kassenbeleg queued", "tisch_id", cmd.TischID, "verkauf_id", cmd.VerkaufID, "rueckgabe_id", cmd.RueckgabeID, "event_id", quelle.Event.ID)
	return BelegStatusEingereiht, nil
}

func (c Command) resolveBelegQuelle(ctx context.Context, ks *kasse.Kassensitzung, cmd KassenbelegDruckenCommand) (belegQuelle, error) {
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

func (c Command) resolveDirektverkaufStornobeleg(ctx context.Context, ks *kasse.Kassensitzung, verkaufID, stornierungID string) (belegQuelle, error) {
	log := logging.FromContext(ctx)

	subject := kasse.DirektverkaufSubject(ks.ZNr, verkaufID)
	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error("Failed to read direktverkauf events for stornobeleg", "error", err, "verkauf_id", verkaufID)
		return belegQuelle{}, ErrDatabase
	}

	verkaufEvent, _, err := findDirektverkaufGetaetigtEvent(events, verkaufID)
	if err != nil {
		if errors.Is(err, ErrVerkaufNichtGefunden) {
			log.Warn("Direktverkauf not found for stornobeleg", "verkauf_id", verkaufID)
			return belegQuelle{}, ErrVerkaufNichtGefunden
		}
		log.Error("Failed to decode direktverkauf event data", "error", err, "verkauf_id", verkaufID)
		return belegQuelle{}, ErrDatabase
	}

	stornoEvent, stornoData, err := findDirektverkaufStorniertEvent(events, stornierungID)
	if err != nil {
		if errors.Is(err, ErrStornierungNichtGefunden) {
			log.Warn("Stornierung not found for stornobeleg", "verkauf_id", verkaufID, "stornierung_id", stornierungID)
			return belegQuelle{}, ErrStornierungNichtGefunden
		}
		log.Error("Failed to decode direktverkauf storno event data", "error", err, "verkauf_id", verkaufID, "stornierung_id", stornierungID)
		return belegQuelle{}, ErrDatabase
	}

	return belegQuelle{
		Event:               stornoEvent,
		Positionen:          kasse.PositionenFromEventData(stornoData.Positionen),
		GesamtbetragCents:   stornoData.GesamtStornierungCents,
		Referenz:            fmt.Sprintf("direktverkauf-storniert:%d", stornoEvent.ID),
		Negativ:             true,
		Stornobeleg:         true,
		StornoZuBelegnummer: fmt.Sprintf("%d", verkaufEvent.ID),
	}, nil
}

// resolvePfandbeleg: the Pfandbeleg is a Kassenbeleg with negative amounts, without Storno title.
func (c Command) resolvePfandbeleg(ctx context.Context, ks *kasse.Kassensitzung, rueckgabeID string) (belegQuelle, error) {
	log := logging.FromContext(ctx)

	events, err := c.EventRepo.ReadEventsBySubject(ctx, kasse.PfandrueckgabeSubject(ks.ZNr, rueckgabeID))
	if err != nil {
		log.Error("Failed to read pfandrueckgabe events for kassenbeleg", "error", err, "rueckgabe_id", rueckgabeID)
		return belegQuelle{}, ErrDatabase
	}

	rueckgabeEvent, rueckgabeData, err := findPfandZurueckgegebenEvent(events, rueckgabeID)
	if err != nil {
		if errors.Is(err, ErrRueckgabeNichtGefunden) {
			log.Warn("Pfandrückgabe not found for kassenbeleg", "rueckgabe_id", rueckgabeID)
			return belegQuelle{}, ErrRueckgabeNichtGefunden
		}
		log.Error("Failed to decode pfandrueckgabe event data", "error", err, "rueckgabe_id", rueckgabeID)
		return belegQuelle{}, ErrDatabase
	}

	return belegQuelle{
		Event:             rueckgabeEvent,
		Positionen:        kasse.PositionenFromEventData(rueckgabeData.Positionen),
		GesamtbetragCents: rueckgabeData.GesamtbetragCents,
		Referenz:          fmt.Sprintf("pfand-zurueckgegeben:%d", rueckgabeEvent.ID),
		Negativ:           true,
	}, nil
}

// resolvePfandStornobeleg: cancelling a Rückgabe books money back in, so its Stornobeleg carries
// positive amounts and points to the Rückgabe's Bon-Nr.
func (c Command) resolvePfandStornobeleg(ctx context.Context, ks *kasse.Kassensitzung, rueckgabeID, stornierungID string) (belegQuelle, error) {
	log := logging.FromContext(ctx)

	events, err := c.EventRepo.ReadEventsBySubject(ctx, kasse.PfandrueckgabeSubject(ks.ZNr, rueckgabeID))
	if err != nil {
		log.Error("Failed to read pfandrueckgabe events for stornobeleg", "error", err, "rueckgabe_id", rueckgabeID)
		return belegQuelle{}, ErrDatabase
	}

	rueckgabeEvent, _, err := findPfandZurueckgegebenEvent(events, rueckgabeID)
	if err != nil {
		if errors.Is(err, ErrRueckgabeNichtGefunden) {
			log.Warn("Pfandrückgabe not found for stornobeleg", "rueckgabe_id", rueckgabeID)
			return belegQuelle{}, ErrRueckgabeNichtGefunden
		}
		log.Error("Failed to decode pfandrueckgabe event data", "error", err, "rueckgabe_id", rueckgabeID)
		return belegQuelle{}, ErrDatabase
	}

	stornoEvent, stornoData, err := findPfandrueckgabeStorniertEvent(events, stornierungID)
	if err != nil {
		if errors.Is(err, ErrStornierungNichtGefunden) {
			log.Warn("Stornierung not found for pfand stornobeleg", "rueckgabe_id", rueckgabeID, "stornierung_id", stornierungID)
			return belegQuelle{}, ErrStornierungNichtGefunden
		}
		log.Error("Failed to decode pfandrueckgabe storno event data", "error", err, "rueckgabe_id", rueckgabeID, "stornierung_id", stornierungID)
		return belegQuelle{}, ErrDatabase
	}

	return belegQuelle{
		Event:               stornoEvent,
		Positionen:          kasse.PositionenFromEventData(stornoData.Positionen),
		GesamtbetragCents:   stornoData.GesamtStornierungCents,
		Referenz:            fmt.Sprintf("pfandrueckgabe-storniert:%d", stornoEvent.ID),
		Stornobeleg:         true,
		StornoZuBelegnummer: fmt.Sprintf("%d", rueckgabeEvent.ID),
	}, nil
}

// resolveTischStornobeleg löst den Tisch-Storno-Beleg (Warenrücknahme) auf.
func (c Command) resolveTischStornobeleg(ctx context.Context, ks *kasse.Kassensitzung, tischID int, stornierungID string) (belegQuelle, error) {
	log := logging.FromContext(ctx)

	subject := kasse.TischSessionSubject(ks.ZNr, tischID)
	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error("Failed to read events for stornobeleg", "error", err, "tisch_id", tischID)
		return belegQuelle{}, ErrDatabase
	}

	stornoEvent, stornoData, err := findStornierungEvent(events, stornierungID)
	if err != nil {
		if errors.Is(err, ErrStornierungNichtGefunden) {
			log.Warn("Stornierung not found for stornobeleg", "tisch_id", tischID, "stornierung_id", stornierungID)
			return belegQuelle{}, ErrStornierungNichtGefunden
		}
		log.Error("Failed to decode stornierung event data", "error", err, "tisch_id", tischID, "stornierung_id", stornierungID)
		return belegQuelle{}, ErrDatabase
	}

	zahlungEvent, _, err := findZahlungEvent(events, stornoData.ZahlungID)
	if err != nil {
		if errors.Is(err, ErrZahlungNichtGefunden) {
			log.Warn("Referenzierte Zahlung not found for stornobeleg", "tisch_id", tischID, "zahlung_id", stornoData.ZahlungID)
			return belegQuelle{}, ErrZahlungNichtGefunden
		}
		log.Error("Failed to decode referenzierte zahlung event data", "error", err, "tisch_id", tischID, "zahlung_id", stornoData.ZahlungID)
		return belegQuelle{}, ErrDatabase
	}

	return belegQuelle{
		Event:               stornoEvent,
		Positionen:          kasse.PositionenFromEventData(stornoData.Positionen),
		GesamtbetragCents:   stornoData.GesamtStornierungCents,
		Referenz:            fmt.Sprintf("stornierung-erteilt:%d", stornoEvent.ID),
		Negativ:             true,
		Stornobeleg:         true,
		StornoZuBelegnummer: fmt.Sprintf("%d", zahlungEvent.ID),
	}, nil
}

func (c Command) resolveDirektverkaufBeleg(ctx context.Context, ks *kasse.Kassensitzung, verkaufID string) (belegQuelle, error) {
	log := logging.FromContext(ctx)

	subject := kasse.DirektverkaufSubject(ks.ZNr, verkaufID)
	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error("Failed to read direktverkauf events for kassenbeleg", "error", err, "verkauf_id", verkaufID)
		return belegQuelle{}, ErrDatabase
	}

	verkaufEvent, verkaufData, err := findDirektverkaufGetaetigtEvent(events, verkaufID)
	if err != nil {
		if errors.Is(err, ErrVerkaufNichtGefunden) {
			log.Warn("Direktverkauf not found for kassenbeleg", "verkauf_id", verkaufID)
			return belegQuelle{}, ErrVerkaufNichtGefunden
		}
		log.Error("Failed to decode direktverkauf event data", "error", err, "verkauf_id", verkaufID)
		return belegQuelle{}, ErrDatabase
	}

	return belegQuelle{
		Event:             verkaufEvent,
		Positionen:        kasse.PositionenFromEventData(verkaufData.Positionen),
		GesamtbetragCents: verkaufData.GesamtbetragCents,
		Referenz:          fmt.Sprintf("direktverkauf-getaetigt:%d", verkaufEvent.ID),
	}, nil
}

func (c Command) resolveTischZahlungsbeleg(ctx context.Context, ks *kasse.Kassensitzung, tischID int, zahlungID string) (belegQuelle, error) {
	log := logging.FromContext(ctx)

	subject := kasse.TischSessionSubject(ks.ZNr, tischID)
	state, err := c.EventRepo.ReadTischSession(ctx, subject)
	if err != nil {
		log.Error("Failed to read table projection for kassenbeleg", "error", err, "tisch_id", tischID)
		return belegQuelle{}, ErrDatabase
	}

	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error("Failed to read events for kassenbeleg", "error", err, "tisch_id", tischID)
		return belegQuelle{}, ErrDatabase
	}

	zahlungEvent, zahlungData, err := findZahlungEvent(events, zahlungID)
	if err != nil {
		if errors.Is(err, ErrZahlungNichtGefunden) {
			log.Warn("Zahlung not found for kassenbeleg", "tisch_id", tischID, "zahlung_id", zahlungID)
			return belegQuelle{}, ErrZahlungNichtGefunden
		}
		log.Error("Failed to decode zahlung event data", "error", err, "tisch_id", tischID, "zahlung_id", zahlungID)
		return belegQuelle{}, ErrDatabase
	}

	return belegQuelle{
		Event:                    zahlungEvent,
		Positionen:               kasse.PositionenFromEventData(zahlungData.Positionen),
		GesamtbetragCents:        zahlungData.GesamtZahlungCents,
		Referenz:                 fmt.Sprintf("zahlung-kassiert:%d", zahlungEvent.ID),
		ErsteBestellungZeitpunkt: state.ErsteBestellungLogTime,
	}, nil
}
