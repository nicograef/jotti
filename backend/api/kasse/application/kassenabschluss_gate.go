package application

import (
	"context"
	"time"

	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
)

// SignaturenAusstehendError blockiert den Kassenabschluss: Mindestens ein
// Signaturauftrag der Kassensitzung ist noch offen und keine Stoerung erklaert
// den Ausfall (Ergebnis ausstehend der Signaturstatus-Funktion). Die TSE holt
// in Kuerze auf; die Abschluss-Operation wird unveraendert erneut angefordert.
type SignaturenAusstehendError struct {
	Anzahl              int
	AeltesterErstelltAm time.Time
}

func (e *SignaturenAusstehendError) Error() string {
	return "signaturen ausstehend"
}

// KassenabschlussErgebnis meldet die beim Abschluss verbliebenen Ausfall-Reste.
// Sie blockieren den Abschluss nicht (die Signaturstatus-Funktion rechnet sie
// dem Ausfall zu), werden aber in der Abschlussmeldung ausgewiesen.
type KassenabschlussErgebnis struct {
	// AusfallResteAnzahl: endgueltig fehlgeschlagene/verworfene Auftraege sowie
	// offene Auftraege waehrend eines aktiven Stoerungszeitraums; werden nach
	// Rueckkehr der TSE nachsigniert.
	AusfallResteAnzahl int
	// OhneKonfigurationAnzahl: Vorgaenge ohne TSE-Signatur, weil keine TSE
	// konfiguriert ist (tse_nicht_konfiguriert); werden nicht nachsigniert.
	OhneKonfigurationAnzahl int
}

// signaturGate ist das interne Urteil des Gates ueber die noch nicht erledigten
// Signaturauftraege der Kassensitzung.
type signaturGate struct {
	ausstehendAnzahl        int
	aeltesterAusstehend     time.Time
	ausfallResteAnzahl      int
	ohneKonfigurationAnzahl int
}

// pruefeSignaturGate klassifiziert jeden noch nicht erledigten Signaturauftrag
// der Kassensitzung ueber die Signaturstatus-Funktion — dieselbe Zurechnung wie
// beim Beleg-Abruf, kein zweiter Zurechnungspfad. Ergebnis ausstehend blockiert
// (frischer offener Auftrag ohne Stoerung), Ausfall laesst durch und wird in der
// Abschlussmeldung ausgewiesen (Ausfall-Rest bzw. fehlende TSE-Konfiguration).
func (c Command) pruefeSignaturGate(ctx context.Context, kassensitzungNr int) (signaturGate, error) {
	log := zerolog.Ctx(ctx)

	staende, err := c.TSERepo.GetOffeneSignaturauftragStaendeFuerKassensitzung(ctx, kassensitzungNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", kassensitzungNr).Msg("Failed to load Signaturauftrag-Staende for Kassenabschluss-Gate")
		return signaturGate{}, ErrDatabase
	}

	aktiveStoerung, err := c.TSERepo.GetAktiveTSEStoerung(ctx)
	if err != nil {
		log.Error().Err(err).Int("z_nr", kassensitzungNr).Msg("Failed to load aktive TSE-Stoerung for Kassenabschluss-Gate")
		return signaturGate{}, ErrDatabase
	}

	var gate signaturGate
	for _, stand := range staende {
		ergebnis := tse.BestimmeSignaturstatus(stand, aktiveStoerung)
		switch ergebnis.Status {
		case tse.SignaturstatusAusstehend:
			if gate.ausstehendAnzahl == 0 || stand.ErstelltAm.Before(gate.aeltesterAusstehend) {
				gate.aeltesterAusstehend = stand.ErstelltAm
			}
			gate.ausstehendAnzahl++
		case tse.SignaturstatusAusfall:
			if ergebnis.AusfallGrund == tse.StatusTSENichtKonfiguriert || ergebnis.AusfallGrund == tse.StoerungGrundKeineKonfiguration {
				gate.ohneKonfigurationAnzahl++
			} else {
				gate.ausfallResteAnzahl++
			}
		}
	}
	return gate, nil
}
