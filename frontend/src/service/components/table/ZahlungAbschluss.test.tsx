import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useVorgangId } from '@/hooks/use-vorgang-id'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { ZahlungKassieren } from '../../table/Zahlung'
import { ZahlungAbschluss } from './ZahlungAbschluss'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

afterEach(() => {
  cleanup()
})

const tisch: Tisch = { id: 1, name: 'Stammtisch', saldoCents: 900 }

const position: Position = {
  positionId: '00000000-0000-0000-0000-000000000001',
  varianteId: 1,
  produktName: 'Bratwurst',
  varianteName: 'Normal',
  kategorie: 'essen',
  steuersatz: 'regel',
  einzelpreisCents: 350,
  menge: 2,
  bestellerUserId: 1,
  bestellerName: 'Tester',
}

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/

// Der Idempotenz-Schlüssel kommt seit dem Heben nach TablePage als Prop; wo
// sein Lebenszyklus nicht Gegenstand des Tests ist, genügt ein fester Wert.
const VORGANG_ID = '11111111-1111-4111-8111-111111111111'

// Die feste Spalte (variant="spalte") ist container-neutral testbar: kein
// Drawer, kein Dock.
function renderSpalte(
  zahlungKassieren: () => Promise<void> = vi.fn().mockResolvedValue(undefined),
) {
  const zahlungKassiert = vi.fn()
  render(
    <ZahlungAbschluss
      variant="spalte"
      backend={{ zahlungKassieren }}
      tisch={tisch}
      positionenToPay={[{ ...position, menge: 2 }]}
      totalCents={700}
      restNachZahlungCents={200}
      vorgangId={VORGANG_ID}
      zahlungKassiert={zahlungKassiert}
      vorgangBereitsGebucht={vi.fn()}
    />,
  )
  return { zahlungKassieren, zahlungKassiert }
}

