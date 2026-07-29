import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

import type { DirektverkaufHistorieEintrag } from '../../direktverkauf/Direktverkauf'
import { DirektverkaufStornoDrawer } from './DirektverkaufStornoDrawer'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
})

afterEach(() => {
  cleanup()
})

const position = {
  positionId: '22222222-2222-2222-2222-222222222222',
  varianteId: 1,
  produktName: 'Cola',
  varianteName: '0,5l',
  kategorie: 'getraenk' as const,
  steuersatz: 'regel' as const,
  einzelpreisCents: 500,
  menge: 2,
}

const verkauf: DirektverkaufHistorieEintrag = {
  verkaufId: '11111111-1111-1111-1111-111111111111',
  userName: 'Anna',
  getaetigtAm: '2026-06-08T10:00:00Z',
  positionen: [position],
  gesamtbetragCents: 1000,
  kommentar: '',
  offenePositionen: [position],
  gesamtStorniertCents: 0,
  stornierungen: [],
}

describe('DirektverkaufStornoDrawer im Vorgangs-Register', () => {
  it('meldet den getippten Kommentar zusätzlich zur Positionsauswahl', async () => {
    const user = userEvent.setup()
    const { unmount } = render(
      <DirektverkaufStornoDrawer
        backend={{
          direktverkaufStornieren: vi.fn().mockResolvedValue(undefined),
        }}
        verkauf={verkauf}
        onClose={vi.fn()}
        onStorniert={vi.fn()}
      />,
    )

    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)

    await user.type(
      screen.getByPlaceholderText('Kommentar (erforderlich)'),
      'Falsch kassiert',
    )
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    // Die Positionsauswahl meldet über useMengen einen zweiten Vorgang.
    await user.click(screen.getByRole('button', { name: /hinzufügen/ }))
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(2)

    unmount()
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})
