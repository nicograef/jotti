import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { BackendError } from '@/lib/Backend'
import type { Produkt } from '@/lib/produktSchemas'
import { signIn, signOut } from '@/test/auth'
import { FakeBackend } from '@/test/FakeBackend'
import { testProdukt } from '@/test/fixtures'
import { renderWithBackend, setViewportWidth } from '@/test/render'

import { DirektverkaufPage } from './DirektverkaufPage'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

// Handy-Pfad: der Fehlerzustand ist in beiden Layouts derselbe.
beforeEach(() => {
  setViewportWidth(375)
})

afterEach(() => {
  cleanup()
  signOut()
  vi.clearAllMocks()
  setViewportWidth(1024)
})

const pfandProdukt: Produkt = {
  ...testProdukt,
  id: 9,
  name: 'Pfand',
  kategorie: 'sonstiges',
  steuersatz: 'regel',
  pfand: true,
  varianten: [
    { ...testProdukt.varianten[0], id: 19, name: 'Glas', preisCents: 200 },
  ],
}

const verkaufId = '00000000-0000-4000-8000-000000000001'
const stornierungId = '00000000-0000-4000-8000-000000000003'

const cola = {
  positionId: '00000000-0000-4000-8000-000000000002',
  varianteId: 1,
  produktName: 'Cola',
  varianteName: '0,5l',
  kategorie: 'getraenk' as const,
  steuersatz: 'regel' as const,
  einzelpreisCents: 500,
  menge: 2,
}

const stornierterVerkauf = {
  verkaufId,
  userName: 'Anna',
  getaetigtAm: '2026-06-08T10:00:00Z',
  positionen: [{ ...cola, kommentar: '' }],
  gesamtbetragCents: 1000,
  kommentar: '',
  offenePositionen: [{ ...cola, menge: 1 }],
  gesamtStorniertCents: 500,
  stornierungen: [
    {
      stornierungId,
      storniertAm: '2026-06-08T11:00:00Z',
      gesamtStornierungCents: 500,
    },
  ],
}

function backend(
  pfandProdukte: Produkt[] = [],
  produkte: Produkt[] = [],
): FakeBackend {
  return new FakeBackend()
    .respond('service/get-aktive-produkte', {
      produkte,
      pfandProdukte,
    })
    .respond('service/get-direktverkauf-historie', { historie: [] })
    .respond('service/get-pfandrueckgabe-historie', { historie: [] })
}

