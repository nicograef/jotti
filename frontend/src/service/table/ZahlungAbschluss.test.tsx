import {
  cleanup,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ComponentProps } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useVorgang } from '@/hooks/use-vorgang'
import { BackendError } from '@/lib/Backend'
import type { Produkt } from '@/lib/produktSchemas'

import type { Position } from './Bestellung'
import type { Tisch } from './Tisch'
import type { ZahlungKassieren } from './Zahlung'
import { ZahlungAbschluss } from './ZahlungAbschluss'

// Der Pfand-Vorgang liegt auf der Seite; der Harness übernimmt ihre Rolle.
function ZahlungAbschlussHarness(
  props: Omit<ComponentProps<typeof ZahlungAbschluss>, 'pfandVorgang'>,
) {
  return <ZahlungAbschluss {...props} pfandVorgang={useVorgang<number>()} />
}

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

afterEach(() => {
  cleanup()
})

const tisch: Tisch = { id: 1, name: 'Stammtisch', saldoCents: 900 }

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

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/

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

// Die feste Spalte (variant="spalte") ist container-neutral testbar: kein
// Drawer, kein Dock.
function renderSpalte(
  zahlungKassieren: () => Promise<void> = vi.fn().mockResolvedValue(undefined),
) {
  const zahlungKassiert = vi.fn()
  render(
    <ZahlungAbschlussHarness
      pfandBackend={{ pfandZurueckgeben: vi.fn() }}
      pfandProdukte={[]}
      pfandAusgezahlt={vi.fn()}
      variant="spalte"
      backend={{ zahlungKassieren }}
      tisch={tisch}
      positionenToPay={[{ ...position, menge: 2 }]}
      totalCents={700}
      restNachZahlungCents={200}
      zahlungKassiert={zahlungKassiert}
    />,
  )
  return { zahlungKassieren, zahlungKassiert }
}

