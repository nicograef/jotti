import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useVorgangId } from '@/hooks/use-vorgang-id'

import type { BestellungAufnehmen } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import { BestellungAbschluss } from './BestellungAbschluss'
import type { ReceiptPosition } from './Receipt'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

afterEach(() => {
  cleanup()
})

const tisch: Tisch = { id: 1, name: 'Stammtisch', saldoCents: 0 }
const receiptItems: ReceiptPosition[] = [
  { name: 'Bratwurst Normal', einzelpreisCents: 350, menge: 2 },
]
const positionen = [{ produktId: 1, varianteId: 1, menge: 2 }]

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/

// Der Idempotenz-Schlüssel kommt seit dem Heben nach TablePage als Prop; wo
// sein Lebenszyklus nicht Gegenstand des Tests ist, genügt ein fester Wert.
const BESTELLUNG_ID = '11111111-1111-4111-8111-111111111111'

// Die feste Spalte (variant="spalte") ist container-neutral testbar: kein
// Drawer, kein Dock.
function renderSpalte(
  bestellungAufnehmen: () => Promise<void> = vi
    .fn()
    .mockResolvedValue(undefined),
) {
  const bestellungAufgenommen = vi.fn()
  render(
    <BestellungAbschluss
      variant="spalte"
      backend={{ bestellungAufnehmen }}
      tisch={tisch}
      receiptItems={receiptItems}
      positionen={positionen}
      totalCents={700}
      bestellungId={BESTELLUNG_ID}
      bestellungAufgenommen={bestellungAufgenommen}
      vorgangBereitsGebucht={vi.fn()}
    />,
  )
  return { bestellungAufnehmen, bestellungAufgenommen }
}

