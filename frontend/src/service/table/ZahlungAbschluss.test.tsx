import { cleanup, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ComponentProps } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useVorgang } from '@/hooks/use-vorgang'
import { BackendError } from '@/lib/Backend'
import type { Produkt } from '@/lib/produktSchemas'
import { FakeBackend } from '@/test/FakeBackend'
import { testPosition } from '@/test/fixtures'
import { renderWithBackend } from '@/test/render'

import type { Tisch } from './Tisch'
import type { ZahlungKassieren } from './Zahlung'
import { ZahlungAbschluss } from './ZahlungAbschluss'

// Kassieren- und Pfand-Vorgang liegen auf der Seite; der Harness übernimmt ihre Rolle.
function ZahlungAbschlussHarness(
  props: Omit<
    ComponentProps<typeof ZahlungAbschluss>,
    'pfandVorgang' | 'vorgang'
  >,
) {
  return (
    <ZahlungAbschluss
      {...props}
      vorgang={useVorgang<string>()}
      pfandVorgang={useVorgang<number>()}
    />
  )
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
    },
  ],
}

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/

const position = testPosition({ menge: 2 })

function backend(): FakeBackend {
  return new FakeBackend()
    .respond('service/zahlung-kassieren', {})
    .respond('service/pfand-zurueckgeben', {})
}

const zahlungen = (fake: FakeBackend) =>
  fake.bodies('service/zahlung-kassieren') as ZahlungKassieren[]
const rueckgaben = (fake: FakeBackend) =>
  fake.bodies('service/pfand-zurueckgeben') as { rueckgabeId: string }[]

// Die feste Spalte (variant="spalte") ist container-neutral testbar: kein
// Drawer, kein Dock.
function renderSpalte(fake = backend()) {
  const zahlungKassiert = vi.fn()
  renderWithBackend(
    <ZahlungAbschlussHarness
      pfandProdukte={[]}
      pfandAusgezahlt={vi.fn()}
      variant="spalte"
      tisch={tisch}
      positionenToPay={[{ ...position, menge: 2 }]}
      totalCents={700}
      restNachZahlungCents={200}
      zahlungKassiert={zahlungKassiert}
    />,
    fake,
  )
  return { fake, zahlungKassiert }
}

