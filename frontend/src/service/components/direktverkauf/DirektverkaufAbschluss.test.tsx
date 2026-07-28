import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { DirektverkaufTaetigen } from '../../direktverkauf/Direktverkauf'
import type { ReceiptPosition } from '../table/Receipt'
import { DirektverkaufAbschluss } from './DirektverkaufAbschluss'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

afterEach(() => {
  cleanup()
})

const receiptItems: ReceiptPosition[] = [
  { name: 'Bratwurst Normal', einzelpreisCents: 350, menge: 2 },
]
const positionen = [{ produktId: 1, varianteId: 1, menge: 2 }]

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/

// Die feste Spalte (variant="spalte") ist container-neutral testbar: kein
// Drawer, kein Dock. Sie ist im Leerzustand deaktiviert, rechnet Rückgeld und
// ruft das Backend genau einmal mit stabilem verkaufId je Vorgang auf.
function renderSpalte(
  direktverkaufTaetigen: () => Promise<void> = vi
    .fn()
    .mockResolvedValue(undefined),
) {
  const verkaufAbgeschlossen = vi.fn()
  render(
    <DirektverkaufAbschluss
      variant="spalte"
      backend={{ direktverkaufTaetigen }}
      receiptItems={receiptItems}
      positionen={positionen}
      totalCents={700}
      verkaufAbgeschlossen={verkaufAbgeschlossen}
      vorgangBereitsGebucht={vi.fn()}
    />,
  )
  return { direktverkaufTaetigen, verkaufAbgeschlossen }
}

