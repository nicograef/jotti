import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { DirektverkaufHistorieEintrag } from '../../direktverkauf/Direktverkauf'
import { DirektverkaufHistorie } from './DirektverkaufHistorie'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), info: vi.fn() },
}))

vi.mock('@/lib/Auth', () => ({
  AuthSingleton: { canCancel: true },
}))

afterEach(() => {
  cleanup()
})

const positionId = '22222222-2222-2222-2222-222222222222'

const verkauf: DirektverkaufHistorieEintrag = {
  verkaufId: '11111111-1111-1111-1111-111111111111',
  userName: 'Anna',
  getaetigtAm: '2026-06-08T10:00:00Z',
  positionen: [
    {
      positionId,
      varianteId: 1,
      produktName: 'Cola',
      varianteName: '0,5l',
      kategorie: 'getraenk',
      steuersatz: 'regel',
      einzelpreisCents: 500,
      menge: 2,
    },
  ],
  gesamtbetragCents: 1000,
  kommentar: '',
  offenePositionen: [
    {
      positionId,
      varianteId: 1,
      produktName: 'Cola',
      varianteName: '0,5l',
      kategorie: 'getraenk',
      steuersatz: 'regel',
      einzelpreisCents: 500,
      menge: 2,
    },
  ],
  gesamtStorniertCents: 0,
  stornierungen: [],
}

describe('DirektverkaufHistorie', () => {
  it('shows the position summary as a sub-line on each sale row', () => {
    const mehrpositionenVerkauf: DirektverkaufHistorieEintrag = {
      ...verkauf,
      positionen: [
        {
          positionId,
          varianteId: 1,
          produktName: 'Cola',
          varianteName: '0,5l',
          kategorie: 'getraenk',
          steuersatz: 'regel',
          einzelpreisCents: 500,
          menge: 2,
        },
        {
          positionId: '44444444-4444-4444-4444-444444444444',
          varianteId: 2,
          produktName: 'Brezel',
          varianteName: '',
          kategorie: 'essen',
          steuersatz: 'ermaessigt',
          einzelpreisCents: 150,
          menge: 1,
        },
      ],
    }
    render(
      <DirektverkaufHistorie
        historie={[mehrpositionenVerkauf]}
        historieLoading={false}
        backend={{
          direktverkaufStornieren: vi.fn().mockResolvedValue(undefined),
          kassenbelegDrucken: vi.fn().mockResolvedValue('eingereiht'),
        }}
        onErfolg={vi.fn()}
      />,
    )

    expect(screen.getByText('2× Cola 0,5l, 1× Brezel')).toBeInTheDocument()
  })

  it('cancels selected positions with exactly one backend call', async () => {
    const user = userEvent.setup()
    const direktverkaufStornieren = vi.fn().mockResolvedValue(undefined)
    const kassenbelegDrucken = vi.fn().mockResolvedValue('eingereiht')
    const onErfolg = vi.fn()
    render(
      <DirektverkaufHistorie
        historie={[verkauf]}
        historieLoading={false}
        backend={{ direktverkaufStornieren, kassenbelegDrucken }}
        onErfolg={onErfolg}
      />,
    )

    // Zeile antippen → Detail-Drawer, dort Stornieren…
    await user.click(screen.getByRole('button', { name: /Verkauf/ }))
    await user.click(screen.getByRole('button', { name: /Stornieren…/ }))
    await user.click(
      screen.getByRole('button', { name: 'Cola 0,5l hinzufügen' }),
    )
    await user.type(
      screen.getByPlaceholderText('Kommentar (erforderlich)'),
      'Rückgabe',
    )
    await user.click(
      screen.getByRole('button', { name: /Stornierung erteilen/ }),
    )

    await waitFor(() => {
      expect(direktverkaufStornieren).toHaveBeenCalledTimes(1)
    })
    expect(direktverkaufStornieren).toHaveBeenCalledWith({
      verkaufId: '11111111-1111-1111-1111-111111111111',
      positionen: [{ positionId, menge: 1 }],
      kommentar: 'Rückgabe',
    })
    // Statt sofortigem Refetch meldet der Storno den Erfolg über den Pop-Text;
    // der Refetch folgt beim Schließen des Pops (DirektverkaufPage).
    expect(onErfolg).toHaveBeenCalledWith('Stornierung gebucht.')
  })

  it('triggers kassenbeleg print for a sale with exactly one backend call', async () => {
    const user = userEvent.setup()
    const kassenbelegDrucken = vi.fn().mockResolvedValue('eingereiht')
    render(
      <DirektverkaufHistorie
        historie={[verkauf]}
        historieLoading={false}
        backend={{
          direktverkaufStornieren: vi.fn().mockResolvedValue(undefined),
          kassenbelegDrucken,
        }}
        onErfolg={vi.fn()}
      />,
    )

    await user.click(screen.getByRole('button', { name: /Verkauf/ }))
    await user.click(
      screen.getByRole('button', { name: 'Kassenbeleg drucken' }),
    )

    await waitFor(() => {
      expect(kassenbelegDrucken).toHaveBeenCalledTimes(1)
    })
    expect(kassenbelegDrucken).toHaveBeenCalledWith({
      verkaufId: '11111111-1111-1111-1111-111111111111',
    })
  })

  it('triggers stornobeleg print with the stornierungId of the cancellation', async () => {
    const user = userEvent.setup()
    const kassenbelegDrucken = vi.fn().mockResolvedValue('eingereiht')
    const stornierterVerkauf: DirektverkaufHistorieEintrag = {
      ...verkauf,
      offenePositionen: [],
      gesamtStorniertCents: 1000,
      stornierungen: [
        {
          stornierungId: '33333333-3333-3333-3333-333333333333',
          storniertAm: '2026-06-08T11:00:00Z',
          gesamtStornierungCents: 1000,
        },
      ],
    }
    render(
      <DirektverkaufHistorie
        historie={[stornierterVerkauf]}
        historieLoading={false}
        backend={{
          direktverkaufStornieren: vi.fn().mockResolvedValue(undefined),
          kassenbelegDrucken,
        }}
        onErfolg={vi.fn()}
      />,
    )

    await user.click(screen.getByRole('button', { name: /Verkauf/ }))
    await user.click(
      screen.getByRole('button', { name: 'Stornobeleg drucken' }),
    )

    await waitFor(() => {
      expect(kassenbelegDrucken).toHaveBeenCalledTimes(1)
    })
    expect(kassenbelegDrucken).toHaveBeenCalledWith({
      verkaufId: '11111111-1111-1111-1111-111111111111',
      stornierungId: '33333333-3333-3333-3333-333333333333',
    })
  })
})
