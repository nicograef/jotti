// Zählhilfe: die Nennwerte des Euro-Bargelds (Münzen und Scheine) in Cent,
// aufsteigend von 1 ct bis 200 €. Grundlage für die Stückzahl-Erfassung beim
// Kassenabschluss.
export const NENNWERTE_CENTS = [
  1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000, 20000,
] as const

export type Nennwert = (typeof NENNWERTE_CENTS)[number]

// summeAusStueckzahlen summiert die gezählten Stückzahlen je Nennwert zu einem
// Gesamtbetrag in Cent. Reine Funktion ohne Seiteneffekte: Schlüssel ist der
// Nennwert in Cent, Wert die Stückzahl. Fehlende, negative oder nicht-ganze
// Stückzahlen zählen als 0 (die Eingabe erlaubt strukturell nur ganze Zahlen
// ≥ 0; diese Absicherung hält die Funktion trotzdem robust und testbar).
export function summeAusStueckzahlen(
  stueckzahlen: Partial<Record<Nennwert, number>>,
): number {
  return NENNWERTE_CENTS.reduce((summe, nennwert) => {
    const anzahl = stueckzahlen[nennwert]
    if (
      typeof anzahl !== 'number' ||
      !Number.isInteger(anzahl) ||
      anzahl <= 0
    ) {
      return summe
    }
    return summe + nennwert * anzahl
  }, 0)
}