describe('DirektverkaufPage', () => {
  it('zeigt bei Produkt-Fehler einen Fehlerzustand statt der Leer-Defaults', async () => {
    renderWithBackend(
      <DirektverkaufPage />,
      backend().fail('service/get-aktive-produkte'),
    )

    expect(
      await screen.findByText('Produkte konnten nicht geladen werden'),
    ).toBeInTheDocument()
    // Der Leer-Default (Verkaufs-Summe 0,00 €) darf bei einem Fehler nicht
    // erscheinen — das Sortiment wirkt sonst leer und der Verkauf abgerechnet.
    expect(screen.queryByText(/0,00 €/)).not.toBeInTheDocument()
  })

  it('lädt die Produkte über „Erneut versuchen" neu', async () => {
    const user = userEvent.setup()
    const getAktiveProdukte = vi
      .fn()
      .mockImplementationOnce(() => {
        throw new BackendError(400, 'test_fehler')
      })
      .mockReturnValue({ produkte: [testProdukt], pfandProdukte: [] })
    renderWithBackend(
      <DirektverkaufPage />,
      backend().respond('service/get-aktive-produkte', getAktiveProdukte),
    )

    await user.click(
      await screen.findByRole('button', { name: 'Erneut versuchen' }),
    )

    expect(await screen.findByText('Bratwurst')).toBeInTheDocument()
    expect(
      screen.queryByText('Produkte konnten nicht geladen werden'),
    ).not.toBeInTheDocument()
  })

  it('zeigt bei Historie-Fehler einen Fehlerzustand im Historie-Reiter', async () => {
    const user = userEvent.setup()
    renderWithBackend(
      <DirektverkaufPage />,
      backend().fail('service/get-direktverkauf-historie'),
    )

    await user.click(screen.getByRole('tab', { name: 'Historie' }))

    expect(
      await screen.findByText('Historie konnte nicht geladen werden'),
    ).toBeInTheDocument()
  })

  // Der Verkauf ist gebucht, die Antwort geht verloren. Nach einem Blick in die
  // Historie müssen Korb, Kommentar und verkaufId stehen, sonst bucht die
  // Wiederholung doppelt.
  it('behält Korb, Kommentar und verkaufId über einen Blick in die Historie', async () => {
    setViewportWidth(1024)
    const direktverkaufTaetigen = vi
      .fn()
      .mockImplementationOnce(() => {
        throw new Error('Netzwerkfehler')
      })
      .mockReturnValue({})
    const fake = backend([], [testProdukt]).respond(
      'service/direktverkauf-taetigen',
      direktverkaufTaetigen,
    )
    const user = userEvent.setup()
    renderWithBackend(<DirektverkaufPage />, fake)

    await user.click(
      await screen.findByRole('button', { name: 'Variante hinzufügen' }),
    )
    await user.type(screen.getByPlaceholderText(/Kommentar/), 'ohne Zwiebeln')
    await user.click(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    )
    await waitFor(() => {
      expect(direktverkaufTaetigen).toHaveBeenCalledTimes(1)
    })

    await user.click(screen.getByRole('tab', { name: 'Historie' }))
    await user.click(screen.getByRole('tab', { name: 'Verkaufen' }))
    await user.click(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    )
    await waitFor(() => {
      expect(direktverkaufTaetigen).toHaveBeenCalledTimes(2)
    })

    const [erster, wiederholung] = fake.bodies(
      'service/direktverkauf-taetigen',
    ) as [
      { verkaufId: string; positionen: unknown[]; kommentar: string },
      { verkaufId: string; positionen: unknown[]; kommentar: string },
    ]
    expect(wiederholung.kommentar).toBe('ohne Zwiebeln')
    expect(wiederholung.positionen).toEqual(erster.positionen)
    expect(wiederholung.positionen).toHaveLength(1)
    expect(wiederholung.verkaufId).toBe(erster.verkaufId)
  })

  // Die Rückgabe ist gebucht, die Antwort geht verloren. Nach einem Blick in die
  // Historie muss die Wiederholung dieselbe rueckgabeId tragen.
  it('behält Pfand-Auswahl und rueckgabeId über einen Blick in die Historie', async () => {
    setViewportWidth(1024)
    const pfandZurueckgeben = vi
      .fn()
      .mockImplementationOnce(() => {
        throw new Error('Netzwerkfehler')
      })
      .mockReturnValue({})
    const fake = backend([pfandProdukt]).respond(
      'service/pfand-zurueckgeben',
      pfandZurueckgeben,
    )
    const user = userEvent.setup()
    renderWithBackend(<DirektverkaufPage />, fake)

    await user.click(
      await screen.findByRole('button', { name: 'Variante hinzufügen' }),
    )
    await user.click(screen.getByRole('button', { name: 'Pfand auszahlen' }))
    await waitFor(() => {
      expect(pfandZurueckgeben).toHaveBeenCalledTimes(1)
    })

    await user.click(screen.getByRole('tab', { name: 'Historie' }))
    await user.click(screen.getByRole('tab', { name: 'Verkaufen' }))
    await user.click(screen.getByRole('button', { name: 'Pfand auszahlen' }))
    await waitFor(() => {
      expect(pfandZurueckgeben).toHaveBeenCalledTimes(2)
    })

    const [erster, wiederholung] = fake.bodies(
      'service/pfand-zurueckgeben',
    ) as [{ rueckgabeId: string }, { rueckgabeId: string }]
    expect(wiederholung.rueckgabeId).toBe(erster.rueckgabeId)
  })

  it.each([
    ['Kassenbeleg drucken', { verkaufId }],
    ['Stornobeleg drucken', { verkaufId, stornierungId }],
  ])('sends %s with the Direktverkauf Beleg body', async (button, body) => {
    const user = userEvent.setup()
    const fake = backend()
      .respond('service/get-direktverkauf-historie', {
        historie: [stornierterVerkauf],
      })
      .respond('service/beleg-drucken', { status: 'eingereiht' })
    renderWithBackend(<DirektverkaufPage />, fake)

    await user.click(screen.getByRole('tab', { name: 'Historie' }))
    await user.click(await screen.findByRole('button', { name: /Verkauf/ }))
    await user.click(screen.getByRole('button', { name: button }))

    await waitFor(() => {
      expect(fake.bodies('service/beleg-drucken')).toEqual([body])
    })
  })

  it('sendet den Storno und behält die stornierungId nur für denselben Inhalt', async () => {
    signIn({ role: 'serviceleitung' })
    const user = userEvent.setup()
    const fake = backend()
      .respond('service/get-direktverkauf-historie', {
        historie: [
          {
            ...stornierterVerkauf,
            offenePositionen: [cola],
            gesamtStorniertCents: 0,
            stornierungen: [],
          },
        ],
      })
      .fail('serviceleitung/direktverkauf-stornieren')
    renderWithBackend(<DirektverkaufPage />, fake)
    const senden = async (anzahl: number) => {
      await user.click(
        screen.getByRole('button', { name: 'Stornierung erteilen' }),
      )
      await waitFor(() => {
        expect(
          fake.bodies('serviceleitung/direktverkauf-stornieren'),
        ).toHaveLength(anzahl)
      })
    }

    await user.click(screen.getByRole('tab', { name: 'Historie' }))
    await user.click(await screen.findByRole('button', { name: /Verkauf/ }))
    await user.click(screen.getByRole('button', { name: /Stornieren…/ }))
    await user.click(
      screen.getByRole('button', { name: 'Cola 0,5l hinzufügen' }),
    )
    await user.type(
      screen.getByPlaceholderText('Kommentar (erforderlich)'),
      'Falsch gebucht',
    )
    await senden(1)
    await senden(2)
    await user.click(
      screen.getByRole('button', { name: 'Cola 0,5l hinzufügen' }),
    )
    await senden(3)

    const bodies = fake.bodies('serviceleitung/direktverkauf-stornieren') as {
      stornierungId: string
    }[]
    expect(bodies[0]).toEqual({
      stornierungId: bodies[0]?.stornierungId,
      verkaufId,
      positionen: [{ positionId: cola.positionId, menge: 1 }],
      kommentar: 'Falsch gebucht',
    })
    expect(bodies[0]?.stornierungId).toMatch(/^[0-9a-f-]{36}$/)
    expect(bodies[1]?.stornierungId).toBe(bodies[0]?.stornierungId)
    expect(bodies[2]).toMatchObject({
      positionen: [{ positionId: cola.positionId, menge: 2 }],
    })
    expect(bodies[2]?.stornierungId).not.toBe(bodies[0]?.stornierungId)
  })
})
