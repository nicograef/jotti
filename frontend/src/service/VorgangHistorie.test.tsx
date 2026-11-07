import {
  cleanup,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { signIn, signOut } from '@/test/auth'

import { type VorgangEintrag, VorgangHistorie } from './VorgangHistorie'

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

function aktionen() {
  return {
    stornieren: vi.fn().mockResolvedValue(undefined),
    belegDrucken: vi.fn().mockResolvedValue('eingereiht'),
  }
}

function renderHistorie(
  eintraege: VorgangEintrag[],
  aktion = aktionen(),
  onErfolg = vi.fn(),
) {
  render(
    <VorgangHistorie
      eintraege={eintraege}
      loading={false}
      bezeichnung="Verkauf"
      leerHinweis="Noch keine Direktverkäufe in dieser Kassensitzung."
      stornieren={aktion.stornieren}
      belegDrucken={aktion.belegDrucken}
      onErfolg={onErfolg}
    />,
  )
  return aktion
}

const verkaufId = '00000000-0000-4000-8000-000000000001'
const positionId = '00000000-0000-4000-8000-000000000002'
const stornierungId = '00000000-0000-4000-8000-000000000003'

const verkauf: VorgangEintrag = {
  id: verkaufId,
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
      kommentar: '',
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

describe('VorgangHistorie', () => {
  it('shows the position summary as a sub-line on each sale row', () => {
    const mehrpositionenVerkauf: VorgangEintrag = {
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
          kommentar: '',
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
          kommentar: '',
        },
      ],
    }
    renderHistorie([mehrpositionenVerkauf])

    expect(screen.getByText('2× Cola 0,5l, 1× Brezel')).toBeInTheDocument()
  })

  it('zeigt den Positionskommentar unter seiner Position im Detail', async () => {
    const user = userEvent.setup()
    renderHistorie([
      {
        ...verkauf,
        positionen: [
          {
            ...verkauf.positionen[0],
            produktName: 'Bratwurst',
            varianteName: 'mit Brot',
            kommentar: '1 davon ohne Senf',
          },
          {
            ...verkauf.positionen[0],
            positionId: '00000000-0000-4000-8000-000000000004',
            produktName: 'Bier',
            varianteName: '0,5l',
            kommentar: '',
          },
        ],
      },
    ])

    await user.click(screen.getByRole('button', { name: /Verkauf/ }))

    const dialog = screen.getByRole('dialog')
    const kommentar = within(dialog).getByText('„1 davon ohne Senf“')
    const zeile = kommentar.parentElement
    expect(zeile).toHaveTextContent(/Bratwurst mit Brot/)
    expect(zeile).not.toHaveTextContent(/Bier/)
    expect(within(dialog).getAllByText(/„/)).toHaveLength(1)
  })

  it('cancels selected positions with exactly one stornieren call', async () => {
    const user = userEvent.setup()
    const onErfolg = vi.fn()
    const aktion = renderHistorie([verkauf], aktionen(), onErfolg)

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
      expect(aktion.stornieren).toHaveBeenCalledTimes(1)
    })
    expect(aktion.stornieren).toHaveBeenCalledWith(
      verkaufId,
      // oxlint-disable-next-line typescript/no-unsafe-argument
      expect.stringMatching(/^[0-9a-f-]{36}$/),
      [{ positionId, menge: 1 }],
      'Rückgabe',
    )
    // Statt sofortigem Refetch meldet der Storno den Erfolg über den Pop-Text;
    // der Refetch folgt beim Schließen des Pops auf der Seite.
    expect(onErfolg).toHaveBeenCalledWith('Stornierung gebucht.')
  })

  it('triggers kassenbeleg print for a sale with exactly one belegDrucken call', async () => {
    const user = userEvent.setup()
    const aktion = renderHistorie([verkauf])

    await user.click(screen.getByRole('button', { name: /Verkauf/ }))
    await user.click(
      screen.getByRole('button', { name: 'Kassenbeleg drucken' }),
    )

    await waitFor(() => {
      expect(aktion.belegDrucken).toHaveBeenCalledTimes(1)
    })
    expect(aktion.belegDrucken).toHaveBeenCalledWith(verkaufId, undefined)
  })

  it('triggers stornobeleg print with the stornierungId of the cancellation', async () => {
    const user = userEvent.setup()
    const stornierterVerkauf: VorgangEintrag = {
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
    const aktion = renderHistorie([stornierterVerkauf])

    await user.click(screen.getByRole('button', { name: /Verkauf/ }))
    await user.click(
      screen.getByRole('button', { name: 'Stornobeleg drucken' }),
    )

    await waitFor(() => {
      expect(aktion.belegDrucken).toHaveBeenCalledTimes(1)
    })
    expect(aktion.belegDrucken).toHaveBeenCalledWith(verkaufId, stornierungId)
  })
})
