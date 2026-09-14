export const NENNWERTE_CENTS = [
  1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000, 20000,
] as const

export type Nennwert = (typeof NENNWERTE_CENTS)[number]

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
