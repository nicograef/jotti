import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ComponentProps } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useVorgang } from '@/hooks/use-vorgang'
import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import type { BestellungAufnehmen } from './Bestellung'
import { BestellungAbschluss } from './BestellungAbschluss'
import type { ReceiptPosition } from './Receipt'
import type { Tisch } from './Tisch'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

afterEach(() => {
  cleanup()
})

// Der Vorgang liegt auf der Seite; der Harness übernimmt ihre Rolle.
function BestellungAbschlussHarness(
  props: Omit<ComponentProps<typeof BestellungAbschluss>, 'vorgang'>,
) {
  return <BestellungAbschluss {...props} vorgang={useVorgang<number>()} />
}

const tisch: Tisch = { id: 1, name: 'Stammtisch', saldoCents: 0 }
const receiptItems: ReceiptPosition[] = [
  { name: 'Bratwurst Normal', einzelpreisCents: 350, menge: 2 },
]
const positionen = [{ produktId: 1, varianteId: 1, menge: 2 }]

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/

// Die feste Spalte (variant="spalte") ist container-neutral testbar: kein
// Drawer, kein Dock.
function renderSpalte(
  fake = new FakeBackend().respond('service/bestellung-aufnehmen', {}),
) {
  const bestellungAufgenommen = vi.fn()
  renderWithBackend(
    <BestellungAbschlussHarness
      variant="spalte"
      tisch={tisch}
      receiptItems={receiptItems}
      positionen={positionen}
      totalCents={700}
      bestellungAufgenommen={bestellungAufgenommen}
    />,
    fake,
  )
  return { fake, bestellungAufgenommen }
}

describe('BestellungAbschluss (Spalte)', () => {
  it('zeigt im Leerzustand einen Hinweis und deaktiviert den Aktionsbutton', () => {
    renderWithBackend(
      <BestellungAbschlussHarness
        variant="spalte"
        tisch={tisch}
        receiptItems={[]}
        positionen={[]}
        totalCents={0}
        bestellungAufgenommen={vi.fn()}
      />,
      new FakeBackend(),
    )

    expect(screen.getByText(/Produkte auswählen/)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Bestellung aufnehmen' }),
    ).toBeDisabled()
  })

  it('zeigt Beleg, Tischname und Gesamtsumme und nimmt mit genau einem Backend-Call auf', async () => {
    const user = userEvent.setup()
    const { fake, bestellungAufgenommen } = renderSpalte()

    expect(screen.getByRole('heading', { name: 'Stammtisch' })).toBeVisible()
    expect(screen.getByText(/Bratwurst/)).toBeInTheDocument()
    // Gesamt-Zeile im Footer zeigt die Summe (der Beleg zeigt dieselbe Zahl
    // zusätzlich in der Positionszeile).
    expect(screen.getByText('Gesamt').nextElementSibling).toHaveTextContent(
      '7,00',
    )

    await user.click(
      screen.getByRole('button', { name: 'Bestellung aufnehmen' }),
    )

    await waitFor(() => {
      expect(bestellungAufgenommen).toHaveBeenCalledTimes(1)
    })
    expect(fake.bodies('service/bestellung-aufnehmen')).toEqual([
      {
        // oxlint-disable-next-line typescript/no-unsafe-assignment
        bestellungId: expect.stringMatching(UUID),
        tischId: 1,
        positionen,
        kommentar: '',
      },
    ])
  })

  it('erneuert den bestellungId, wenn sich der Kommentar nach einem Fehlversuch ändert', async () => {
    const user = userEvent.setup()
    const { fake } = renderSpalte(
      new FakeBackend().respond(
        'service/bestellung-aufnehmen',
        vi.fn().mockRejectedValueOnce(new Error('kaputt')).mockReturnValue({}),
      ),
    )
    const bestellungen = () =>
      fake.bodies('service/bestellung-aufnehmen') as BestellungAufnehmen[]

    const aufnehmen = () =>
      screen.getByRole('button', { name: 'Bestellung aufnehmen' })

    await user.click(aufnehmen())
    await waitFor(() => {
      expect(bestellungen()).toHaveLength(1)
    })
    const ersterKey = bestellungen()[0].bestellungId

    // Geänderter Inhalt ist ein anderer Vorgang: Mit dem alten Schlüssel würde
    // das Backend ihn als Duplikat verwerfen.
    await user.type(screen.getByPlaceholderText(/Kommentar/), 'Ohne Zwiebeln')
    await user.click(aufnehmen())
    await waitFor(() => {
      expect(bestellungen()).toHaveLength(2)
    })
    expect(bestellungen()[1].bestellungId).not.toBe(ersterKey)
    expect(bestellungen()[1].kommentar).toBe('Ohne Zwiebeln')
  })
})
