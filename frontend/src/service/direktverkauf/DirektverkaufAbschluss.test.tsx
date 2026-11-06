import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { type ComponentProps, useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useVorgang } from '@/hooks/use-vorgang'
import type { Produkt } from '@/lib/produktSchemas'

import type { ReceiptPosition } from '../table/Receipt'
import type { DirektverkaufTaetigen } from './Direktverkauf'
import { DirektverkaufAbschluss } from './DirektverkaufAbschluss'

// Verkauf und Pfand-Vorgang liegen auf der Seite; der Harness übernimmt ihre Rolle.
function DirektverkaufAbschlussHarness(
  props: Omit<
    ComponentProps<typeof DirektverkaufAbschluss>,
    'pfandVorgang' | 'schluesselFuer'
  >,
) {
  const verkauf = useVorgang<number>()
  return (
    <DirektverkaufAbschluss
      {...props}
      pfandVorgang={useVorgang<number>()}
      schluesselFuer={verkauf.schluesselFuer}
    />
  )
}

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

const pfandProdukt: Produkt = {
  id: 9,
  name: 'Pfand',
  kategorie: 'sonstiges',
  steuersatz: 'regel',
  pfand: true,
  status: 'active',
  varianten: [
    {
      id: 19,
      name: 'Glas',
      preisCents: 200,
      status: 'active',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    },
  ],
  createdAt: '2025-01-01T00:00:00Z',
  updatedAt: '2025-01-01T00:00:00Z',
}

// Die feste Spalte (variant="spalte") ist container-neutral testbar: kein
// Drawer, kein Dock.
function renderSpalte(
  direktverkaufTaetigen: (verkauf: DirektverkaufTaetigen) => Promise<void> = vi
    .fn()
    .mockResolvedValue(undefined),
) {
  const verkaufAbgeschlossen = vi.fn()
  render(
    <DirektverkaufAbschlussHarness
      pfandBackend={{ pfandZurueckgeben: vi.fn() }}
      pfandProdukte={[]}
      pfandAusgezahlt={vi.fn()}
      variant="spalte"
      backend={{ direktverkaufTaetigen }}
      receiptItems={receiptItems}
      positionen={positionen}
      totalCents={700}
      verkaufAbgeschlossen={verkaufAbgeschlossen}
      verkaufVerworfen={vi.fn()}
    />,
  )
  return { direktverkaufTaetigen, verkaufAbgeschlossen }
}

