import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ComponentProps } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useVorgang } from '@/hooks/use-vorgang'
import type { Produkt } from '@/lib/produktSchemas'
import { FakeBackend } from '@/test/FakeBackend'
import { testProdukt } from '@/test/fixtures'
import { renderWithBackend } from '@/test/render'

import type { ReceiptPosition } from '../Receipt'
import type { BestellungAufnehmen } from './Bestellung'
import { BestellungAbschluss } from './BestellungAbschluss'
import { BestellungTab } from './BestellungTab'
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
  { id: '1', name: 'Bratwurst Normal', einzelpreisCents: 350, menge: 2 },
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

  // Eine verlorene Antwort darf nicht doppelt bestellen: auch mit geändertem
  // Kommentar wiederholt der Retry dieselbe Bestellung.
  it('behält den bestellungId, wenn sich der Kommentar nach einem Fehlversuch ändert', async () => {
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

    await user.type(screen.getByPlaceholderText(/Kommentar/), 'Ohne Zwiebeln')
    await user.click(aufnehmen())
    await waitFor(() => {
      expect(bestellungen()).toHaveLength(2)
    })
    expect(bestellungen()[1].bestellungId).toBe(ersterKey)
    expect(bestellungen()[1].kommentar).toBe('Ohne Zwiebeln')
  })
})

// Dieselbe Kategorie wie testProdukt: die Produktliste zeigt eine Kategorie zur Zeit.
const pommes: Produkt = {
  ...testProdukt,
  id: 2,
  name: 'Pommes',
  varianten: [
    {
      id: 2,
      name: 'Groß',
      preisCents: 400,
      status: 'active',
      pfandVarianteId: null,
    },
  ],
}

// Die Positionskommentare laufen vom Abschluss über den Vorgang zurück in die
// Positionen, die der Tab baut; der Harness nimmt deshalb den ganzen Tab.
function TabHarness() {
  return (
    <BestellungTab
      tisch={tisch}
      products={[testProdukt, pommes]}
      productsLoading={false}
      vorgang={useVorgang<number>()}
      onErfolg={vi.fn()}
    />
  )
}

async function waehleBratwurstUndPommes(
  user: ReturnType<typeof userEvent.setup>,
) {
  const [bratwurst, pommesHinzufuegen] = screen.getAllByRole('button', {
    name: 'Variante hinzufügen',
  })
  await user.click(bratwurst)
  await user.click(pommesHinzufuegen)
}

describe('BestellungAbschluss (Positionskommentar)', () => {
  it('sendet den Positionskommentar nur an seiner Position', async () => {
    const user = userEvent.setup()
    const fake = new FakeBackend().respond('service/bestellung-aufnehmen', {})
    renderWithBackend(<TabHarness />, fake)

    await waehleBratwurstUndPommes(user)
    await user.click(screen.getAllByRole('button', { name: 'Kommentar' })[0])
    await user.type(
      screen.getByPlaceholderText('Kommentar zur Position'),
      ' ohne Senf ',
    )
    await user.click(
      screen.getByRole('button', { name: 'Bestellung aufnehmen' }),
    )

    await waitFor(() => {
      expect(fake.bodies('service/bestellung-aufnehmen')).toHaveLength(1)
    })
    const [body] = fake.bodies(
      'service/bestellung-aufnehmen',
    ) as BestellungAufnehmen[]
    expect(body.positionen).toStrictEqual([
      { produktId: 1, varianteId: 1, menge: 1, kommentar: 'ohne Senf' },
      { produktId: 2, varianteId: 2, menge: 1 },
    ])
  })

  // Ein geänderter Positionskommentar ändert die Bestellung: der Retry darf sie
  // nicht als Duplikat der verlorenen verlieren.
  it('erneuert den bestellungId, wenn sich ein Positionskommentar nach einem Fehlversuch ändert', async () => {
    const user = userEvent.setup()
    const fake = new FakeBackend().respond(
      'service/bestellung-aufnehmen',
      vi.fn().mockRejectedValueOnce(new Error('kaputt')).mockReturnValue({}),
    )
    renderWithBackend(<TabHarness />, fake)
    const bestellungen = () =>
      fake.bodies('service/bestellung-aufnehmen') as BestellungAufnehmen[]
    const aufnehmen = () =>
      screen.getByRole('button', { name: 'Bestellung aufnehmen' })

    await waehleBratwurstUndPommes(user)
    await user.click(aufnehmen())
    await waitFor(() => {
      expect(bestellungen()).toHaveLength(1)
    })

    await user.click(screen.getAllByRole('button', { name: 'Kommentar' })[1])
    await user.type(
      screen.getByPlaceholderText('Kommentar zur Position'),
      'ohne Ketchup',
    )
    await user.click(aufnehmen())
    await waitFor(() => {
      expect(bestellungen()).toHaveLength(2)
    })

    expect(bestellungen()[1].bestellungId).not.toBe(
      bestellungen()[0].bestellungId,
    )
    expect(bestellungen()[1].positionen[1].kommentar).toBe('ohne Ketchup')
  })
})
