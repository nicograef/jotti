import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ComponentProps } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useVorgang } from '@/hooks/use-vorgang'

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
  bestellungAufnehmen: () => Promise<void> = vi
    .fn()
    .mockResolvedValue(undefined),
) {
  const bestellungAufgenommen = vi.fn()
  render(
    <BestellungAbschlussHarness
      variant="spalte"
      backend={{ bestellungAufnehmen }}
      tisch={tisch}
      receiptItems={receiptItems}
      positionen={positionen}
      totalCents={700}
      bestellungAufgenommen={bestellungAufgenommen}
    />,
  )
  return { bestellungAufnehmen, bestellungAufgenommen }
}

describe('BestellungAbschluss (Spalte)', () => {
  it('zeigt im Leerzustand einen Hinweis und deaktiviert den Aktionsbutton', () => {
    render(
      <BestellungAbschlussHarness
        variant="spalte"
        backend={{ bestellungAufnehmen: vi.fn() }}
        tisch={tisch}
        receiptItems={[]}
        positionen={[]}
        totalCents={0}
        bestellungAufgenommen={vi.fn()}
      />,
    )

    expect(screen.getByText(/Produkte auswählen/)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Bestellung aufnehmen' }),
    ).toBeDisabled()
  })

  it('zeigt Beleg, Tischname und Gesamtsumme und nimmt mit genau einem Backend-Call auf', async () => {
    const user = userEvent.setup()
    const { bestellungAufnehmen, bestellungAufgenommen } = renderSpalte()

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
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(1)
    })
    expect(bestellungAufnehmen).toHaveBeenCalledWith(
      expect.objectContaining({
        // oxlint-disable-next-line typescript/no-unsafe-assignment
        bestellungId: expect.stringMatching(UUID),
        tischId: 1,
        positionen,
        kommentar: '',
      }),
    )
    expect(bestellungAufgenommen).toHaveBeenCalledTimes(1)
  })

  it('erneuert den bestellungId, wenn sich der Kommentar nach einem Fehlversuch ändert', async () => {
    const user = userEvent.setup()
    const bestellungAufnehmen = vi
      .fn<(b: BestellungAufnehmen) => Promise<void>>()
      .mockRejectedValueOnce(new Error('kaputt'))
      .mockResolvedValue(undefined)
    render(
      <BestellungAbschlussHarness
        variant="spalte"
        backend={{ bestellungAufnehmen }}
        tisch={tisch}
        receiptItems={receiptItems}
        positionen={positionen}
        totalCents={700}
        bestellungAufgenommen={vi.fn()}
      />,
    )

    const aufnehmen = () =>
      screen.getByRole('button', { name: 'Bestellung aufnehmen' })

    await user.click(aufnehmen())
    await waitFor(() => {
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(1)
    })
    const ersterKey = bestellungAufnehmen.mock.calls[0][0].bestellungId

    // Geänderter Inhalt ist ein anderer Vorgang: Mit dem alten Schlüssel würde
    // das Backend ihn als Duplikat verwerfen.
    await user.type(screen.getByPlaceholderText(/Kommentar/), 'Ohne Zwiebeln')
    await user.click(aufnehmen())
    await waitFor(() => {
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(2)
    })
    expect(bestellungAufnehmen.mock.calls[1][0].bestellungId).not.toBe(
      ersterKey,
    )
    expect(bestellungAufnehmen.mock.calls[1][0].kommentar).toBe('Ohne Zwiebeln')
  })
})