describe('DirektverkaufAbschluss (Spalte)', () => {
  it('zeigt im Leerzustand einen Hinweis und deaktiviert den Aktionsbutton', () => {
    render(
      <DirektverkaufAbschluss
        variant="spalte"
        backend={{ direktverkaufTaetigen: vi.fn() }}
        receiptItems={[]}
        positionen={[]}
        totalCents={0}
        verkaufAbgeschlossen={vi.fn()}
        vorgangBereitsGebucht={vi.fn()}
      />,
    )

    expect(screen.getByText(/Produkte auswählen/)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    ).toBeDisabled()
  })

  it('rechnet das Rückgeld aus dem erhaltenen Betrag', async () => {
    const user = userEvent.setup()
    renderSpalte()

    await user.type(screen.getByLabelText('Erhalten'), '10,00')

    const rueckgeld = screen.getByText('Rückgeld').nextElementSibling
    expect(rueckgeld).toHaveTextContent('3,00')
  })

  it('zeigt Trinkgeld und Rückgeld, wenn ein Aufrunden-Chip gewählt ist', async () => {
    const user = userEvent.setup()
    renderSpalte()

    // Gesamt 7,00 €: Chip auf 8,00 € rundet auf, Erhalten 10,00 €.
    await user.click(screen.getByRole('button', { name: /8,00/ }))
    await user.type(screen.getByLabelText('Erhalten'), '10,00')

    expect(screen.getByText('Trinkgeld').nextElementSibling).toHaveTextContent(
      '1,00',
    )
    expect(screen.getByText('Rückgeld').nextElementSibling).toHaveTextContent(
      '2,00',
    )
  })

  it('schließt den Verkauf mit genau einem Backend-Call und erwarteter Nutzlast ab', async () => {
    const user = userEvent.setup()
    const { direktverkaufTaetigen, verkaufAbgeschlossen } = renderSpalte()

    await user.click(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    )

    await waitFor(() => {
      expect(direktverkaufTaetigen).toHaveBeenCalledTimes(1)
    })
    expect(direktverkaufTaetigen).toHaveBeenCalledWith(
      expect.objectContaining({
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
        verkaufId: expect.stringMatching(UUID),
        positionen,
        kommentar: '',
      }),
    )
    expect(verkaufAbgeschlossen).toHaveBeenCalledTimes(1)
  })

  it('setzt Erhalten und Kommentar beim neuen Vorgang zurück (kein Übertrag)', async () => {
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
          <DirektverkaufAbschluss
            variant="spalte"
            backend={{ direktverkaufTaetigen: vi.fn() }}
            receiptItems={leer ? [] : receiptItems}
            positionen={leer ? [] : positionen}
            totalCents={leer ? 0 : 700}
            verkaufAbgeschlossen={vi.fn()}
            vorgangBereitsGebucht={vi.fn()}
          />
        </>
      )
    }
    render(<Harness />)

    await user.type(screen.getByLabelText('Erhalten'), '10,00')
    await user.type(screen.getByPlaceholderText(/Kommentar/), 'Für Tisch 3')
    expect(screen.getByLabelText('Erhalten')).toHaveValue('10,00')

    // Auswahl leeren (abgebrochener Vorgang) und neu beginnen: der neue Vorgang
    // startet mit leeren Eingaben, damit nichts aus dem alten übertragen wird.
    await user.click(screen.getByRole('button', { name: 'toggle' }))
    await user.click(screen.getByRole('button', { name: 'toggle' }))

    expect(screen.getByLabelText('Erhalten')).toHaveValue('')
    expect(screen.getByPlaceholderText(/Kommentar/)).toHaveValue('')
  })

  it('behält den verkaufId über einen Retry und wechselt ihn beim neuen Vorgang', async () => {
    const user = userEvent.setup()
    const direktverkaufTaetigen = vi
      .fn<(verkauf: DirektverkaufTaetigen) => Promise<void>>()
      .mockRejectedValueOnce(new Error('kaputt'))
      .mockResolvedValue(undefined)

    // Testrahmen, der die Auswahl von „gefüllt" nach „leer" (erfolgreicher
    // Abschluss) und zurück schalten kann, um den Vorgangswechsel zu prüfen.
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
          <DirektverkaufAbschluss
            variant="spalte"
            backend={{ direktverkaufTaetigen }}
            receiptItems={leer ? [] : receiptItems}
            positionen={leer ? [] : positionen}
            totalCents={leer ? 0 : 700}
            verkaufAbgeschlossen={vi.fn()}
            vorgangBereitsGebucht={vi.fn()}
          />
        </>
      )
    }
    render(<Harness />)

    const abschliessen = () =>
      screen.getByRole('button', { name: 'Verkauf abschließen' })

    // Erster (fehlschlagender) Versuch.
    await user.click(abschliessen())
    await waitFor(() => {
      expect(direktverkaufTaetigen).toHaveBeenCalledTimes(1)
    })
    const ersterKey = direktverkaufTaetigen.mock.calls[0][0].verkaufId

    // Retry desselben Vorgangs: gleicher Schlüssel.
    await user.click(abschliessen())
    await waitFor(() => {
      expect(direktverkaufTaetigen).toHaveBeenCalledTimes(2)
    })
    expect(direktverkaufTaetigen.mock.calls[1][0].verkaufId).toBe(ersterKey)

    // Neuer logischer Vorgang: Auswahl leeren (Abschluss) und neu füllen.
    await user.click(screen.getByRole('button', { name: 'toggle' }))
    await user.click(screen.getByRole('button', { name: 'toggle' }))
    await user.click(abschliessen())
    await waitFor(() => {
      expect(direktverkaufTaetigen).toHaveBeenCalledTimes(3)
    })
    expect(direktverkaufTaetigen.mock.calls[2][0].verkaufId).not.toBe(ersterKey)
  })

  // Scheitert der erste Versuch scheinbar (Antwort im WLAN verloren) und wächst
  // die Auswahl danach, muss der zweite Versuch denselben Schlüssel tragen: Nur
  // dann erkennt der Server den Konflikt mit dem bereits gebuchten Verkauf und
  // meldet ihn, statt ein zweites Mal zu buchen.
  it('behält den verkaufId, wenn die Auswahl nach einem Fehlversuch wächst', async () => {
    const user = userEvent.setup()
    const direktverkaufTaetigen = vi
      .fn<(verkauf: DirektverkaufTaetigen) => Promise<void>>()
      .mockRejectedValueOnce(new Error('kaputt'))
      .mockResolvedValue(undefined)

    const erweitertePositionen = [
      ...positionen,
      { produktId: 2, varianteId: 3, menge: 1 },
    ]

    function Harness() {
      const [erweitert, setErweitert] = useState(false)
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
          <DirektverkaufAbschluss
            variant="spalte"
            backend={{ direktverkaufTaetigen }}
            receiptItems={receiptItems}
            positionen={erweitert ? erweitertePositionen : positionen}
            totalCents={erweitert ? 950 : 700}
            verkaufAbgeschlossen={vi.fn()}
            vorgangBereitsGebucht={vi.fn()}
          />
        </>
      )
    }
    render(<Harness />)

    const abschliessen = () =>
      screen.getByRole('button', { name: 'Verkauf abschließen' })

    await user.click(abschliessen())
    await waitFor(() => {
      expect(direktverkaufTaetigen).toHaveBeenCalledTimes(1)
    })
    const ersterKey = direktverkaufTaetigen.mock.calls[0][0].verkaufId

    // Geänderte Nutzdaten nach dem Fehlversuch: derselbe Vorgang, derselbe
    // Schlüssel — die Abweichung beanstandet der Server.
    await user.click(screen.getByRole('button', { name: 'erweitern' }))
    await user.click(abschliessen())
    await waitFor(() => {
      expect(direktverkaufTaetigen).toHaveBeenCalledTimes(2)
    })
    const zweiterAufruf = direktverkaufTaetigen.mock.calls[1][0]
    expect(zweiterAufruf.verkaufId).toBe(ersterKey)
    expect(zweiterAufruf.positionen).toEqual(erweitertePositionen)
  })
})
