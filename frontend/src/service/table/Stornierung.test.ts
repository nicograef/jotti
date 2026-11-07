import { describe, expect, it } from 'vitest'

import { StornierungSchema } from './Stornierung'

const stornierungResponse = {
  art: 'stornierung',
  id: '11111111-1111-4111-8111-111111111111',
  userId: 1,
  userName: 'Clara',
  tischId: 1,
  positionen: [
    {
      positionId: '22222222-2222-4222-8222-222222222222',
      varianteId: 1,
      produktName: 'Bratwurst',
      varianteName: 'Normal',
      kategorie: 'essen',
      steuersatz: 'regel',
      einzelpreisCents: 350,
      menge: 1,
      bestellerUserId: 1,
      bestellerName: 'Anna',
    },
  ],
  gesamtStornierungCents: 350,
  barRueckgabe: false,
  storniertAm: '2026-06-18T12:10:00Z',
}

describe('StornierungSchema (Response)', () => {
  it('akzeptiert eine geldneutrale Korrektur mit leerem Kommentar', () => {
    const result = StornierungSchema.safeParse({
      ...stornierungResponse,
      kommentar: '',
    })

    expect(result.success).toBe(true)
  })
})
