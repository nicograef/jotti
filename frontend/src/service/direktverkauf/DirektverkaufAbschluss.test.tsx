import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { type ComponentProps, useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useVorgang } from '@/hooks/use-vorgang'
import type { Produkt } from '@/lib/produktSchemas'
import { FakeBackend } from '@/test/FakeBackend'
import { testProdukt } from '@/test/fixtures'
import { renderWithBackend } from '@/test/render'

import type { ReceiptPosition } from '../Receipt'
import type { DirektverkaufTaetigen } from './Direktverkauf'
import { DirektverkaufAbschluss } from './DirektverkaufAbschluss'
import { DirektverkaufTab } from './DirektverkaufTab'

// Verkauf und Pfand-Vorgang liegen auf der Seite; der Harness übernimmt ihre Rolle.
function DirektverkaufAbschlussHarness(
  props: Omit<
    ComponentProps<typeof DirektverkaufAbschluss>,
    'pfandVorgang' | 'vorgang'
  >,
) {
  return (
    <DirektverkaufAbschluss
      {...props}
      pfandVorgang={useVorgang<number>()}
      vorgang={useVorgang<number>()}
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
  { id: '1', name: 'Bratwurst Normal', einzelpreisCents: 350, menge: 2 },
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
    },
  ],
}

// Die feste Spalte (variant="spalte") ist container-neutral testbar: kein
// Drawer, kein Dock.
function backend(): FakeBackend {
  return new FakeBackend()
    .respond('service/direktverkauf-taetigen', {})
    .respond('service/pfand-zurueckgeben', {})
}

function renderSpalte(fake = backend()) {
  const verkaufAbgeschlossen = vi.fn()
  renderWithBackend(
    <DirektverkaufAbschlussHarness
      pfandProdukte={[]}
      pfandAusgezahlt={vi.fn()}
      variant="spalte"
      receiptItems={receiptItems}
      positionen={positionen}
      totalCents={700}
      verkaufAbgeschlossen={verkaufAbgeschlossen}
      verkaufVerworfen={vi.fn()}
    />,
    fake,
  )
  return { fake, verkaufAbgeschlossen }
}

