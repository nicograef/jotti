import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { signIn, signOut } from '@/test/auth'
import { FakeBackend } from '@/test/FakeBackend'

import type { DirektverkaufHistorieEintrag } from './Direktverkauf'
import { DirektverkaufBackend } from './DirektverkaufBackend'
import { DirektverkaufHistorie } from './DirektverkaufHistorie'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), info: vi.fn() },
}))

// Serviceleitung so the cancellation path (canCancel) applies.
beforeEach(() => {
  signIn({ role: 'serviceleitung' })
})

afterEach(() => {
  cleanup()
  signOut()
})

function backend(): FakeBackend {
  return new FakeBackend()
    .respond('serviceleitung/direktverkauf-stornieren', {})
    .respond('service/beleg-drucken', { status: 'eingereiht' })
}

const verkaufId = '00000000-0000-4000-8000-000000000001'
const positionId = '00000000-0000-4000-8000-000000000002'
const stornierungId = '00000000-0000-4000-8000-000000000003'

const verkauf: DirektverkaufHistorieEintrag = {
  verkaufId,
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
          positionId: '00000000-0000-4000-8000-000000000004',
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
        backend={new DirektverkaufBackend(backend())}
        onErfolg={vi.fn()}
      />,
    )

    expect(screen.getByText('2× Cola 0,5l, 1× Brezel')).toBeInTheDocument()
  })

  it('cancels selected positions with exactly one backend call', async () => {
    const user = userEvent.setup()
    const fake = backend()
    const onErfolg = vi.fn()
    render(
      <DirektverkaufHistorie
        historie={[verkauf]}
        historieLoading={false}
        backend={new DirektverkaufBackend(fake)}
        onErfolg={onErfolg}
      />,
    )

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
      expect(
        fake.bodies('serviceleitung/direktverkauf-stornieren'),
      ).toHaveLength(1)
    })
    expect(fake.bodies('serviceleitung/direktverkauf-stornieren')).toEqual([
      {
        verkaufId,
        positionen: [{ positionId, menge: 1 }],
        kommentar: 'Rückgabe',
      },
    ])
    // Statt sofortigem Refetch meldet der Storno den Erfolg über den Pop-Text;
    // der Refetch folgt beim Schließen des Pops (DirektverkaufPage).
    expect(onErfolg).toHaveBeenCalledWith('Stornierung gebucht.')
  })

  it('triggers kassenbeleg print for a sale with exactly one backend call', async () => {
    const user = userEvent.setup()
    const fake = backend()
    render(
      <DirektverkaufHistorie
        historie={[verkauf]}
        historieLoading={false}
        backend={new DirektverkaufBackend(fake)}
        onErfolg={vi.fn()}
      />,
    )

    await user.click(screen.getByRole('button', { name: /Verkauf/ }))
    await user.click(
      screen.getByRole('button', { name: 'Kassenbeleg drucken' }),
    )

    await waitFor(() => {
      expect(fake.bodies('service/beleg-drucken')).toHaveLength(1)
    })
    expect(fake.bodies('service/beleg-drucken')).toEqual([{ verkaufId }])
  })

  it('triggers stornobeleg print with the stornierungId of the cancellation', async () => {
    const user = userEvent.setup()
    const fake = backend()
    const stornierterVerkauf: DirektverkaufHistorieEintrag = {
      ...verkauf,
      offenePositionen: [],
      gesamtStorniertCents: 1000,
      stornierungen: [
        {
          stornierungId,
          storniertAm: '2026-06-08T11:00:00Z',
          gesamtStornierungCents: 1000,
        },
      ],
    }
    render(
      <DirektverkaufHistorie
        historie={[stornierterVerkauf]}
        historieLoading={false}
        backend={new DirektverkaufBackend(fake)}
        onErfolg={vi.fn()}
      />,
    )

    await user.click(screen.getByRole('button', { name: /Verkauf/ }))
    await user.click(
      screen.getByRole('button', { name: 'Stornobeleg drucken' }),
    )

    await waitFor(() => {
      expect(fake.bodies('service/beleg-drucken')).toHaveLength(1)
    })
    expect(fake.bodies('service/beleg-drucken')).toEqual([
      { verkaufId, stornierungId },
    ])
  })
})
