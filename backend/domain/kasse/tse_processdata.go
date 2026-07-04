package kasse

import (
	"fmt"
	"strings"
	"time"

	"github.com/nicograef/jotti/backend/domain/steuer"
)

const zahlungsartBar = "Bar"

// BuildKassenbelegProcessData erzeugt Kassenbeleg-V1-processData nach
// DSFinV-K Anhang I: Bruttobetraege je Steuersatz plus Zahlungsteil. faktor -1
// stellt Stornierungen dar: alle Steuerbetraege werden negiert.
func BuildKassenbelegProcessData(positionen []Position, zahlbetragCents int, faktor int) (string, error) {
	if faktor != 1 && faktor != -1 {
		return "", fmt.Errorf("invalid faktor %d", faktor)
	}

	var betragNormalCents int
	var betragErmaessigtCents int
	var betragBefreitCents int

	for _, pos := range positionen {
		basisBrutto := pos.Einzelpreis * pos.Menge
		aufteilungen := steuer.Aufteilen(basisBrutto, steuer.Steuersatz(pos.Steuersatz))
		if len(aufteilungen) == 0 {
			return "", fmt.Errorf("unsupported steuersatz %q", pos.Steuersatz)
		}
		for _, aufteilung := range aufteilungen {
			brutto := aufteilung.Brutto * faktor
			switch aufteilung.Satz {
			case steuer.RegelSteuersatz:
				betragNormalCents += brutto
			case steuer.ErmaessigtSteuersatz:
				betragErmaessigtCents += brutto
			case steuer.BefreitSteuersatz:
				betragBefreitCents += brutto
			default:
				return "", fmt.Errorf("unsupported steuersatz in aufteilung %q", aufteilung.Satz)
			}
		}
	}

	// DSFinV-K Anhang I: Zahlungen von 0.00 müssen entfallen.
	zahlungen := ""
	if zahlbetragCents != 0 {
		zahlungen = betragString(zahlbetragCents) + ":" + zahlungsartBar
	}

	return fmt.Sprintf(
		"Beleg^%s_%s_%s_%s_%s^%s",
		betragString(betragNormalCents),
		betragString(betragErmaessigtCents),
		betragString(0),
		betragString(0),
		betragString(betragBefreitCents),
		zahlungen,
	), nil
}

// BuildBestellungProcessData erzeugt die CSV-Darstellung nach
// DSFinV-K Anhang I: pro Position `<Menge>;"<Bezeichnung>";<Brutto-Einzelpreis>`,
// Zeilentrenner \r, Anführungszeichen in der Bezeichnung werden verdoppelt.
// faktor -1 stellt Rücknahmen dar (geldneutrale Korrektur, Abgang einer
// Umbuchung) — DSFinV-K Anhang I sieht für Bestell-Storni negative Mengen
// vor. Ohne Vorzeichen wäre eine Rücknahme TSE-seitig von einer zusätzlichen
// Neubestellung nicht unterscheidbar.
func BuildBestellungProcessData(positionen []Position, faktor int) (string, error) {
	if len(positionen) == 0 {
		return "", fmt.Errorf("bestellung processData requires at least one position")
	}
	if faktor != 1 && faktor != -1 {
		return "", fmt.Errorf("unsupported faktor %d", faktor)
	}

	zeilen := make([]string, 0, len(positionen))
	for _, pos := range positionen {
		// Eine nicht-positive Menge wäre das Symptom eines Fehlers im Aufrufer —
		// hart melden statt still zu verschlucken (Vollständigkeit der Absicherung).
		if pos.Menge <= 0 {
			return "", fmt.Errorf("bestellung processData requires positive quantities, got %d for %q", pos.Menge, pos.Bezeichnung())
		}

		name := pos.Bezeichnung()
		if name == "" {
			name = "Unbekannt"
		}

		bezeichnung := `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
		zeilen = append(zeilen, fmt.Sprintf("%d;%s;%s", faktor*pos.Menge, bezeichnung, betragString(pos.Einzelpreis)))
	}

	return strings.Join(zeilen, "\r"), nil
}

// BuildGeldtransitProcessData bildet Einlage/Entnahme als Eigenbeleg ab:
// Einlagen mit positivem, Entnahmen mit negativem Zahlbetrag.
func BuildGeldtransitProcessData(richtung string, betragCents int) (string, error) {
	switch richtung {
	case "einlage":
		return BuildEigenbelegProcessData(betragCents), nil
	case "entnahme":
		return BuildEigenbelegProcessData(-betragCents), nil
	default:
		return "", fmt.Errorf("unsupported richtung %q", richtung)
	}
}

// BuildEigenbelegProcessData erzeugt Kassenbeleg-V1-processData für USt-neutrale
// Bargeldbewegungen (Eigenbelege nach AEAO 2.2.3.6.1, z. B. Geldtransit und
// Kassendifferenz). Der Betrag steht im 0-%-Feld (Feld 5, UST_SCHLUESSEL 5 =
// nicht steuerbar) und gleicht so die Bar-Zahlung aus; das Vorzeichen folgt dem
// übergebenen Bargeldbetrag (Abfluss negativ). DSFinV-K Anhang I: Zahlungen von
// 0.00 müssen entfallen.
func BuildEigenbelegProcessData(zahlbetragCents int) string {
	zahlungen := ""
	if zahlbetragCents != 0 {
		zahlungen = betragString(zahlbetragCents) + ":" + zahlungsartBar
	}
	return fmt.Sprintf("Beleg^0.00_0.00_0.00_0.00_%s^%s", betragString(zahlbetragCents), zahlungen)
}

// BuildTagesabschlussProcessData erzeugt SonstigerVorgang-processData fuer den
// Tagesabschluss (Z-Bon): Z-Nummer plus Abschlusszeitraum.
func BuildTagesabschlussProcessData(zNr int, zeitraumVon time.Time, zeitraumBis time.Time) string {
	return fmt.Sprintf(
		"Tagesabschluss^ZNr:%d^Von:%s^Bis:%s",
		zNr,
		zeitraumVon.UTC().Format(time.RFC3339),
		zeitraumBis.UTC().Format(time.RFC3339),
	)
}

func betragString(cents int) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}