describe('DirektverkaufAbschluss (Spalte)', () => {
  it('zeigt im Leerzustand einen Hinweis und deaktiviert den Aktionsbutton', () => {
    renderWithBackend(
      <DirektverkaufAbschlussHarness
        pfandProdukte={[]}
        pfandAusgezahlt={vi.fn()}
        variant="spalte"
        receiptItems={[]}
        positionen={[]}
        totalCents={0}
        verkaufAbgeschlossen={vi.fn()}
        verkaufVerworfen={vi.fn()}
      />,
      backend(),
    )

    expect(screen.getByText(/Produkte auswählen/)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    ).toBeDisabled()
  })

  it('sends the Pfand return with table id 0', async () => {
    const user = userEvent.setup()
    const fake = backend()
    const pfandAusgezahlt = vi.fn()
    renderWithBackend(
      <DirektverkaufAbschlussHarness
        variant="spalte"
        pfandProdukte={[pfandProdukt]}
        receiptItems={[]}
        positionen={[]}
        totalCents={0}
        verkaufAbgeschlossen={vi.fn()}
        verkaufVerworfen={vi.fn()}
        pfandAusgezahlt={pfandAusgezahlt}
      />,
      fake,
    )

    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await user.click(screen.getByRole('button', { name: 'Pfand auszahlen' }))

    await waitFor(() => {
      expect(pfandAusgezahlt).toHaveBeenCalledTimes(1)
    })
    expect(fake.bodies('service/pfand-zurueckgeben')).toEqual([
      {
        // oxlint-disable-next-line typescript/no-unsafe-assignment
        rueckgabeId: expect.stringMatching(UUID),
        tischId: 0,
        positionen: [{ produktId: 9, varianteId: 19, menge: 1 }],
        kommentar: '',
      },
    ])
    expect(fake.bodies('service/direktverkauf-taetigen')).toHaveLength(0)
  })

  it('shows Auszahlen when the Pfand exceeds the sale', async () => {
    const user = userEvent.setup()
    const fake = backend()
    renderWithBackend(
      <DirektverkaufAbschlussHarness
        variant="spalte"
        pfandProdukte={[pfandProdukt]}
        receiptItems={receiptItems}
        positionen={positionen}
        totalCents={700}
        verkaufAbgeschlossen={vi.fn()}
        verkaufVerworfen={vi.fn()}
        pfandAusgezahlt={vi.fn()}
      />,
      fake,
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
      expect(fake.bodies('service/direktverkauf-taetigen')).toEqual([
        {
          // oxlint-disable-next-line typescript/no-unsafe-assignment
          verkaufId: expect.stringMatching(UUID),
          positionen,
          kommentar: '',
          pfand: {
            // oxlint-disable-next-line typescript/no-unsafe-assignment
            rueckgabeId: expect.stringMatching(UUID),
            positionen: [{ produktId: 9, varianteId: 19, menge: 4 }],
          },
        },
      ])
    })
  })

  it('schließt den Verkauf mit genau einem Backend-Call und erwarteter Nutzlast ab', async () => {
    const user = userEvent.setup()
    const { fake, verkaufAbgeschlossen } = renderSpalte()

    await user.click(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    )

    await waitFor(() => {
      expect(verkaufAbgeschlossen).toHaveBeenCalledTimes(1)
    })
    expect(fake.bodies('service/direktverkauf-taetigen')).toEqual([
      {
        // oxlint-disable-next-line typescript/no-unsafe-assignment
        verkaufId: expect.stringMatching(UUID),
        positionen,
        kommentar: '',
      },
    ])
  })

  it('setzt Erhalten beim neuen Vorgang zurück (kein Übertrag)', async () => {
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
            pfandProdukte={[]}
            pfandAusgezahlt={vi.fn()}
            variant="spalte"
            receiptItems={leer ? [] : receiptItems}
            positionen={leer ? [] : positionen}
            totalCents={leer ? 0 : 700}
            verkaufAbgeschlossen={vi.fn()}
            verkaufVerworfen={vi.fn()}
          />
        </>
      )
    }
    renderWithBackend(<Harness />, backend())

    await user.type(screen.getByLabelText('Erhalten'), '10,00')
    expect(screen.getByLabelText('Erhalten')).toHaveValue('10,00')

    // Auswahl leeren (abgebrochener Vorgang) und neu beginnen: der neue Vorgang
    // startet mit leeren Eingaben, damit nichts aus dem alten übertragen wird.
    await user.click(screen.getByRole('button', { name: 'toggle' }))
    await user.click(screen.getByRole('button', { name: 'toggle' }))

    expect(screen.getByLabelText('Erhalten')).toHaveValue('')
  })

  // Eine verlorene Antwort darf nicht doppelt buchen: auch mit geändertem
  // Kommentar wiederholt der Retry denselben Verkauf.
  it('behält den verkaufId über einen Retry, auch mit geändertem Kommentar', async () => {
    const user = userEvent.setup()
    const { fake } = renderSpalte(
      new FakeBackend().fail(
        'service/direktverkauf-taetigen',
        new Error('kaputt'),
      ),
    )
    const verkaeufe = () =>
      fake.bodies('service/direktverkauf-taetigen') as DirektverkaufTaetigen[]

    const abschliessen = () =>
      screen.getByRole('button', { name: 'Verkauf abschließen' })

    await user.click(abschliessen())
    await waitFor(() => {
      expect(verkaeufe()).toHaveLength(1)
    })
    await user.click(abschliessen())
    await waitFor(() => {
      expect(verkaeufe()).toHaveLength(2)
    })
    await user.type(screen.getByPlaceholderText(/Kommentar/), 'Ohne Senf')
    await user.click(abschliessen())
    await waitFor(() => {
      expect(verkaeufe()).toHaveLength(3)
    })

    const [erster, retry, mitKommentar] = verkaeufe()
    expect(erster.verkaufId).toMatch(UUID)
    expect(retry.verkaufId).toBe(erster.verkaufId)
    expect(mitKommentar.verkaufId).toBe(erster.verkaufId)
    expect(mitKommentar.kommentar).toBe('Ohne Senf')
  })
})

// Dieselbe Kategorie wie testProdukt: die Produktliste zeigt eine Kategorie zur Zeit.
const pommes: Produkt = {
  ...testProdukt,
  id: 2,
  name: 'Pommes',
  varianten: [{ id: 2, name: 'Groß', preisCents: 400, status: 'active' }],
}

// Die Positionskommentare laufen vom Abschluss über den Vorgang zurück in die
// Positionen, die der Tab baut; der Harness nimmt deshalb den ganzen Tab.
function TabHarness() {
  return (
    <DirektverkaufTab
      products={[testProdukt, pommes]}
      pfandProdukte={[]}
      productsLoading={false}
      vorgang={useVorgang<number>()}
      pfandVorgang={useVorgang<number>()}
      onErfolg={vi.fn()}
    />
  )
}

describe('DirektverkaufAbschluss (Positionskommentar)', () => {
  it('sendet den Positionskommentar nur an seiner Position', async () => {
    const user = userEvent.setup()
    const fake = backend()
    renderWithBackend(<TabHarness />, fake)

    const [bratwurst, pommesHinzufuegen] = screen.getAllByRole('button', {
      name: 'Variante hinzufügen',
    })
    await user.click(bratwurst)
    await user.click(bratwurst)
    await user.click(pommesHinzufuegen)
    await user.click(screen.getAllByRole('button', { name: 'Kommentar' })[0])
    await user.type(
      screen.getByPlaceholderText('Kommentar zur Position'),
      ' 1 davon ohne Senf ',
    )
    await user.click(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    )

    await waitFor(() => {
      expect(fake.bodies('service/direktverkauf-taetigen')).toHaveLength(1)
    })
    const [body] = fake.bodies(
      'service/direktverkauf-taetigen',
    ) as DirektverkaufTaetigen[]
    expect(body.positionen).toStrictEqual([
      { produktId: 1, varianteId: 1, menge: 2, kommentar: '1 davon ohne Senf' },
      { produktId: 2, varianteId: 2, menge: 1 },
    ])
  })
})
