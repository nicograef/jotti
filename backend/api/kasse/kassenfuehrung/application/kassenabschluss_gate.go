package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
)

// SignaturenAusstehendError blockiert den Kassenabschluss: mindestens ein Signaturauftrag ist offen
// und keine Störung erklärt den Ausfall. Die Operation wird unverändert erneut angefordert.
type SignaturenAusstehendError struct {
	Anzahl int
}

func (e *SignaturenAusstehendError) Error() string {
	return "signaturen ausstehend"
}

// KassenabschlussErgebnis meldet die verbliebenen Ausfall-Reste; sie blockieren den Abschluss nicht,
// werden aber in der Abschlussmeldung ausgewiesen.
type KassenabschlussErgebnis struct {
	// AusfallResteAnzahl: endgültig fehlgeschlagene sowie während eines Störungszeitraums offene
	// Aufträge; nur die offenen werden nach Rückkehr der TSE nachsigniert.
	AusfallResteAnzahl int
	// OhneKonfigurationAnzahl: Vorgänge ohne Signatur mangels TSE-Konfiguration
	// (tse_nicht_konfiguriert); werden nicht nachsigniert.
	OhneKonfigurationAnzahl int
}

type signaturGate struct {
	ausstehendAnzahl        int
	ausfallResteAnzahl      int
	ohneKonfigurationAnzahl int
}

// checkSignaturGate klassifiziert jeden offenen Signaturauftrag über dieselbe
// Signaturstatus-Funktion wie der Beleg-Abruf — kein zweiter Zurechnungspfad. Ausstehend blockiert,
// Ausfall lässt durch und wird in der Abschlussmeldung ausgewiesen.
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
