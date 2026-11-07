import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { signIn, signOut } from '@/test/auth'
import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import type { PfandrueckgabeHistorieEintrag } from './Pfand'
import { PfandrueckgabenAbschnitt } from './PfandrueckgabenAbschnitt'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), info: vi.fn() },
}))

beforeEach(() => {
  signIn({ role: 'serviceleitung' })
})

afterEach(() => {
  cleanup()
  signOut()
})

const rueckgabeId = '00000000-0000-4000-8000-000000000011'
const positionId = '00000000-0000-4000-8000-000000000012'

const glas = {
  positionId,
  varianteId: 2,
  produktName: 'Pfand',
  varianteName: 'Glas',
  kategorie: 'sonstiges' as const,
  steuersatz: 'regel' as const,
  einzelpreisCents: 200,
  menge: 2,
}

const rueckgabe: PfandrueckgabeHistorieEintrag = {
  rueckgabeId,
  tischId: 4,
  userName: 'Anna',
  getaetigtAm: '2026-06-08T10:00:00Z',
  positionen: [{ ...glas, kommentar: '' }],
  gesamtbetragCents: 400,
  kommentar: '',
  offenePositionen: [glas],
  gesamtStorniertCents: 0,
  stornierungen: [],
}

const stornierungId = '00000000-0000-4000-8000-000000000013'

const stornierteRueckgabe: PfandrueckgabeHistorieEintrag = {
  ...rueckgabe,
  offenePositionen: [{ ...glas, menge: 1 }],
  gesamtStorniertCents: 200,
  stornierungen: [
    {
      stornierungId,
      storniertAm: '2026-06-08T11:00:00Z',
      gesamtStornierungCents: 200,
    },
  ],
}