describe('DirektverkaufAbschluss (Spalte)', () => {
  it('zeigt im Leerzustand einen Hinweis und deaktiviert den Aktionsbutton', () => {
    render(
      <DirektverkaufAbschlussHarness
        pfandBackend={{ pfandZurueckgeben: vi.fn() }}
        pfandProdukte={[]}
        pfandAusgezahlt={vi.fn()}
        variant="spalte"
        backend={{ direktverkaufTaetigen: vi.fn() }}
        receiptItems={[]}
        positionen={[]}
        totalCents={0}
        verkaufAbgeschlossen={vi.fn()}
        verkaufVerworfen={vi.fn()}
      />,
    )

    expect(screen.getByText(/Produkte auswählen/)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    ).toBeDisabled()
  })

  it('sends the Pfand return with table id 0', async () => {
    const user = userEvent.setup()
    const direktverkaufTaetigen = vi.fn()
    const pfandZurueckgeben = vi.fn().mockResolvedValue(undefined)
    const pfandAusgezahlt = vi.fn()
    render(
      <DirektverkaufAbschlussHarness
        variant="spalte"
        backend={{ direktverkaufTaetigen }}
        pfandBackend={{ pfandZurueckgeben }}
        pfandProdukte={[pfandProdukt]}
        receiptItems={[]}
        positionen={[]}
        totalCents={0}
        verkaufAbgeschlossen={vi.fn()}
        verkaufVerworfen={vi.fn()}
        pfandAusgezahlt={pfandAusgezahlt}
      />,
    )

    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await user.click(screen.getByRole('button', { name: 'Pfand auszahlen' }))

    await waitFor(() => {
      expect(pfandZurueckgeben).toHaveBeenCalledTimes(1)
    })
    expect(pfandZurueckgeben).toHaveBeenCalledWith({
      // oxlint-disable-next-line typescript/no-unsafe-assignment
      rueckgabeId: expect.stringMatching(UUID),
      tischId: 0,
      positionen: [{ produktId: 9, varianteId: 19, menge: 1 }],
      kommentar: '',
    })
    expect(direktverkaufTaetigen).not.toHaveBeenCalled()
    expect(pfandAusgezahlt).toHaveBeenCalledTimes(1)
  })

  it('shows Auszahlen when the Pfand exceeds the sale', async () => {
    const user = userEvent.setup()
    const direktverkaufTaetigen = vi.fn().mockResolvedValue(undefined)
    render(
      <DirektverkaufAbschlussHarness
        variant="spalte"
        backend={{ direktverkaufTaetigen }}
        pfandBackend={{ pfandZurueckgeben: vi.fn() }}
        pfandProdukte={[pfandProdukt]}
        receiptItems={receiptItems}
        positionen={positionen}
        totalCents={700}
        verkaufAbgeschlossen={vi.fn()}
        verkaufVerworfen={vi.fn()}
        pfandAusgezahlt={vi.fn()}
      />,
    )

    for (let i = 0; i < 4; i++) {
      await user.click(
        screen.getByRole('button', { name: 'Variante hinzufügen' }),
      )
    }

    expect(screen.getByText('Auszahlen').nextElementSibling).toHaveTextContent(
      '1,00',
    )
    expect(screen.queryByLabelText('Erhalten')).not.toBeInTheDocument()

    await user.click(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    )
    await waitFor(() => {
      expect(direktverkaufTaetigen).toHaveBeenCalledTimes(1)
    })
    expect(direktverkaufTaetigen).toHaveBeenCalledWith({
      // oxlint-disable-next-line typescript/no-unsafe-assignment
      verkaufId: expect.stringMatching(UUID),
      positionen,
      kommentar: '',
      pfand: {
        // oxlint-disable-next-line typescript/no-unsafe-assignment
        rueckgabeId: expect.stringMatching(UUID),
        positionen: [{ produktId: 9, varianteId: 19, menge: 4 }],
      },
    })
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
        // oxlint-disable-next-line typescript/no-unsafe-assignment
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
          <DirektverkaufAbschlussHarness
            pfandBackend={{ pfandZurueckgeben: vi.fn() }}
            pfandProdukte={[]}
            pfandAusgezahlt={vi.fn()}
            variant="spalte"
            backend={{ direktverkaufTaetigen: vi.fn() }}
            receiptItems={leer ? [] : receiptItems}
            positionen={leer ? [] : positionen}
            totalCents={leer ? 0 : 700}
            verkaufAbgeschlossen={vi.fn()}
            verkaufVerworfen={vi.fn()}
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

  it('behält den verkaufId über einen Retry und erneuert ihn, wenn sich der Inhalt danach ändert', async () => {
    const user = userEvent.setup()
    const direktverkaufTaetigen = vi
      .fn<(verkauf: DirektverkaufTaetigen) => Promise<void>>()
      .mockRejectedValue(new Error('kaputt'))
    renderSpalte(direktverkaufTaetigen)

    const abschliessen = () =>
      screen.getByRole('button', { name: 'Verkauf abschließen' })

    await user.click(abschliessen())
    await waitFor(() => {
      expect(direktverkaufTaetigen).toHaveBeenCalledTimes(1)
    })
    await user.click(abschliessen())
    await waitFor(() => {
      expect(direktverkaufTaetigen).toHaveBeenCalledTimes(2)
    })
    const [[erster], [retry]] = direktverkaufTaetigen.mock.calls
    expect(retry.verkaufId).toBe(erster.verkaufId)

    // Ohne Pfand vergleicht das Backend den Inhalt einer bekannten verkaufId
    // nicht: der geänderte Verkauf bräuchte sonst nie gebucht.
    await user.type(screen.getByPlaceholderText(/Kommentar/), 'Ohne Senf')
    await user.click(abschliessen())
    await waitFor(() => {
      expect(direktverkaufTaetigen).toHaveBeenCalledTimes(3)
    })
    const geaendert = direktverkaufTaetigen.mock.calls[2][0]
    expect(geaendert.verkaufId).toMatch(UUID)
    expect(geaendert.verkaufId).not.toBe(erster.verkaufId)
    expect(geaendert.kommentar).toBe('Ohne Senf')
  })
})