describe('ZahlungAbschluss (Spalte)', () => {
  it('zeigt im Leerzustand einen Hinweis und deaktiviert den Aktionsbutton', () => {
    render(
      <ZahlungAbschluss
        variant="spalte"
        backend={{ zahlungKassieren: vi.fn() }}
        tisch={tisch}
        positionenToPay={[]}
        totalCents={0}
        restNachZahlungCents={900}
        vorgangId={VORGANG_ID}
        zahlungKassiert={vi.fn()}
        vorgangBereitsGebucht={vi.fn()}
      />,
    )

    expect(screen.getByText(/Positionen auswählen/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Kassieren' })).toBeDisabled()
  })

  it('zeigt die Restbetrag-Zeile in der Spalte', () => {
    renderSpalte()
    expect(screen.getByText('Nach dieser Zahlung noch offen')).toBeVisible()
    expect(screen.getByText('2,00 €')).toBeInTheDocument()
  })

  it('rechnet Rückgeld ohne Zielbetrag und Trinkgeld mit Aufrunden-Chip', async () => {
    const user = userEvent.setup()
    renderSpalte()

    // Ohne Zielbetrag („genau"): Rückgeld = Erhalten − Gesamt.
    await user.type(screen.getByLabelText('Erhalten'), '10,00')
    expect(screen.getByText('Rückgeld').nextElementSibling).toHaveTextContent(
      '3,00',
    )

    // Aufrunden-Chip auf 8,00 € (Gesamt 7,00 €): Trinkgeld = Zielbetrag − Gesamt.
    await user.click(screen.getByRole('button', { name: /8,00/ }))
    expect(screen.getByText('Trinkgeld').nextElementSibling).toHaveTextContent(
      '1,00',
    )
    expect(screen.getByText('Rückgeld').nextElementSibling).toHaveTextContent(
      '2,00',
    )
  })

  it('kassiert mit genau einem Backend-Call und der erwarteten Nutzlast', async () => {
    const user = userEvent.setup()
    const { zahlungKassieren, zahlungKassiert } = renderSpalte()

    await user.click(screen.getByRole('button', { name: 'Kassieren' }))

    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(1)
    })
    expect(zahlungKassieren).toHaveBeenCalledWith({
      vorgangId: VORGANG_ID,
      tischId: 1,
      positionen: [{ positionId: position.positionId, menge: 2 }],
      kommentar: '',
    })
    expect(zahlungKassiert).toHaveBeenCalledTimes(1)
  })

  it('behält die vorgangId über Retry und Kommentaränderung und wechselt sie beim neuen Vorgang', async () => {
    const user = userEvent.setup()
    const zahlungKassieren = vi
      .fn<(z: ZahlungKassieren) => Promise<void>>()
      .mockRejectedValueOnce(new Error('kaputt'))
      .mockResolvedValue(undefined)

    // Der Harness steht für TablePage: Dort liegt der Schlüssel bei der
    // Auswahl, und von dort kommt er als Prop.
    function Harness() {
      const [leer, setLeer] = useState(false)
      const vorgangId = useVorgangId(leer)
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
          <ZahlungAbschluss
            variant="spalte"
            backend={{ zahlungKassieren }}
            tisch={tisch}
            positionenToPay={leer ? [] : [{ ...position, menge: 2 }]}
            totalCents={leer ? 0 : 700}
            restNachZahlungCents={leer ? 900 : 200}
            vorgangId={vorgangId}
            zahlungKassiert={vi.fn()}
            vorgangBereitsGebucht={vi.fn()}
          />
        </>
      )
    }
    render(<Harness />)

    const kassieren = () => screen.getByRole('button', { name: 'Kassieren' })

    await user.click(kassieren())
    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(1)
    })
    const ersterKey = zahlungKassieren.mock.calls[0][0].vorgangId
    expect(ersterKey).toMatch(UUID)

    // Wiederholversuch desselben Vorgangs: derselbe Schlüssel, der Server bucht
    // daher kein zweites Mal.
    await user.click(kassieren())
    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(2)
    })
    expect(zahlungKassieren.mock.calls[1][0].vorgangId).toBe(ersterKey)

    // Auch ein geänderter Kommentar ist kein neuer Vorgang: Der Schlüssel darf
    // sich nicht aus den Nutzdaten ableiten, sonst sähe der Server denselben
    // Schlüssel nie zweimal und könnte die Abweichung nicht beanstanden.
    await user.type(screen.getByPlaceholderText(/Kommentar/), 'Tisch 3')
    await user.click(kassieren())
    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(3)
    })
    expect(zahlungKassieren.mock.calls[2][0].vorgangId).toBe(ersterKey)
    expect(zahlungKassieren.mock.calls[2][0].kommentar).toBe('Tisch 3')

    // Neuer logischer Vorgang: Auswahl leeren (wie nach einem Erfolg) und neu füllen.
    await user.click(screen.getByRole('button', { name: 'toggle' }))
    await user.click(screen.getByRole('button', { name: 'toggle' }))
    await user.click(kassieren())
    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(4)
    })
    expect(zahlungKassieren.mock.calls[3][0].vorgangId).not.toBe(ersterKey)
  })

  it('setzt Erhalten, Zielbetrag und Kommentar beim neuen Vorgang zurück (kein Übertrag)', async () => {
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
          <ZahlungAbschluss
            variant="spalte"
            backend={{ zahlungKassieren: vi.fn() }}
            tisch={tisch}
            positionenToPay={leer ? [] : [{ ...position, menge: 2 }]}
            totalCents={leer ? 0 : 700}
            restNachZahlungCents={leer ? 900 : 200}
            vorgangId={VORGANG_ID}
            zahlungKassiert={vi.fn()}
            vorgangBereitsGebucht={vi.fn()}
          />
        </>
      )
    }
    render(<Harness />)

    await user.type(screen.getByLabelText('Erhalten'), '10,00')
    await user.click(screen.getByRole('button', { name: /8,00/ }))
    await user.type(screen.getByPlaceholderText(/Kommentar/), 'Für Tisch 3')
    expect(screen.getByLabelText('Erhalten')).toHaveValue('10,00')
    expect(screen.getByText('Trinkgeld')).toBeVisible()

    // Auswahl leeren (abgerechneter oder abgebrochener Vorgang) und neu
    // beginnen: Die dauerhafte Spalte bleibt dabei montiert, der Eingabe-State
    // überlebt den Auswahl-Reset also — ohne den Leerzustands-Reset trüge der
    // nächste Gast Erhalten, Zielbetrag und Kommentar des vorigen.
    await user.click(screen.getByRole('button', { name: 'toggle' }))
    await user.click(screen.getByRole('button', { name: 'toggle' }))

    expect(screen.getByLabelText('Erhalten')).toHaveValue('')
    expect(screen.getByPlaceholderText(/Kommentar/)).toHaveValue('')
    // Kein Zielbetrag mehr: ohne Trinkgeld auch keine Trinkgeld-Zeile.
    expect(screen.queryByText('Trinkgeld')).not.toBeInTheDocument()
  })

  // Das Befund-Szenario des Reviews: Kassieren schlägt scheinbar fehl (Antwort
  // im WLAN verloren), der Helfer nimmt eine weitere Position in die Auswahl
  // und kassiert erneut. Der zweite Versuch muss denselben Schlüssel tragen —
  // nur dann sieht der Server den Konflikt zwischen der bereits gebuchten und
  // der jetzt gesendeten Auswahl und meldet ihn, statt ein zweites Mal zu
  // buchen.
  it('behält die vorgangId, wenn die Auswahl nach einem Fehlversuch wächst', async () => {
    const user = userEvent.setup()
    const zahlungKassieren = vi
      .fn<(z: ZahlungKassieren) => Promise<void>>()
      .mockRejectedValueOnce(new Error('kaputt'))
      .mockResolvedValue(undefined)

    const positionB: Position = {
      ...position,
      positionId: '00000000-0000-0000-0000-000000000002',
      produktName: 'Pommes',
      einzelpreisCents: 400,
      menge: 1,
    }

    function Harness() {
      const [erweitert, setErweitert] = useState(false)
      const auswahl = erweitert
        ? [{ ...position, menge: 2 }, positionB]
        : [{ ...position, menge: 2 }]
      const vorgangId = useVorgangId(auswahl.length === 0)
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
          <ZahlungAbschluss
            variant="spalte"
            backend={{ zahlungKassieren }}
            tisch={tisch}
            positionenToPay={auswahl}
            totalCents={erweitert ? 1100 : 700}
            restNachZahlungCents={erweitert ? 0 : 400}
            vorgangId={vorgangId}
            zahlungKassiert={vi.fn()}
            vorgangBereitsGebucht={vi.fn()}
          />
        </>
      )
    }
    render(<Harness />)

    const kassieren = () => screen.getByRole('button', { name: 'Kassieren' })

    await user.click(kassieren())
    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(1)
    })
    const ersterKey = zahlungKassieren.mock.calls[0][0].vorgangId

    await user.click(screen.getByRole('button', { name: 'erweitern' }))
    await user.click(kassieren())
    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(2)
    })
    const zweiterAufruf = zahlungKassieren.mock.calls[1][0]
    expect(zweiterAufruf.vorgangId).toBe(ersterKey)
    expect(zweiterAufruf.positionen).toEqual([
      { positionId: position.positionId, menge: 2 },
      { positionId: positionB.positionId, menge: 1 },
    ])
  })

  it('löst bei Doppelklick keinen zweiten Aufruf aus', async () => {
    const user = userEvent.setup()
    let resolve: () => void = () => undefined
    const zahlungKassieren = vi.fn(
      () =>
        new Promise<void>((r) => {
          resolve = r
        }),
    )
    render(
      <ZahlungAbschluss
        variant="spalte"
        backend={{ zahlungKassieren }}
        tisch={tisch}
        positionenToPay={[{ ...position, menge: 2 }]}
        totalCents={700}
        restNachZahlungCents={200}
        vorgangId={VORGANG_ID}
        zahlungKassiert={vi.fn()}
        vorgangBereitsGebucht={vi.fn()}
      />,
    )

    const button = screen.getByRole('button', { name: 'Kassieren' })
    await user.click(button)
    // Zweiter Klick während des laufenden Submits: der Loading-Guard
    // deaktiviert den Button, es darf kein zweiter Aufruf entstehen.
    await user.click(button)
    resolve()

    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(1)
    })
  })
})