describe('PfandrueckgabenAbschnitt', () => {
  it('cancels a return through pfandrueckgabeStornieren', async () => {
    const user = userEvent.setup()
    const fake = new FakeBackend()
      .respond('service/get-pfandrueckgabe-historie', { historie: [rueckgabe] })
      .respond('serviceleitung/pfandrueckgabe-stornieren', {})
    const onErfolg = vi.fn()
    renderWithBackend(
      <PfandrueckgabenAbschnitt tischId={4} onErfolg={onErfolg} />,
      fake,
    )

    await user.click(
      await screen.findByRole('button', { name: /Pfandrückgabe/ }),
    )
    await user.click(screen.getByRole('button', { name: /Stornieren…/ }))
    await user.click(
      screen.getByRole('button', { name: 'Pfand Glas hinzufügen' }),
    )
    await user.type(
      screen.getByPlaceholderText('Kommentar (erforderlich)'),
      'Falsch gezählt',
    )
    await user.click(
      screen.getByRole('button', { name: /Stornierung erteilen/ }),
    )

    await waitFor(() => {
      expect(onErfolg).toHaveBeenCalledWith('Stornierung gebucht.')
    })
    expect(fake.bodies('service/get-pfandrueckgabe-historie')).toEqual([
      { tischId: 4 },
    ])
    const bodies = fake.bodies('serviceleitung/pfandrueckgabe-stornieren')
    expect(bodies).toHaveLength(1)
    const { stornierungId, ...rest } = bodies[0] as Record<string, unknown>
    expect(stornierungId).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/,
    )
    expect(rest).toEqual({
      rueckgabeId,
      positionen: [{ positionId, menge: 1 }],
      kommentar: 'Falsch gezählt',
    })
  })

  // Eine verlorene Antwort darf nicht doppelt buchen: Die Wiederholung derselben
  // Auswahl trägt dieselbe stornierungId, eine andere Auswahl eine neue.
  it('keeps the stornierungId for a retry of the same selection only', async () => {
    const user = userEvent.setup()
    const stornieren = vi
      .fn()
      .mockImplementationOnce(() => {
        throw new Error('Netzwerkfehler')
      })
      .mockImplementationOnce(() => {
        throw new Error('Netzwerkfehler')
      })
      .mockReturnValue({})
    const fake = new FakeBackend()
      .respond('service/get-pfandrueckgabe-historie', { historie: [rueckgabe] })
      .respond('serviceleitung/pfandrueckgabe-stornieren', stornieren)
    renderWithBackend(
      <PfandrueckgabenAbschnitt tischId={4} onErfolg={vi.fn()} />,
      fake,
    )
    const erteilen = async (aufrufe: number) => {
      await user.click(
        screen.getByRole('button', { name: /Stornierung erteilen/ }),
      )
      await waitFor(() => {
        expect(stornieren).toHaveBeenCalledTimes(aufrufe)
      })
    }
    const glasHinzufuegen = () =>
      user.click(screen.getByRole('button', { name: 'Pfand Glas hinzufügen' }))

    await user.click(
      await screen.findByRole('button', { name: /Pfandrückgabe/ }),
    )
    await user.click(screen.getByRole('button', { name: /Stornieren…/ }))
    await glasHinzufuegen()
    await user.type(
      screen.getByPlaceholderText('Kommentar (erforderlich)'),
      'Falsch gezählt',
    )
    await erteilen(1)
    await erteilen(2)
    await glasHinzufuegen()
    await erteilen(3)

    const [erster, wiederholung, andereAuswahl] = fake.bodies(
      'serviceleitung/pfandrueckgabe-stornieren',
    ) as { stornierungId: string; positionen: unknown }[]
    expect(wiederholung.stornierungId).toBe(erster.stornierungId)
    expect(andereAuswahl.positionen).toEqual([{ positionId, menge: 2 }])
    expect(andereAuswahl.stornierungId).not.toBe(erster.stornierungId)
  })

  // Nach einem Fehlversuch schließt die Helferin den Drawer und öffnet ihn aus
  // der unveränderten Historie neu: dieselbe Auswahl trägt dieselbe stornierungId.
  it('keeps the stornierungId when the drawer is closed and reopened', async () => {
    const user = userEvent.setup()
    const stornieren = vi
      .fn()
      .mockImplementationOnce(() => {
        throw new Error('Netzwerkfehler')
      })
      .mockReturnValue({})
    const fake = new FakeBackend()
      .respond('service/get-pfandrueckgabe-historie', { historie: [rueckgabe] })
      .respond('serviceleitung/pfandrueckgabe-stornieren', stornieren)
    renderWithBackend(
      <PfandrueckgabenAbschnitt tischId={4} onErfolg={vi.fn()} />,
      fake,
    )
    const stornoErteilen = async (aufrufe: number) => {
      await user.click(
        await screen.findByRole('button', { name: /Pfandrückgabe/ }),
      )
      await user.click(screen.getByRole('button', { name: /Stornieren…/ }))
      await user.click(
        screen.getByRole('button', { name: 'Pfand Glas hinzufügen' }),
      )
      await user.type(
        screen.getByPlaceholderText('Kommentar (erforderlich)'),
        'Falsch gezählt',
      )
      await user.click(
        screen.getByRole('button', { name: /Stornierung erteilen/ }),
      )
      await waitFor(() => {
        expect(stornieren).toHaveBeenCalledTimes(aufrufe)
      })
    }

    await stornoErteilen(1)
    await user.click(screen.getByRole('button', { name: 'Abbrechen' }))
    await stornoErteilen(2)

    const [erster, wiederholung] = fake.bodies(
      'serviceleitung/pfandrueckgabe-stornieren',
    ) as { stornierungId: string }[]
    expect(wiederholung.stornierungId).toBe(erster.stornierungId)
  })

  it.each([
    ['Kassenbeleg drucken', { rueckgabeId }],
    ['Stornobeleg drucken', { rueckgabeId, stornierungId }],
  ])('sends %s with the Pfand Beleg body', async (button, body) => {
    const user = userEvent.setup()
    const fake = new FakeBackend()
      .respond('service/get-pfandrueckgabe-historie', {
        historie: [stornierteRueckgabe],
      })
      .respond('service/beleg-drucken', { status: 'eingereiht' })
    renderWithBackend(
      <PfandrueckgabenAbschnitt tischId={4} onErfolg={vi.fn()} />,
      fake,
    )

    await user.click(
      await screen.findByRole('button', { name: /Pfandrückgabe/ }),
    )
    await user.click(screen.getByRole('button', { name: button }))

    await waitFor(() => {
      expect(fake.bodies('service/beleg-drucken')).toEqual([body])
    })
  })
})
