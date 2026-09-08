package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
)

// SignaturenAusstehendError blockiert den Kassenabschluss: Mindestens ein
// Signaturauftrag der Kassensitzung ist noch offen und keine Störung erklärt
// den Ausfall (Ergebnis ausstehend der Signaturstatus-Funktion). Die TSE holt
// in Kürze auf; die Abschluss-Operation wird unverändert erneut angefordert.
type SignaturenAusstehendError struct {
	Anzahl int
}

func (e *SignaturenAusstehendError) Error() string {
	return "signaturen ausstehend"
}

// KassenabschlussErgebnis meldet die beim Abschluss verbliebenen Ausfall-Reste.
// Sie blockieren den Abschluss nicht (die Signaturstatus-Funktion rechnet sie
// dem Ausfall zu), werden aber in der Abschlussmeldung ausgewiesen.
type KassenabschlussErgebnis struct {
	// AusfallResteAnzahl: endgültig fehlgeschlagene Aufträge sowie offene
	// Aufträge während eines aktiven Störungszeitraums; werden nach Rückkehr
	// der TSE nachsigniert.
	AusfallResteAnzahl int
	// OhneKonfigurationAnzahl: Vorgänge ohne TSE-Signatur, weil keine TSE
	// konfiguriert ist (tse_nicht_konfiguriert); werden nicht nachsigniert.
	OhneKonfigurationAnzahl int
}

// signaturGate ist das interne Urteil des Gates über die noch nicht erledigten
// Signaturaufträge der Kassensitzung.
type signaturGate struct {
	ausstehendAnzahl        int
	ausfallResteAnzahl      int
	ohneKonfigurationAnzahl int
}

// checkSignaturGate klassifiziert jeden noch nicht erledigten Signaturauftrag
// der Kassensitzung über die Signaturstatus-Funktion — dieselbe Zurechnung wie
// beim Beleg-Abruf, kein zweiter Zurechnungspfad. Ergebnis ausstehend blockiert
// (frischer offener Auftrag ohne Störung), Ausfall lässt durch und wird in der
// Abschlussmeldung ausgewiesen (Ausfall-Rest bzw. fehlende TSE-Konfiguration).
func (c Command) checkSignaturGate(ctx context.Context, kassensitzungNr int) (signaturGate, error) {
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
		ergebnis := tse.DetermineSignaturstatus(stand, aktiveStoerung)
		switch ergebnis.Status {
		case tse.SignaturstatusAusstehend:
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
