import type { Produkt } from '@/lib/produktSchemas'

export interface PfandZiel {
  varianteId: number
  preisCents: number
}

export type PfandSchritt = 'plus' | 'minus'

// Je verknüpfender Variante ihr Pfand; ein Verweis, dessen Ziel nicht in `produkte` steht, fällt weg.
export function pfandVerknuepfungen(
  produkte: Produkt[],
): Map<number, PfandZiel> {
  const preise = new Map<number, number>()
  for (const produkt of produkte) {
    for (const variante of produkt.varianten) {
      preise.set(variante.id, variante.preisCents)
    }
  }

  const verknuepfungen = new Map<number, PfandZiel>()
  for (const produkt of produkte) {
    for (const variante of produkt.varianten) {
      if (variante.pfandVarianteId === null) continue
      const preisCents = preise.get(variante.pfandVarianteId)
      if (preisCents === undefined) continue
      verknuepfungen.set(variante.id, {
        varianteId: variante.pfandVarianteId,
        preisCents,
      })
    }
  }
  return verknuepfungen
}

// Ein Schritt am Warenkorb: eine verknüpfte Variante zieht ihr Pfand je Stück mit, das Pfand bleibt nie unter null.
// Minus an einer Variante auf null ändert nichts, auch nicht am Pfand.
export function mitPfand(
  mengen: Record<number, number>,
  varianteId: number,
  schritt: PfandSchritt,
  verknuepfungen: Map<number, PfandZiel>,
): Record<number, number> {
  const menge = mengen[varianteId] || 0
  const pfand = verknuepfungen.get(varianteId)

  if (schritt === 'plus') {
    const nachher = { ...mengen, [varianteId]: menge + 1 }
    if (pfand) {
      nachher[pfand.varianteId] = (nachher[pfand.varianteId] || 0) + 1
    }
    return nachher
  }

  if (menge <= 0) return mengen
  const nachher = { ...mengen, [varianteId]: menge - 1 }
  if (pfand) {
    const pfandMenge = nachher[pfand.varianteId] || 0
    if (pfandMenge > 0) nachher[pfand.varianteId] = pfandMenge - 1
  }
  return nachher
}