describe('ZahlungAbschluss (Spalte)', () => {
  it('zeigt im Leerzustand einen Hinweis und deaktiviert den Aktionsbutton', () => {
    render(
      <ZahlungAbschlussHarness
        pfandBackend={{ pfandZurueckgeben: vi.fn() }}
        pfandProdukte={[]}
        pfandAusgezahlt={vi.fn()}
        variant="spalte"
        backend={{ zahlungKassieren: vi.fn() }}
        tisch={tisch}
        positionenToPay={[]}
        totalCents={0}
        restNachZahlungCents={900}
        zahlungKassiert={vi.fn()}
      />,
    )

    expect(screen.getByText(/Positionen auswählen/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Kassieren' })).toBeDisabled()
  })

  it('sends only the Pfand return with the table id', async () => {
    const user = userEvent.setup()
    const zahlungKassieren = vi.fn()
    const pfandZurueckgeben = vi.fn().mockResolvedValue(undefined)
    const pfandAusgezahlt = vi.fn()
    render(
      <ZahlungAbschlussHarness
        variant="spalte"
        backend={{ zahlungKassieren }}
        pfandBackend={{ pfandZurueckgeben }}
        pfandProdukte={[pfandProdukt]}
        tisch={tisch}
        positionenToPay={[]}
        totalCents={0}
        restNachZahlungCents={900}
        zahlungKassiert={vi.fn()}
        pfandAusgezahlt={pfandAusgezahlt}
      />,
    )

    expect(screen.getByText('Pfand zurück')).toBeInTheDocument()
    expect(screen.queryByText(/Positionen auswählen/)).not.toBeInTheDocument()
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await user.type(screen.getByPlaceholderText(/Kommentar/), 'Zwei Gläser')
    await user.click(screen.getByRole('button', { name: 'Pfand auszahlen' }))

    await waitFor(() => {
      expect(pfandZurueckgeben).toHaveBeenCalledTimes(1)
    })
    expect(pfandZurueckgeben).toHaveBeenCalledWith({
      // oxlint-disable-next-line typescript/no-unsafe-assignment
      rueckgabeId: expect.stringMatching(UUID),
      tischId: 1,
      positionen: [{ produktId: 9, varianteId: 19, menge: 2 }],
      kommentar: 'Zwei Gläser',
    })
    expect(zahlungKassieren).not.toHaveBeenCalled()
    expect(pfandAusgezahlt).toHaveBeenCalledTimes(1)
  })

  it('nets the Pfand against the payment', async () => {
    const user = userEvent.setup()
    const zahlungKassieren = vi.fn().mockResolvedValue(undefined)
    render(
      <ZahlungAbschlussHarness
        variant="spalte"
        backend={{ zahlungKassieren }}
        pfandBackend={{ pfandZurueckgeben: vi.fn() }}
        pfandProdukte={[pfandProdukt]}
        tisch={tisch}
        positionenToPay={[{ ...position, menge: 2 }]}
        totalCents={700}
        restNachZahlungCents={200}
        zahlungKassiert={vi.fn()}
        pfandAusgezahlt={vi.fn()}
      />,
    )

    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    const verrechnung = within(screen.getByLabelText('Verrechnung'))
    expect(
      verrechnung.getByText('Kaufsumme').nextElementSibling,
    ).toHaveTextContent('7,00')
    expect(
      verrechnung.getByText('Pfand zurück').nextElementSibling,
    ).toHaveTextContent('-2,00')
    expect(
      verrechnung.getByText('Zu zahlen').nextElementSibling,
    ).toHaveTextContent('5,00')

    await user.type(screen.getByLabelText('Erhalten'), '10,00')
    expect(screen.getByText('Rückgeld').nextElementSibling).toHaveTextContent(
      '5,00',
    )

    await user.click(screen.getByRole('button', { name: 'Kassieren' }))
    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(1)
    })
    expect(zahlungKassieren).toHaveBeenCalledWith({
      tischId: 1,
      positionen: [{ positionId: position.positionId, menge: 2 }],
      kommentar: '',
      pfand: {
        // oxlint-disable-next-line typescript/no-unsafe-assignment
        rueckgabeId: expect.stringMatching(UUID),
        positionen: [{ produktId: 9, varianteId: 19, menge: 1 }],
      },
    })
  })

  // Zahlung und Rückgabe sind gebucht, die Antwort geht verloren; der Refetch
  // nimmt die bezahlten Positionen aus der Auswahl. Erst die alte rueckgabeId
  // lässt den Server die Wiederholung als verrechnet abweisen.
  it('keeps the rueckgabeId when a refetch prunes the paid positions', async () => {
    const user = userEvent.setup()
    const zahlungKassieren = vi
      .fn<(zahlung: ZahlungKassieren) => Promise<void>>()
      .mockRejectedValue(new Error('Netzwerkfehler'))
    const pfandZurueckgeben = vi.fn().mockResolvedValue(undefined)
    const props = {
      variant: 'spalte' as const,
      backend: { zahlungKassieren },
      pfandBackend: { pfandZurueckgeben },
      pfandProdukte: [pfandProdukt],
      tisch,
      restNachZahlungCents: 200,
      zahlungKassiert: vi.fn(),
      pfandAusgezahlt: vi.fn(),
    }
    const { rerender } = render(
      <ZahlungAbschlussHarness
        {...props}
        positionenToPay={[{ ...position, menge: 2 }]}
        totalCents={700}
      />,
    )

    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await user.click(screen.getByRole('button', { name: 'Kassieren' }))
    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(1)
    })

    rerender(
      <ZahlungAbschlussHarness
        {...props}
        positionenToPay={[]}
        totalCents={0}
      />,
    )
    await user.click(screen.getByRole('button', { name: 'Pfand auszahlen' }))
    await waitFor(() => {
      expect(pfandZurueckgeben).toHaveBeenCalledTimes(1)
    })

    const [[zahlung]] = zahlungKassieren.mock.calls
    const [[rueckgabe]] = pfandZurueckgeben.mock.calls as [
      [{ rueckgabeId: string }],
    ]
    expect(zahlung.pfand?.rueckgabeId).toMatch(UUID)
    expect(rueckgabe.rueckgabeId).toBe(zahlung.pfand?.rueckgabeId)
  })

  it('keeps the rueckgabeId across Pfand quantities and payment selections and renews it with an empty Pfand selection', async () => {
    const user = userEvent.setup()
    // Jede Antwort geht verloren: Auswahl und rueckgabeId bleiben stehen.
    const zahlungKassieren = vi
      .fn<(zahlung: ZahlungKassieren) => Promise<void>>()
      .mockRejectedValue(new Error('Netzwerkfehler'))
    const props = {
      variant: 'spalte' as const,
      backend: { zahlungKassieren },
      pfandBackend: { pfandZurueckgeben: vi.fn() },
      pfandProdukte: [pfandProdukt],
      tisch,
      restNachZahlungCents: 200,
      zahlungKassiert: vi.fn(),
      pfandAusgezahlt: vi.fn(),
    }
    const { rerender } = render(
      <ZahlungAbschlussHarness
        {...props}
        positionenToPay={[{ ...position, menge: 2 }]}
        totalCents={700}
      />,
    )
    const kassieren = async (aufrufe: number) => {
      await user.click(screen.getByRole('button', { name: 'Kassieren' }))
      await waitFor(() => {
        expect(zahlungKassieren).toHaveBeenCalledTimes(aufrufe)
      })
    }
    const hinzufuegen = () =>
      user.click(screen.getByRole('button', { name: 'Variante hinzufügen' }))
    const entfernen = () =>
      user.click(screen.getByRole('button', { name: 'Variante entfernen' }))

    await hinzufuegen()
    await kassieren(1)
    await hinzufuegen()
    await kassieren(2)
    rerender(
      <ZahlungAbschlussHarness
        {...props}
        positionenToPay={[{ ...position, menge: 1 }]}
        totalCents={350}
      />,
    )
    await kassieren(3)
    await entfernen()
    await entfernen()
    await hinzufuegen()
    await kassieren(4)

    const [erste, mehrPfand, andereZahlung, neuesPfand] =
      zahlungKassieren.mock.calls.map(([zahlung]) => zahlung.pfand)
    expect(erste?.positionen).toEqual([
      { produktId: 9, varianteId: 19, menge: 1 },
    ])
    expect(mehrPfand?.positionen).toEqual([
      { produktId: 9, varianteId: 19, menge: 2 },
    ])
    expect(erste?.rueckgabeId).toMatch(UUID)
    expect(mehrPfand?.rueckgabeId).toBe(erste?.rueckgabeId)
    expect(andereZahlung?.rueckgabeId).toBe(erste?.rueckgabeId)
    expect(neuesPfand?.rueckgabeId).toMatch(UUID)
    expect(neuesPfand?.rueckgabeId).not.toBe(erste?.rueckgabeId)
  })

  // Die abgewiesene Rückgabe ist anders gebucht: Erst die geleerte Auswahl gibt
  // der neu erfassten Rückgabe eine neue rueckgabeId.
  it('clears the Pfand selection after pfandrueckgabe_abweichend', async () => {
    const user = userEvent.setup()
    const pfandZurueckgeben = vi
      .fn()
      .mockRejectedValueOnce(new BackendError(409, 'pfandrueckgabe_abweichend'))
      .mockResolvedValue(undefined)
    render(
      <ZahlungAbschlussHarness
        variant="spalte"
        backend={{ zahlungKassieren: vi.fn() }}
        pfandBackend={{ pfandZurueckgeben }}
        pfandProdukte={[pfandProdukt]}
        tisch={tisch}
        positionenToPay={[]}
        totalCents={0}
        restNachZahlungCents={900}
        zahlungKassiert={vi.fn()}
        pfandAusgezahlt={vi.fn()}
      />,
    )
    const auszahlen = () =>
      user.click(screen.getByRole('button', { name: 'Pfand auszahlen' }))

    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await auszahlen()
    await waitFor(() => {
      expect(pfandZurueckgeben).toHaveBeenCalledTimes(1)
    })
    expect(
      screen.queryByRole('button', { name: 'Pfand auszahlen' }),
    ).not.toBeInTheDocument()

    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await auszahlen()
    await waitFor(() => {
      expect(pfandZurueckgeben).toHaveBeenCalledTimes(2)
    })

    const [[abgewiesen], [neu]] = pfandZurueckgeben.mock.calls as [
      [{ rueckgabeId: string }],
      [{ rueckgabeId: string }],
    ]
    expect(neu.rueckgabeId).toMatch(UUID)
    expect(neu.rueckgabeId).not.toBe(abgewiesen.rueckgabeId)
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

  it('kassiert mit genau einem Backend-Call und der erwarteten Nutzlast (ohne Schlüssel)', async () => {
    const user = userEvent.setup()
    const { zahlungKassieren, zahlungKassiert } = renderSpalte()

    await user.click(screen.getByRole('button', { name: 'Kassieren' }))

    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(1)
    })
    expect(zahlungKassieren).toHaveBeenCalledWith({
      tischId: 1,
      positionen: [{ positionId: position.positionId, menge: 2 }],
      kommentar: '',
    })
    expect(zahlungKassiert).toHaveBeenCalledTimes(1)
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
      <ZahlungAbschlussHarness
        pfandBackend={{ pfandZurueckgeben: vi.fn() }}
        pfandProdukte={[]}
        pfandAusgezahlt={vi.fn()}
        variant="spalte"
        backend={{ zahlungKassieren }}
        tisch={tisch}
        positionenToPay={[{ ...position, menge: 2 }]}
        totalCents={700}
        restNachZahlungCents={200}
        zahlungKassiert={vi.fn()}
      />,
    )

    const button = screen.getByRole('button', { name: 'Kassieren' })
    await user.click(button)
    // Der Loading-Guard deaktiviert den Button: kein zweiter Aufruf.
    await user.click(button)
    resolve()

    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(1)
    })
  })
})