describe('BestellungAbschluss (Spalte)', () => {
  it('zeigt im Leerzustand einen Hinweis und deaktiviert den Aktionsbutton', () => {
    render(
      <BestellungAbschluss
        variant="spalte"
        backend={{ bestellungAufnehmen: vi.fn() }}
        tisch={tisch}
        receiptItems={[]}
        positionen={[]}
        totalCents={0}
        bestellungId={BESTELLUNG_ID}
        bestellungAufgenommen={vi.fn()}
        vorgangBereitsGebucht={vi.fn()}
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
        bestellungId: BESTELLUNG_ID,
        tischId: 1,
        positionen,
        kommentar: '',
      }),
    )
    expect(bestellungAufgenommen).toHaveBeenCalledTimes(1)
  })

  it('behält den bestellungId über einen Retry und wechselt ihn beim neuen Vorgang', async () => {
    const user = userEvent.setup()
    const bestellungAufnehmen = vi
      .fn<(b: BestellungAufnehmen) => Promise<void>>()
      .mockRejectedValueOnce(new Error('kaputt'))
      .mockResolvedValue(undefined)

    // Der Harness steht für TablePage: Dort liegt der Schlüssel beim Korb, und
    // von dort kommt er als Prop.
    function Harness() {
      const [leer, setLeer] = useState(false)
      const bestellungId = useVorgangId(leer)
      return (
        <>
          <button
            type="button"
            onClick={() => {
              setLeer((v) => !v)
            }}
          >
            toggle
          </button>
          <BestellungAbschluss
            variant="spalte"
            backend={{ bestellungAufnehmen }}
            tisch={tisch}
            receiptItems={leer ? [] : receiptItems}
            positionen={leer ? [] : positionen}
            totalCents={leer ? 0 : 700}
            bestellungId={bestellungId}
            bestellungAufgenommen={vi.fn()}
            vorgangBereitsGebucht={vi.fn()}
          />
        </>
      )
    }
    render(<Harness />)

    const aufnehmen = () =>
      screen.getByRole('button', { name: 'Bestellung aufnehmen' })

    await user.click(aufnehmen())
    await waitFor(() => {
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(1)
    })
    const ersterKey = bestellungAufnehmen.mock.calls[0][0].bestellungId
    expect(ersterKey).toMatch(UUID)

    await user.click(aufnehmen())
    await waitFor(() => {
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(2)
    })
    expect(bestellungAufnehmen.mock.calls[1][0].bestellungId).toBe(ersterKey)

    // Neuer logischer Vorgang: Auswahl leeren und neu füllen.
    await user.click(screen.getByRole('button', { name: 'toggle' }))
    await user.click(screen.getByRole('button', { name: 'toggle' }))
    await user.click(aufnehmen())
    await waitFor(() => {
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(3)
    })
    expect(bestellungAufnehmen.mock.calls[2][0].bestellungId).not.toBe(
      ersterKey,
    )
  })

  // Scheitert der erste Versuch scheinbar (Antwort im WLAN verloren) und wächst
  // die Auswahl danach, muss der zweite Versuch denselben Schlüssel tragen: Nur
  // dann erkennt der Server den Konflikt mit der bereits gebuchten Bestellung
  // und meldet ihn, statt ein zweites Mal zu buchen.
  it('behält den bestellungId, wenn die Auswahl nach einem Fehlversuch wächst', async () => {
    const user = userEvent.setup()
    const bestellungAufnehmen = vi
      .fn<(b: BestellungAufnehmen) => Promise<void>>()
      .mockRejectedValueOnce(new Error('kaputt'))
      .mockResolvedValue(undefined)

    const erweitertePositionen = [
      ...positionen,
      { produktId: 2, varianteId: 3, menge: 1 },
    ]

    function Harness() {
      const [erweitert, setErweitert] = useState(false)
      const bestellungId = useVorgangId(false)
      return (
        <>
          <button
            type="button"
            onClick={() => {
              setErweitert(true)
            }}
          >
            erweitern
          </button>
          <BestellungAbschluss
            variant="spalte"
            backend={{ bestellungAufnehmen }}
            tisch={tisch}
            receiptItems={receiptItems}
            positionen={erweitert ? erweitertePositionen : positionen}
            totalCents={erweitert ? 950 : 700}
            bestellungId={bestellungId}
            bestellungAufgenommen={vi.fn()}
            vorgangBereitsGebucht={vi.fn()}
          />
        </>
      )
    }
    render(<Harness />)

    const aufnehmen = () =>
      screen.getByRole('button', { name: 'Bestellung aufnehmen' })

    await user.click(aufnehmen())
    await waitFor(() => {
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(1)
    })
    const ersterKey = bestellungAufnehmen.mock.calls[0][0].bestellungId

    // Geänderte Nutzdaten nach dem Fehlversuch: derselbe Vorgang, derselbe
    // Schlüssel — die Abweichung beanstandet der Server.
    await user.click(screen.getByRole('button', { name: 'erweitern' }))
    await user.click(aufnehmen())
    await waitFor(() => {
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(2)
    })
    const zweiterAufruf = bestellungAufnehmen.mock.calls[1][0]
    expect(zweiterAufruf.bestellungId).toBe(ersterKey)
    expect(zweiterAufruf.positionen).toEqual(erweitertePositionen)
  })

  it('setzt den Kommentar beim neuen Vorgang zurück', async () => {
    const user = userEvent.setup()

    function Harness() {
      const [leer, setLeer] = useState(false)
      return (
        <>
          <button
            type="button"
            onClick={() => {
              setLeer((v) => !v)
            }}
          >
            toggle
          </button>
          <BestellungAbschluss
            variant="spalte"
            backend={{ bestellungAufnehmen: vi.fn() }}
            tisch={tisch}
            receiptItems={leer ? [] : receiptItems}
            positionen={leer ? [] : positionen}
            totalCents={leer ? 0 : 700}
            bestellungId={BESTELLUNG_ID}
            bestellungAufgenommen={vi.fn()}
            vorgangBereitsGebucht={vi.fn()}
          />
        </>
      )
    }
    render(<Harness />)

    await user.type(screen.getByPlaceholderText(/Kommentar/), 'Ohne Zwiebeln')
    await user.click(screen.getByRole('button', { name: 'toggle' }))
    await user.click(screen.getByRole('button', { name: 'toggle' }))

    expect(screen.getByPlaceholderText(/Kommentar/)).toHaveValue('')
  })
})