describe('ZahlungAbschluss (Spalte)', () => {
  it('zeigt im Leerzustand einen Hinweis und deaktiviert den Aktionsbutton', () => {
    renderWithBackend(
      <ZahlungAbschlussHarness
        pfandProdukte={[]}
        pfandAusgezahlt={vi.fn()}
        variant="spalte"
        tisch={tisch}
        positionenToPay={[]}
        totalCents={0}
        restNachZahlungCents={900}
        zahlungKassiert={vi.fn()}
      />,
      backend(),
    )

    expect(screen.getByText(/Positionen auswählen/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Kassieren' })).toBeDisabled()
  })

  it('sends only the Pfand return with the table id', async () => {
    const user = userEvent.setup()
    const fake = backend()
    const pfandAusgezahlt = vi.fn()
    renderWithBackend(
      <ZahlungAbschlussHarness
        variant="spalte"
        pfandProdukte={[pfandProdukt]}
        tisch={tisch}
        positionenToPay={[]}
        totalCents={0}
        restNachZahlungCents={900}
        zahlungKassiert={vi.fn()}
        pfandAusgezahlt={pfandAusgezahlt}
      />,
      fake,
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
      expect(pfandAusgezahlt).toHaveBeenCalledTimes(1)
    })
    expect(fake.bodies('service/pfand-zurueckgeben')).toEqual([
      {
        // oxlint-disable-next-line typescript/no-unsafe-assignment
        rueckgabeId: expect.stringMatching(UUID),
        tischId: 1,
        positionen: [{ produktId: 9, varianteId: 19, menge: 2 }],
        kommentar: 'Zwei Gläser',
      },
    ])
    expect(zahlungen(fake)).toHaveLength(0)
  })

  it('nets the Pfand against the payment', async () => {
    const user = userEvent.setup()
    const fake = backend()
    renderWithBackend(
      <ZahlungAbschlussHarness
        variant="spalte"
        pfandProdukte={[pfandProdukt]}
        tisch={tisch}
        positionenToPay={[{ ...position, menge: 2 }]}
        totalCents={700}
        restNachZahlungCents={200}
        zahlungKassiert={vi.fn()}
        pfandAusgezahlt={vi.fn()}
      />,
      fake,
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
      expect(zahlungen(fake)).toEqual([
        {
          // oxlint-disable-next-line typescript/no-unsafe-assignment
          zahlungId: expect.stringMatching(UUID),
          tischId: 1,
          positionen: [{ positionId: position.positionId, menge: 2 }],
          kommentar: '',
          pfand: {
            // oxlint-disable-next-line typescript/no-unsafe-assignment
            rueckgabeId: expect.stringMatching(UUID),
            positionen: [{ produktId: 9, varianteId: 19, menge: 1 }],
          },
        },
      ])
    })
  })

  // Zahlung und Rückgabe sind gebucht, die Antwort geht verloren; der Refetch
  // nimmt die bezahlten Positionen aus der Auswahl. Erst die alte rueckgabeId
  // lässt den Server die Wiederholung als verrechnet abweisen.
  it('keeps the rueckgabeId when a refetch prunes the paid positions', async () => {
    const user = userEvent.setup()
    const fake = backend().fail(
      'service/zahlung-kassieren',
      new Error('Netzwerkfehler'),
    )
    const props = {
      variant: 'spalte' as const,
      pfandProdukte: [pfandProdukt],
      tisch,
      restNachZahlungCents: 200,
      zahlungKassiert: vi.fn(),
      pfandAusgezahlt: vi.fn(),
    }
    const { rerender } = renderWithBackend(
      <ZahlungAbschlussHarness
        {...props}
        positionenToPay={[{ ...position, menge: 2 }]}
        totalCents={700}
      />,
      fake,
    )

    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await user.click(screen.getByRole('button', { name: 'Kassieren' }))
    await waitFor(() => {
      expect(zahlungen(fake)).toHaveLength(1)
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
      expect(rueckgaben(fake)).toHaveLength(1)
    })

    const [zahlung] = zahlungen(fake)
    const [rueckgabe] = rueckgaben(fake)
    expect(zahlung.pfand?.rueckgabeId).toMatch(UUID)
    expect(rueckgabe.rueckgabeId).toBe(zahlung.pfand?.rueckgabeId)
  })

  it('keeps the rueckgabeId across Pfand quantities and payment selections and renews it with an empty Pfand selection', async () => {
    const user = userEvent.setup()
    // Jede Antwort geht verloren: Auswahl und rueckgabeId bleiben stehen.
    const fake = backend().fail(
      'service/zahlung-kassieren',
      new Error('Netzwerkfehler'),
    )
    const props = {
      variant: 'spalte' as const,
      pfandProdukte: [pfandProdukt],
      tisch,
      restNachZahlungCents: 200,
      zahlungKassiert: vi.fn(),
      pfandAusgezahlt: vi.fn(),
    }
    const { rerender } = renderWithBackend(
      <ZahlungAbschlussHarness
        {...props}
        positionenToPay={[{ ...position, menge: 2 }]}
        totalCents={700}
      />,
      fake,
    )
    const kassieren = async (aufrufe: number) => {
      await user.click(screen.getByRole('button', { name: 'Kassieren' }))
      await waitFor(() => {
        expect(zahlungen(fake)).toHaveLength(aufrufe)
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

    const [erste, mehrPfand, andereZahlung, neuesPfand] = zahlungen(fake).map(
      (zahlung) => zahlung.pfand,
    )
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
    const fake = backend().respond(
      'service/pfand-zurueckgeben',
      vi
        .fn()
        .mockRejectedValueOnce(
          new BackendError(409, 'pfandrueckgabe_abweichend'),
        )
        .mockReturnValue({}),
    )
    renderWithBackend(
      <ZahlungAbschlussHarness
        variant="spalte"
        pfandProdukte={[pfandProdukt]}
        tisch={tisch}
        positionenToPay={[]}
        totalCents={0}
        restNachZahlungCents={900}
        zahlungKassiert={vi.fn()}
        pfandAusgezahlt={vi.fn()}
      />,
      fake,
    )
    const auszahlen = () =>
      user.click(screen.getByRole('button', { name: 'Pfand auszahlen' }))

    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await auszahlen()
    await waitFor(() => {
      expect(rueckgaben(fake)).toHaveLength(1)
    })
    expect(
      screen.queryByRole('button', { name: 'Pfand auszahlen' }),
    ).not.toBeInTheDocument()

    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await auszahlen()
    await waitFor(() => {
      expect(rueckgaben(fake)).toHaveLength(2)
    })

    const [abgewiesen, neu] = rueckgaben(fake)
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

  it('kassiert mit genau einem Backend-Call und der erwarteten Nutzlast', async () => {
    const user = userEvent.setup()
    const { fake, zahlungKassiert } = renderSpalte()

    await user.click(screen.getByRole('button', { name: 'Kassieren' }))

    await waitFor(() => {
      expect(zahlungKassiert).toHaveBeenCalledTimes(1)
    })
    expect(zahlungen(fake)).toEqual([
      {
        // oxlint-disable-next-line typescript/no-unsafe-assignment
        zahlungId: expect.stringMatching(UUID),
        tischId: 1,
        positionen: [{ positionId: position.positionId, menge: 2 }],
        kommentar: '',
      },
    ])
  })

  // Eine verlorene Antwort darf eine Teilzahlung nicht doppelt buchen: die
  // Wiederholung trägt dieselbe zahlungId, auch mit geändertem Kommentar.
  it('keeps the zahlungId for a retry of the same payment', async () => {
    const user = userEvent.setup()
    const { fake } = renderSpalte(
      backend().fail('service/zahlung-kassieren', new Error('Netzwerkfehler')),
    )
    const kassieren = async (aufrufe: number) => {
      await user.click(screen.getByRole('button', { name: 'Kassieren' }))
      await waitFor(() => {
        expect(zahlungen(fake)).toHaveLength(aufrufe)
      })
    }

    await kassieren(1)
    await user.type(screen.getByPlaceholderText(/Kommentar/), 'Rest später')
    await kassieren(2)

    const [erster, wiederholung] = zahlungen(fake)
    expect(erster.zahlungId).toMatch(UUID)
    expect(wiederholung.kommentar).toBe('Rest später')
    expect(wiederholung.zahlungId).toBe(erster.zahlungId)
  })

  it('löst bei Doppelklick keinen zweiten Aufruf aus', async () => {
    const user = userEvent.setup()
    let resolve: (antwort: object) => void = () => undefined
    const { fake, zahlungKassiert } = renderSpalte(
      backend().respond(
        'service/zahlung-kassieren',
        () =>
          new Promise<object>((r) => {
            resolve = r
          }),
      ),
    )

    const button = screen.getByRole('button', { name: 'Kassieren' })
    await user.click(button)
    // Der Loading-Guard deaktiviert den Button: kein zweiter Aufruf.
    await user.click(button)
    resolve({})

    await waitFor(() => {
      expect(zahlungKassiert).toHaveBeenCalledTimes(1)
    })
    expect(zahlungen(fake)).toHaveLength(1)
  })
})
