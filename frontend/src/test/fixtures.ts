import type { Produkt } from '@/lib/produktSchemas'
import type { Position } from '@/service/table/Bestellung'
import type { UnbezahltePosition } from '@/service/table/Tisch'

// Ein Produkt mit einer aktiven Variante „Normal" zu 3,50 €.
export const testProdukt: Produkt = {
  id: 1,
  name: 'Bratwurst',
  kategorie: 'essen',
  steuersatz: 'ermaessigt',
  pfand: false,
  status: 'active',
  varianten: [
    {
      id: 1,
      name: 'Normal',
      preisCents: 350,
      status: 'active',
      pfandVarianteId: null,
    },
  ],
}

// Eine unbezahlte Position „Bratwurst Normal" zu 3,50 €, bestellt von User 1.
export function testPosition(overrides: Partial<Position> = {}): Position {
  return {
    positionId: '00000000-0000-4000-8000-000000000001',
    varianteId: 1,
    produktName: 'Bratwurst',
    varianteName: 'Normal',
    kategorie: 'essen',
    steuersatz: 'regel',
    einzelpreisCents: 350,
    menge: 1,
    bestellerUserId: 1,
    bestellerName: 'Tester',
    ...overrides,
  }
}

// testPosition aus Sicht der anfragenden Servicekraft; ohne Override ihre eigene.
export function testUnbezahltePosition(
  overrides: Partial<UnbezahltePosition> = {},
): UnbezahltePosition {
  return { ...testPosition(overrides), eigene: true, ...overrides }
}
