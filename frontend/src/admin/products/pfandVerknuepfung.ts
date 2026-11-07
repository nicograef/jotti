import { EntityStatus } from '@/lib/entityStatus'
import type { Produkt } from '@/lib/produktSchemas'
import { formatEuro, formatPositionName } from '@/lib/utils'

export interface PfandOption {
  varianteId: number
  label: string
}

// Die Varianten aller Pfand-Produkte mit dem Steuersatz des Produkts; ein
// Pfand-Produkt bringt selbst kein Pfand mit und bekommt keine Option. Das
// Backend prüft dieselben Regeln beim Speichern.
export function pfandOptionen(
  katalog: Produkt[],
  produkt: Produkt,
): PfandOption[] {
  if (produkt.pfand) return []
  return katalog
    .filter((p) => p.pfand && p.steuersatz === produkt.steuersatz)
    .flatMap((p) =>
      p.varianten.map((v) => {
        const inaktiv =
          p.status !== EntityStatus.ACTIVE || v.status !== EntityStatus.ACTIVE
        const label = `${formatPositionName(p.name, v.name)} · ${formatEuro(v.preisCents)}`
        return {
          varianteId: v.id,
          label: inaktiv ? `${label} (inaktiv)` : label,
        }
      }),
    )
}

// Ob eine Variante des Katalogs eine der Varianten als Pfand verknüpft. Das
// Backend entfernt solche Verknüpfungen beim Löschen; der Lösch-Dialog nennt es.
export function verliertPfand(
  katalog: Produkt[],
  varianteIds: number[],
): boolean {
  return katalog.some((p) =>
    p.varianten.some(
      (v) =>
        v.pfandVarianteId !== null && varianteIds.includes(v.pfandVarianteId),
    ),
  )
}
